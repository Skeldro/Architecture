package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// Decision 5 builds the signals in-process and defers the watcher to Phase 2,
// so Constraint 6 (no third-party monitoring service) holds. The exposition
// format is Prometheus text so a scraper can be pointed at it later without
// re-instrumenting anything; the computed p50/p95 gauges exist so that a human
// reading /metrics today gets an answer without a scraper — which is what FR4
// actually asks for.

var latencyBounds = []float64{1, 2, 5, 10, 25, 50, 100, 200, 500, 1000, 2500}

type histogram struct {
	counts []uint64 // len(latencyBounds)+1; final element is the overflow bucket
	sum    float64
	count  uint64
}

func newHistogram() *histogram {
	return &histogram{counts: make([]uint64, len(latencyBounds)+1)}
}

func (h *histogram) add(ms float64) {
	h.sum += ms
	h.count++
	// SearchFloat64s returns the first index whose bound is >= ms, which is the
	// bucket that contains it. A value above every bound returns len(bounds),
	// which is the overflow slot.
	h.counts[sort.SearchFloat64s(latencyBounds, ms)]++
}

// quantile interpolates linearly inside the bucket that contains the target
// observation, which is how Prometheus itself estimates quantiles.
func (h *histogram) quantile(q float64) float64 {
	if h.count == 0 {
		return 0
	}
	target := q * float64(h.count)
	var cum uint64
	lower := 0.0
	for i, ub := range latencyBounds {
		cum += h.counts[i]
		if float64(cum) >= target {
			inBucket := h.counts[i]
			if inBucket == 0 {
				return ub
			}
			start := float64(cum - inBucket)
			return lower + ((target-start)/float64(inBucket))*(ub-lower)
		}
		lower = ub
	}
	return lower // everything fell in the overflow bucket
}

type metrics struct {
	mu        sync.Mutex
	requests  map[string]uint64
	hists     map[string]*histogram
	conflicts atomic.Uint64
	started   time.Time
}

func newMetrics() *metrics {
	return &metrics{
		requests: map[string]uint64{},
		hists:    map[string]*histogram{},
		started:  time.Now(),
	}
}

// knownRoutes bounds metric label cardinality. Without this, every unmatched
// path an unknown client requests would create a new label set and the metrics
// endpoint would grow without limit.
var knownRoutes = map[string]bool{
	"/": true, "/create": true, "/doc": true, "/save": true,
	"/healthz": true, "/livez": true, "/metrics": true,
}

