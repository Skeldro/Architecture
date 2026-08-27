package main

import (
	"context"
	"database/sql"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"sync"
	"testing"
)

// The suite runs from a clean checkout against a database supplied by
// scripts/localpg.sh (see `make test`). It covers FR1 through FR4.

func dsn(t *testing.T) string {
	t.Helper()
	for _, k := range []string{"TEST_DATABASE_URL", "DATABASE_URL"} {
		if v := os.Getenv(k); v != "" {
			return v
		}
	}
	t.Skip("no TEST_DATABASE_URL or DATABASE_URL set; run `make test`")
	return ""
}

func newTestApp(t *testing.T) *app {
	t.Helper()
	db, err := openDB(dsn(t), 10)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	if err := ensureSchema(context.Background(), db); err != nil {
		t.Fatalf("schema: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return &app{
		db:  db,
		log: slog.New(slog.NewJSONHandler(io.Discard, nil)),
		m:   newMetrics(),
	}
}

func clean(t *testing.T, a *app) {
	t.Helper()
	if _, err := a.db.Exec(`TRUNCATE documents RESTART IDENTITY`); err != nil {
		t.Fatalf("truncate: %v", err)
	}
}

func post(h http.Handler, path string, form url.Values) *httptest.ResponseRecorder {
	r := httptest.NewRequest(http.MethodPost, path, strings.NewReader(form.Encode()))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	return w
}

func get(h http.Handler, path string) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest(http.MethodGet, path, nil))
	return w
}

func createDoc(t *testing.T, h http.Handler, title string) string {
	t.Helper()
	w := post(h, "/create", url.Values{"title": {title}})
	if w.Code != http.StatusSeeOther {
		t.Fatalf("create %q: got %d, want 303 (%s)", title, w.Code, w.Body.String())
	}
	loc := w.Header().Get("Location")
	id := strings.TrimPrefix(loc, "/doc?id=")
	if id == loc || id == "" {
		t.Fatalf("create: unexpected redirect %q", loc)
	}
	return id
}

// --- The three functional behaviours carried forward from Phase 0 ---

func TestCreateEditAndList(t *testing.T) {
	a := newTestApp(t)
	clean(t, a)
	h := a.routes()

	id := createDoc(t, h, "sprint notes")

	body := get(h, "/doc?id="+id).Body.String()
	if !strings.Contains(body, "sprint notes") {
		t.Fatalf("editor did not render the title: %s", body)
	}

	if w := post(h, "/save", url.Values{
		"id": {id}, "version": {"1"}, "content": {"line one\r\nline two"},
	}); w.Code != http.StatusSeeOther {
		t.Fatalf("save: got %d, want 303", w.Code)
	}

	var content string
	if err := a.db.QueryRow(`SELECT content FROM documents WHERE id=$1`, id).Scan(&content); err != nil {
		t.Fatal(err)
	}
	// Decision 2 says documents hold plain \n; browsers submit CRLF.
	if content != "line one\nline two" {
		t.Fatalf("line endings not normalised: %q", content)
	}

	if list := get(h, "/").Body.String(); !strings.Contains(list, "sprint notes") {
		t.Fatalf("document missing from list: %s", list)
	}
}

// --- FR1: configuration comes from the environment; schema init is repeatable ---

func TestFR1_SchemaInitIsIdempotentAcrossInstances(t *testing.T) {
	a := newTestApp(t)
	// Every instance runs ensureSchema on boot and they start simultaneously,
	// so it has to be safe to run repeatedly and concurrently.
	var wg sync.WaitGroup
	errs := make(chan error, 4)
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := ensureSchema(context.Background(), a.db); err != nil {
				errs <- err
			}
		}()
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		t.Fatalf("concurrent schema init failed: %v", err)
	}
}

