// CollabDocs — Phase 1.
//
// Phase 1 constraints deliberately in force (see phase1-decisions.md):
//   - No abstraction around storage. SQL is written inline in the handlers,
//     and the repetition that causes is intentional; Phase 3 earns the
//     abstraction by having a second reason to exist.
//   - No authentication, no real-time mechanism, no business-rule extraction.
//
// The `app` struct exists only so tests can construct independent instances
// against one database (acceptance criterion 6 requires a test suite, and
// FR2's test needs two lifetimes). It holds the pool; it does not wrap it.
package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"html/template"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

const schema = `
CREATE TABLE IF NOT EXISTS documents (
    id         BIGSERIAL   PRIMARY KEY,
    title      TEXT        NOT NULL UNIQUE,
    content    TEXT        NOT NULL DEFAULT '',
    version    INTEGER     NOT NULL DEFAULT 1,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);`

type app struct {
	db  *sql.DB
	log *slog.Logger
	m   *metrics
}

// ---------- templates ----------

var listTmpl = template.Must(template.New("list").Parse(`<!doctype html>
<title>Collab Docs</title>
<h1>Collab Docs</h1>
<form method="POST" action="/create">
  <input name="title" placeholder="New document title" autofocus>
  <button>Create</button>
</form>
{{if .}}<ul>
{{range .}}  <li><a href="/doc?id={{.ID}}">{{.Title}}</a></li>
{{end}}</ul>{{else}}<p>No documents yet.</p>{{end}}`))

var editorTmpl = template.Must(template.New("editor").Parse(`<!doctype html>
<title>{{.Title}} — Collab Docs</title>
<p><a href="/">&larr; all documents</a></p>
<h1>{{.Title}}</h1>
<form method="POST" action="/save">
  <input type="hidden" name="id" value="{{.ID}}">
  <input type="hidden" name="version" value="{{.Version}}">
  <textarea name="content" rows="25" cols="90">{{.Content}}</textarea>
  <br><button>Save</button>
</form>`))

// The conflict page is how FR3 is satisfied: the losing writer's text is never
// discarded, it is handed back alongside the winning version so a human decides.
var conflictTmpl = template.Must(template.New("conflict").Parse(`<!doctype html>
<title>Conflict — {{.Title}}</title>
<p><a href="/">&larr; all documents</a></p>
<h1>{{.Title}} — save conflict</h1>
<p><strong>Someone else saved this document while you were editing.</strong>
Nothing has been lost. Their version is now saved; your text is below.</p>
<h2>Currently saved (version {{.Version}})</h2>
<pre>{{.Content}}</pre>
<h2>Your unsaved text</h2>
<form method="POST" action="/save">
  <input type="hidden" name="id" value="{{.ID}}">
  <input type="hidden" name="version" value="{{.Version}}">
  <textarea name="content" rows="20" cols="90">{{.Yours}}</textarea>
  <br><button>Save mine over theirs</button>
</form>`))

type docView struct {
	ID      int64
	Title   string
	Content string
	Version int
	Yours   string
}

type listItem struct {
	ID    int64
	Title string
}

// ---------- handlers ----------

func (a *app) handleList(w http.ResponseWriter, r *http.Request) {
	// ServeMux routes "/" as a catch-all, so anything unmatched arrives here.
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	rows, err := a.db.QueryContext(r.Context(), `SELECT id, title FROM documents ORDER BY title`)
	if err != nil {
		a.fail(w, r, http.StatusInternalServerError, "list query", err)
		return
	}
	defer rows.Close()
	var items []listItem
	for rows.Next() {
		var it listItem
		if err := rows.Scan(&it.ID, &it.Title); err != nil {
			a.fail(w, r, http.StatusInternalServerError, "list scan", err)
			return
		}
		items = append(items, it)
	}
	if err := rows.Err(); err != nil {
		a.fail(w, r, http.StatusInternalServerError, "list rows", err)
		return
	}
	listTmpl.Execute(w, items)
}

