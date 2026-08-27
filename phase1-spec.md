# Phase 1 — Build Specification

What to build, derived from `phase1-decisions.md`. The decisions document captures *why*; this captures *what*.

> **Ordering note, stated honestly:** the course sequence is decisions → spec → build. This specification was written after the implementation rather than before it, so it documents the contract the build satisfies rather than one it was built against. The decisions it derives from were settled first.

---

## HTTP surface

| Method | Path | Behaviour |
|---|---|---|
| `GET` | `/` | List documents (id, title), ordered by title, plus the create form. Any unmatched path returns 404 — `ServeMux` treats `/` as a catch-all, so this is explicit. |
| `POST` | `/create` | Insert a document with the given title. `303` to `/doc?id=N`. `400` on empty title, `409` on duplicate. |
| `GET` | `/doc?id=N` | Editor form carrying hidden `id` and `version`. `404` if absent, `400` on unparseable id. |
| `POST` | `/save` | Optimistic update. `303` back to the editor on success; `409` with the conflict page when the version is stale; `404` if the document is gone. |
| `GET` | `/healthz` | Readiness. Executes `SELECT 1`. `200 {"status":"healthy"}` or `503 {"status":"unhealthy"}`. |
| `GET` | `/livez` | Liveness. `200 {"status":"alive"}` whenever the process is running, regardless of storage. |
| `GET` | `/metrics` | Prometheus text exposition. Excluded from its own instrumentation. |

## Schema

```sql
CREATE TABLE IF NOT EXISTS documents (
    id         BIGSERIAL   PRIMARY KEY,
    title      TEXT        NOT NULL UNIQUE,
    content    TEXT        NOT NULL DEFAULT '',
    version    INTEGER     NOT NULL DEFAULT 1,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
```

Applied on boot by every instance, inside a transaction holding a PostgreSQL advisory lock, because instances start simultaneously.

Documents are now identified by `id` rather than by filename. That retires two costs Phase 0 accepted knowingly: titles no longer need to be filename-safe, and the uniqueness constraint is enforced by the database rather than by the filesystem.

## The concurrency rule (FR3)

Save is a single statement:

```sql
UPDATE documents SET content = $1, version = version + 1, updated_at = now()
WHERE id = $2 AND version = $3
```

Zero rows affected means another writer got there first. The version predicate is what makes the read-modify-write atomic *inside the database*, which is the only place it can be correct — an in-process lock protects one instance, and there are several.

**Rejecting the write is not sufficient.** FR3 says content must not be lost *silently*, so the conflict response must carry back the rejected writer's own text alongside the now-current version, and offer to save it over the top. Losing the race must cost the user a decision, never their work.

## Configuration

Environment only — nothing read from the local filesystem, because a clean machine has none of it.

| Variable | Required | Default |
|---|---|---|
| `DATABASE_URL` | yes | — |
| `PORT` | no | `8080` |
| `DB_MAX_CONNS` | no | `10` |

## Runtime behaviour

- **Connection pool:** `MaxOpenConns` and `MaxIdleConns` both set to `DB_MAX_CONNS`; `ConnMaxLifetime` 5 minutes. The Go default of unlimited open connections must be overridden explicitly, and idle defaults to 2, which causes reconnect churn.
- **Line endings:** submitted content has `\r\n` normalised to `\n` before storage.
- **Shutdown:** `SIGTERM` drains in-flight requests for up to 20 seconds before exit. Without this, every rolling deploy drops live requests and Decision 1's zero-downtime claim is false.
- **Logging:** structured JSON to stdout, one record per request with method, path, status, duration and document id. Conflicts log at warning level.

## Metrics

- `collabdocs_http_requests_total{method,route,status}` — route labels restricted to a known set so cardinality stays bounded.
- `collabdocs_http_error_ratio` — 5xx as a fraction of all requests.
- `collabdocs_request_duration_ms` — histogram per route, plus `_p50_ms` and `_p95_ms` gauges so a human reading the endpoint gets an answer without a scraper.
- `collabdocs_save_conflicts_total` — the FR3 signal: whether users are actually colliding.
- `collabdocs_db_pool{state}` and `collabdocs_db_pool_wait_total` — Decision 2's ceiling is a connection ceiling, so waits above zero are the signal to add a pooler.

## Deployment

Container image plus `deploy/main.tf`: a Cloud Run service with `min_instance_count = 2`, `max_instance_count = 8` and CPU always allocated, against a Cloud SQL PostgreSQL instance with automated backups and point-in-time recovery, reached over the Cloud SQL socket so no public database IP exists. Startup and liveness probes point at `/healthz` and `/livez`.

## Verification plan

| Criterion | How it is checked |
|---|---|
| Create, edit, list end to end | Test suite |
| FR1 — environment-only configuration; repeatable schema init | Test suite, including four concurrent initialisations |
| FR2 — state survives runtime recreation | Test suite (pool destroyed, new instance reads back) and a real process restart |
| FR3 — concurrent writes do not silently lose content | Test suite (two concurrent saves at the same version) and two live browser tabs |
| FR4 — health detects unavailable storage | Test suite against both a live and an unreachable database |
| NFR1/NFR2 — latency and sustained writes | `cmd/loadtest`: 100 virtual users, 1 write/sec, 60 seconds, latency reported per 10-second window |

## Out of scope this phase

Authentication, real-time synchronisation, any abstraction over storage, extraction of business rules from handlers, use cases or DTOs, and third-party monitoring services. Each belongs to a named later phase.