func TestFR1_RequiresOnlyEnvironmentConfiguration(t *testing.T) {
	// Nothing may be read from the local filesystem: a clean machine has no
	// docs directory, no config file, only environment variables.
	if _, err := os.Stat("docs"); err == nil {
		t.Log("a stale Phase 0 docs/ directory exists on disk; the app must not use it")
	}
	a := newTestApp(t)
	clean(t, a)
	h := a.routes()
	id := createDoc(t, h, "env only")
	if w := get(h, "/doc?id="+id); w.Code != http.StatusOK {
		t.Fatalf("serving with only DATABASE_URL set failed: %d", w.Code)
	}
}

// --- FR2: state survives recreation of the runtime ---

func TestFR2_StateSurvivesRuntimeRecreation(t *testing.T) {
	first := newTestApp(t)
	clean(t, first)
	id := createDoc(t, first.routes(), "outlives the process")
	if w := post(first.routes(), "/save", url.Values{
		"id": {id}, "version": {"1"}, "content": {"written by the first instance"},
	}); w.Code != http.StatusSeeOther {
		t.Fatalf("save: %d", w.Code)
	}

	// Destroy the runtime completely: pool closed, instance discarded.
	first.db.Close()

	second := newTestApp(t)
	body := get(second.routes(), "/doc?id="+id).Body.String()
	if !strings.Contains(body, "written by the first instance") {
		t.Fatalf("content did not survive runtime recreation: %s", body)
	}
	if list := get(second.routes(), "/").Body.String(); !strings.Contains(list, "outlives the process") {
		t.Fatalf("document missing after runtime recreation: %s", list)
	}
}

// --- FR3: concurrent writes do not silently lose content ---

func TestFR3_ConcurrentSavesDoNotSilentlyLoseContent(t *testing.T) {
	a := newTestApp(t)
	clean(t, a)
	h := a.routes()
	id := createDoc(t, h, "contended document")

	const winnerMarker = "AAA-from-writer-one"
	const loserMarker = "BBB-from-writer-two"

	// Both writers read version 1 and submit against it, which is exactly the
	// lost-update race Phase 0 resolved by silently overwriting.
	var wg sync.WaitGroup
	results := make([]*httptest.ResponseRecorder, 2)
	bodies := []string{winnerMarker, loserMarker}
	start := make(chan struct{})
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			<-start
			results[i] = post(h, "/save", url.Values{
				"id": {id}, "version": {"1"}, "content": {bodies[i]},
			})
		}(i)
	}
	close(start)
	wg.Wait()

	var accepted, rejected int
	var rejectedBody, acceptedMarker string
	for i, w := range results {
		switch w.Code {
		case http.StatusSeeOther:
			accepted++
			acceptedMarker = bodies[i]
		case http.StatusConflict:
			rejected++
			rejectedBody = w.Body.String()
		default:
			t.Fatalf("unexpected status %d: %s", w.Code, w.Body.String())
		}
	}
	if accepted != 1 || rejected != 1 {
		t.Fatalf("want exactly one accepted and one rejected, got %d/%d", accepted, rejected)
	}

	var stored string
	var version int
	if err := a.db.QueryRow(`SELECT content, version FROM documents WHERE id=$1`, id).
		Scan(&stored, &version); err != nil {
		t.Fatal(err)
	}
	if stored != acceptedMarker {
		t.Fatalf("stored content %q is not the accepted write %q", stored, acceptedMarker)
	}
	if version != 2 {
		t.Fatalf("version should have advanced exactly once, got %d", version)
	}

	// The point of FR3: the rejected writer's text must come back to them.
	// Detecting the conflict is not enough if the content is dropped.
	loser := loserMarker
	if acceptedMarker == loserMarker {
		loser = winnerMarker
	}
	if !strings.Contains(rejectedBody, loser) {
		t.Fatalf("rejected writer's content was not returned to them:\n%s", rejectedBody)
	}
	if !strings.Contains(rejectedBody, acceptedMarker) {
		t.Fatalf("conflict page did not show the winning version:\n%s", rejectedBody)
	}
	if a.m.conflicts.Load() != 1 {
		t.Fatalf("conflict counter = %d, want 1", a.m.conflicts.Load())
	}
}