func (a *app) handleCreate(w http.ResponseWriter, r *http.Request) {
	title := strings.TrimSpace(r.FormValue("title"))
	if title == "" {
		http.Error(w, "title is required", http.StatusBadRequest)
		return
	}
	var id int64
	err := a.db.QueryRowContext(r.Context(),
		`INSERT INTO documents (title) VALUES ($1) RETURNING id`, title).Scan(&id)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			http.Error(w, "a document with that title already exists", http.StatusConflict)
			return
		}
		a.fail(w, r, http.StatusInternalServerError, "create", err)
		return
	}
	http.Redirect(w, r, "/doc?id="+strconv.FormatInt(id, 10), http.StatusSeeOther)
}

func (a *app) handleDoc(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.FormValue("id"), 10, 64)
	if err != nil {
		http.Error(w, "bad id", http.StatusBadRequest)
		return
	}
	var d docView
	d.ID = id
	err = a.db.QueryRowContext(r.Context(),
		`SELECT title, content, version FROM documents WHERE id = $1`, id).
		Scan(&d.Title, &d.Content, &d.Version)
	if errors.Is(err, sql.ErrNoRows) {
		http.Error(w, "no such document", http.StatusNotFound)
		return
	}
	if err != nil {
		a.fail(w, r, http.StatusInternalServerError, "doc query", err)
		return
	}
	editorTmpl.Execute(w, d)
}

func (a *app) handleSave(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.FormValue("id"), 10, 64)
	if err != nil {
		http.Error(w, "bad id", http.StatusBadRequest)
		return
	}
	version, err := strconv.Atoi(r.FormValue("version"))
	if err != nil {
		http.Error(w, "bad version", http.StatusBadRequest)
		return
	}
	// Browsers submit textarea content with CRLF; documents hold plain \n.
	content := strings.ReplaceAll(r.FormValue("content"), "\r\n", "\n")

	// The whole of FR3 is this one statement: the version predicate makes the
	// read-modify-write atomic inside the database, so it holds across every
	// application instance. No in-process lock could do this.
	res, err := a.db.ExecContext(r.Context(),
		`UPDATE documents SET content = $1, version = version + 1, updated_at = now()
		 WHERE id = $2 AND version = $3`, content, id, version)
	if err != nil {
		a.fail(w, r, http.StatusInternalServerError, "save", err)
		return
	}
	n, err := res.RowsAffected()
	if err != nil {
		a.fail(w, r, http.StatusInternalServerError, "save rows", err)
		return
	}
	if n == 1 {
		http.Redirect(w, r, "/doc?id="+strconv.FormatInt(id, 10), http.StatusSeeOther)
		return
	}

	// Zero rows: either the document is gone, or somebody else wrote first.
	var d docView
	d.ID, d.Yours = id, content
	err = a.db.QueryRowContext(r.Context(),
		`SELECT title, content, version FROM documents WHERE id = $1`, id).
		Scan(&d.Title, &d.Content, &d.Version)
	if errors.Is(err, sql.ErrNoRows) {
		http.Error(w, "no such document", http.StatusNotFound)
		return
	}
	if err != nil {
		a.fail(w, r, http.StatusInternalServerError, "conflict lookup", err)
		return
	}
	a.m.conflicts.Add(1)
	a.log.Warn("save conflict", "doc_id", id, "submitted_version", version, "current_version", d.Version)
	w.WriteHeader(http.StatusConflict)
	conflictTmpl.Execute(w, d)
}

// handleReadyz is the FR4 signal: it proves storage is reachable, not merely
// that the process is alive. A liveness-only check would report healthy while
// every save failed.
func (a *app) handleReadyz(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()
	var ok int
	if err := a.db.QueryRowContext(ctx, `SELECT 1`).Scan(&ok); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusServiceUnavailable)
		fmt.Fprintf(w, `{"status":"unhealthy","storage":"unreachable","error":%q}`+"\n", err.Error())
		return
	}
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprint(w, `{"status":"healthy","storage":"reachable"}`+"\n")
}