func (m *metrics) observe(method, route string, status int, ms float64) {
	if !knownRoutes[route] {
		route = "other"
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.requests[fmt.Sprintf("%s|%s|%d", method, route, status)]++
	h, ok := m.hists[route]
	if !ok {
		h = newHistogram()
		m.hists[route] = h
	}
	h.add(ms)
}

func (m *metrics) handler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		m.mu.Lock()
		defer m.mu.Unlock()
		w.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")

		fmt.Fprintf(w, "# HELP collabdocs_uptime_seconds Seconds since process start.\n")
		fmt.Fprintf(w, "# TYPE collabdocs_uptime_seconds gauge\n")
		fmt.Fprintf(w, "collabdocs_uptime_seconds %.1f\n", time.Since(m.started).Seconds())

		fmt.Fprintf(w, "# HELP collabdocs_http_requests_total Requests by method, route and status.\n")
		fmt.Fprintf(w, "# TYPE collabdocs_http_requests_total counter\n")
		keys := make([]string, 0, len(m.requests))
		for k := range m.requests {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		var total, errors5xx uint64
		for _, k := range keys {
			method, route, status := splitKey(k)
			n := m.requests[k]
			total += n
			if status >= 500 {
				errors5xx += n
			}
			fmt.Fprintf(w, "collabdocs_http_requests_total{method=%q,route=%q,status=\"%d\"} %d\n",
				method, route, status, n)
		}

		fmt.Fprintf(w, "# HELP collabdocs_http_error_ratio Fraction of requests answered 5xx.\n")
		fmt.Fprintf(w, "# TYPE collabdocs_http_error_ratio gauge\n")
		ratio := 0.0
		if total > 0 {
			ratio = float64(errors5xx) / float64(total)
		}
		fmt.Fprintf(w, "collabdocs_http_error_ratio %.6f\n", ratio)

		fmt.Fprintf(w, "# HELP collabdocs_request_duration_ms Request duration in milliseconds.\n")
		fmt.Fprintf(w, "# TYPE collabdocs_request_duration_ms histogram\n")
		routes := make([]string, 0, len(m.hists))
		for rt := range m.hists {
			routes = append(routes, rt)
		}
		sort.Strings(routes)
		for _, rt := range routes {
			h := m.hists[rt]
			var cum uint64
			for i, ub := range latencyBounds {
				cum += h.counts[i]
				fmt.Fprintf(w, "collabdocs_request_duration_ms_bucket{route=%q,le=\"%g\"} %d\n", rt, ub, cum)
			}
			fmt.Fprintf(w, "collabdocs_request_duration_ms_bucket{route=%q,le=\"+Inf\"} %d\n", rt, h.count)
			fmt.Fprintf(w, "collabdocs_request_duration_ms_sum{route=%q} %.3f\n", rt, h.sum)
			fmt.Fprintf(w, "collabdocs_request_duration_ms_count{route=%q} %d\n", rt, h.count)
		}

		// Directly readable percentiles. NFR1 is a statement about the "/doc"
		// route specifically, so it can be checked by eye here.
		fmt.Fprintf(w, "# HELP collabdocs_request_duration_p50_ms Estimated median duration.\n")
		fmt.Fprintf(w, "# TYPE collabdocs_request_duration_p50_ms gauge\n")
		for _, rt := range routes {
			fmt.Fprintf(w, "collabdocs_request_duration_p50_ms{route=%q} %.3f\n", rt, m.hists[rt].quantile(0.50))
		}
		fmt.Fprintf(w, "# HELP collabdocs_request_duration_p95_ms Estimated 95th percentile duration.\n")
		fmt.Fprintf(w, "# TYPE collabdocs_request_duration_p95_ms gauge\n")
		for _, rt := range routes {
			fmt.Fprintf(w, "collabdocs_request_duration_p95_ms{route=%q} %.3f\n", rt, m.hists[rt].quantile(0.95))
		}

		fmt.Fprintf(w, "# HELP collabdocs_save_conflicts_total Saves rejected because another writer won.\n")
		fmt.Fprintf(w, "# TYPE collabdocs_save_conflicts_total counter\n")
		fmt.Fprintf(w, "collabdocs_save_conflicts_total %d\n", m.conflicts.Load())

		// Decision 2's ceiling is a database-connection ceiling, so the pool is
		// the number to watch. wait_count above zero means instances are
		// queueing for connections, which is the signal to add PgBouncer.
		s := db.Stats()
		fmt.Fprintf(w, "# HELP collabdocs_db_pool Connection pool state.\n")
		fmt.Fprintf(w, "# TYPE collabdocs_db_pool gauge\n")
		fmt.Fprintf(w, "collabdocs_db_pool{state=\"max_open\"} %d\n", s.MaxOpenConnections)
		fmt.Fprintf(w, "collabdocs_db_pool{state=\"open\"} %d\n", s.OpenConnections)
		fmt.Fprintf(w, "collabdocs_db_pool{state=\"in_use\"} %d\n", s.InUse)
		fmt.Fprintf(w, "collabdocs_db_pool{state=\"idle\"} %d\n", s.Idle)
		fmt.Fprintf(w, "# HELP collabdocs_db_pool_wait_total Times a request waited for a connection.\n")
		fmt.Fprintf(w, "# TYPE collabdocs_db_pool_wait_total counter\n")
		fmt.Fprintf(w, "collabdocs_db_pool_wait_total %d\n", s.WaitCount)
		fmt.Fprintf(w, "# HELP collabdocs_db_pool_wait_seconds_total Total time spent waiting for a connection.\n")
		fmt.Fprintf(w, "# TYPE collabdocs_db_pool_wait_seconds_total counter\n")
		fmt.Fprintf(w, "collabdocs_db_pool_wait_seconds_total %.6f\n", s.WaitDuration.Seconds())
	}
}

// splitKey reverses the "method|route|status" map key.
func splitKey(k string) (method, route string, status int) {
	parts := strings.SplitN(k, "|", 3)
	if len(parts) != 3 {
		return k, "", 0
	}
	status, _ = strconv.Atoi(parts[2])
	return parts[0], parts[1], status
}