func TestFR3_StaleVersionIsRejectedDeterministically(t *testing.T) {
	a := newTestApp(t)
	clean(t, a)
	h := a.routes()
	id := createDoc(t, h, "stale write")

	if w := post(h, "/save", url.Values{"id": {id}, "version": {"1"}, "content": {"first"}}); w.Code != http.StatusSeeOther {
		t.Fatalf("first save: %d", w.Code)
	}
	// Same base version again: this writer's editor is now out of date.
	w := post(h, "/save", url.Values{"id": {id}, "version": {"1"}, "content": {"second"}})
	if w.Code != http.StatusConflict {
		t.Fatalf("stale save: got %d, want 409", w.Code)
	}
	var stored string
	if err := a.db.QueryRow(`SELECT content FROM documents WHERE id=$1`, id).Scan(&stored); err != nil {
		t.Fatal(err)
	}
	if stored != "first" {
		t.Fatalf("stale write overwrote the newer one: %q", stored)
	}
}

// --- FR4: health is determinable without reading logs ---

func TestFR4_HealthReportsHealthyWhenStorageReachable(t *testing.T) {
	a := newTestApp(t)
	w := get(a.routes(), "/healthz")
	if w.Code != http.StatusOK {
		t.Fatalf("healthz: got %d, want 200", w.Code)
	}
	if !strings.Contains(w.Body.String(), `"status":"healthy"`) {
		t.Fatalf("healthz body: %s", w.Body.String())
	}
}

func TestFR4_HealthDetectsUnreachableStorage(t *testing.T) {
	// A liveness-only check would answer "healthy" here, which is the precise
	// failure FR4 exists to catch.
	db, err := openDB("postgres://postgres@127.0.0.1:1/nope?sslmode=disable&connect_timeout=1", 2)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	dead := &app{db: db, log: slog.New(slog.NewJSONHandler(io.Discard, nil)), m: newMetrics()}

	w := get(dead.routes(), "/healthz")
	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("healthz with dead storage: got %d, want 503", w.Code)
	}
	if !strings.Contains(w.Body.String(), `"status":"unhealthy"`) {
		t.Fatalf("healthz body: %s", w.Body.String())
	}

	// Liveness must still pass: the process is fine, its dependency is not.
	if w := get(dead.routes(), "/livez"); w.Code != http.StatusOK {
		t.Fatalf("livez with dead storage: got %d, want 200", w.Code)
	}
}

func TestFR4_MetricsExposeTheDecidedSignals(t *testing.T) {
	a := newTestApp(t)
	clean(t, a)
	h := a.routes()
	id := createDoc(t, h, "metrics subject")
	get(h, "/doc?id="+id)
	post(h, "/save", url.Values{"id": {id}, "version": {"99"}, "content": {"stale"}}) // force a conflict

	body := get(h, "/metrics").Body.String()
	for _, want := range []string{
		"collabdocs_http_requests_total",
		"collabdocs_http_error_ratio",
		"collabdocs_request_duration_p95_ms",
		"collabdocs_save_conflicts_total 1",
		`collabdocs_db_pool{state="max_open"} 10`,
		"collabdocs_db_pool_wait_total",
	} {
		if !strings.Contains(body, want) {
			t.Errorf("/metrics is missing %q", want)
		}
	}
}

func TestMetricsHistogramQuantiles(t *testing.T) {
	h := newHistogram()
	for i := 0; i < 95; i++ {
		h.add(3) // lands in the 5ms bucket
	}
	for i := 0; i < 5; i++ {
		h.add(800) // lands in the 1000ms bucket
	}
	if p50 := h.quantile(0.50); p50 > 5 {
		t.Errorf("p50 = %.2f, want <= 5", p50)
	}
	if p95 := h.quantile(0.95); p95 > 5 {
		t.Errorf("p95 = %.2f, want <= 5 (95%% of samples are 3ms)", p95)
	}
	if p99 := h.quantile(0.99); p99 <= 500 {
		t.Errorf("p99 = %.2f, want > 500 (the slow tail must show)", p99)
	}
}

var _ = sql.ErrNoRows