func (a *app) handleLivez(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprint(w, `{"status":"alive"}`+"\n")
}

func (a *app) fail(w http.ResponseWriter, r *http.Request, code int, where string, err error) {
	a.log.Error("request failed", "where", where, "error", err.Error(), "path", r.URL.Path)
	http.Error(w, http.StatusText(code), code)
}

// ---------- wiring ----------

func (a *app) routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/", a.handleList)
	mux.HandleFunc("/create", a.handleCreate)
	mux.HandleFunc("/doc", a.handleDoc)
	mux.HandleFunc("/save", a.handleSave)
	mux.HandleFunc("/healthz", a.handleReadyz)
	mux.HandleFunc("/livez", a.handleLivez)
	mux.HandleFunc("/metrics", a.m.handler(a.db))
	return a.observe(mux)
}

type recorder struct {
	http.ResponseWriter
	status int
}

func (r *recorder) WriteHeader(c int) { r.status = c; r.ResponseWriter.WriteHeader(c) }

func (a *app) observe(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/metrics" {
			next.ServeHTTP(w, r)
			return
		}
		start := time.Now()
		rec := &recorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rec, r)
		ms := float64(time.Since(start).Microseconds()) / 1000.0

		route := r.URL.Path
		a.m.observe(r.Method, route, rec.status, ms)
		a.log.Info("request",
			"method", r.Method, "path", route, "status", rec.status,
			"duration_ms", ms, "doc_id", r.FormValue("id"))
	})
}

func openDB(dsn string, maxConns int) (*sql.DB, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}
	// Decision 2. database/sql defaults MaxOpenConns to UNLIMITED, which is how
	// "stateless instances plus one database" designs exhaust the server. The
	// per-instance ceiling multiplied by the maximum instance count is what
	// must stay under the server's max_connections.
	db.SetMaxOpenConns(maxConns)
	db.SetMaxIdleConns(maxConns) // default of 2 causes constant reconnect churn
	db.SetConnMaxLifetime(5 * time.Minute)
	return db, nil
}

// ensureSchema is guarded by an advisory lock because every instance runs it on
// boot and they start simultaneously.
func ensureSchema(ctx context.Context, db *sql.DB) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err := tx.ExecContext(ctx, `SELECT pg_advisory_xact_lock(4711)`); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, schema); err != nil {
		return err
	}
	return tx.Commit()
}

func env(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func main() {
	log := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Error("DATABASE_URL is required")
		os.Exit(1)
	}
	maxConns, err := strconv.Atoi(env("DB_MAX_CONNS", "10"))
	if err != nil {
		log.Error("DB_MAX_CONNS must be an integer", "error", err.Error())
		os.Exit(1)
	}

	db, err := openDB(dsn, maxConns)
	if err != nil {
		log.Error("cannot open database", "error", err.Error())
		os.Exit(1)
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := ensureSchema(ctx, db); err != nil {
		log.Error("schema initialisation failed", "error", err.Error())
		os.Exit(1)
	}

	a := &app{db: db, log: log, m: newMetrics()}
	srv := &http.Server{Addr: ":" + env("PORT", "8080"), Handler: a.routes()}

	// Managed container platforms send SIGTERM and then wait. Without this,
	// every deploy drops in-flight requests, and Decision 1's zero-downtime
	// rolling deploy claim would be false.
	idle := make(chan struct{})
	go func() {
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, syscall.SIGTERM, syscall.SIGINT)
		<-sig
		log.Info("shutdown signal received, draining")
		sctx, scancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer scancel()
		if err := srv.Shutdown(sctx); err != nil {
			log.Error("graceful shutdown failed", "error", err.Error())
		}
		close(idle)
	}()

	log.Info("collabdocs listening", "addr", srv.Addr, "db_max_conns", maxConns)
	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Error("server error", "error", err.Error())
		os.Exit(1)
	}
	<-idle
	log.Info("stopped cleanly")
}
