# Architecture Course

Learning software architecture through worksheets and a course project. This folder is the territory: worksheets, decisions, session handoffs, and the project itself.

The project is **CollabDocs** — a collaborative document editor, built in phases. Each phase adds architecture to a deliberately naive core, and each phase's problems are the ones the previous phase created.

---

## Running it

No container runtime is required. `scripts/localpg.sh` runs a userland PostgreSQL, downloading it on first use.

```
make dev
```

Then open <http://localhost:8080>.

```
make test        # FR1-FR4 suite, against its own database
make loadtest    # 100 virtual users x 1 write/sec for 60s, against a running instance
make check       # fmt + vet + test
make db-stop     # stop the local database
```

Machines with a container runtime can use `compose.yaml` instead.

### Configuration

The application reads only environment variables — nothing from the local filesystem, which is what makes it deployable to a clean machine.

| Variable | Required | Default | Meaning |
|---|---|---|---|
| `DATABASE_URL` | yes | — | PostgreSQL connection string |
| `PORT` | no | `8080` | Listen port |
| `DB_MAX_CONNS` | no | `10` | Connection-pool ceiling per instance |

`DB_MAX_CONNS` matters more than it looks: `database/sql` defaults to *unlimited*, and the per-instance pool multiplied by the maximum instance count must stay below the server's `max_connections`. See Decision 2.

## Deploying

The deployment path is a container image plus `deploy/main.tf`, which declares the Cloud Run service and the Cloud SQL instance. From a fresh clone:

```
docker build -t REGION-docker.pkg.dev/PROJECT/collabdocs/app:TAG .
docker push REGION-docker.pkg.dev/PROJECT/collabdocs/app:TAG
cd deploy && terraform init && terraform apply -var project_id=PROJECT -var image=REGION-docker.pkg.dev/PROJECT/collabdocs/app:TAG
curl "$(terraform output -raw url)/healthz"
```

There is no cluster to provision — that is deliberate, and it is what makes the procedure complete rather than nearly complete.

**Honest status of this path:** the Terraform has never been applied and the image has never been built, because both require cloud credentials and a container runtime that this development environment does not have. Everything below the deployment layer *is* verified — see the next section.

## Phase 1 verification status

| Acceptance criterion | Status |
|---|---|
| 1. Clean-machine deploy in under 10 minutes | **Partly** — `make dev` verified end to end; the cloud path is written but unapplied |
| 2. State survives destroying and recreating the runtime | **Verified** — process stopped and restarted; documents and content intact |
| 3. Two simultaneous saves do not silently lose content | **Verified** — automated concurrency test, and two live browser tabs |
| 4. Health reports healthy, and detects unavailable storage | **Verified** — `/healthz` returns 200 healthy, 503 unhealthy against dead storage |
| 5. Load test completes without errors or unbounded latency growth | **Verified** — 12,000 requests, 0 errors, read p95 1.4 ms, flat per-window latency |
| 6. Automated test suite runs from a clean checkout | **Verified** — 10 tests covering FR1 through FR4 |

Load-test figures are local, with no network between client and server, so they validate the *server-side* portion of NFR1's 200 ms budget (which Decision 2 sizes at 50 ms) rather than end-user latency.

## Why the code looks like this

Phase 1's constraints are deliberately in force, and several would be defects in an ordinary codebase:

- **No abstraction around storage.** SQL is written inline in the route handlers, and the repetition that causes is intentional. Phase 3 earns the abstraction by having a second implementation to justify it.
- **No authentication.** Every request is anonymous; anyone with a URL can read or edit anything. Phase 2.
- **No real-time synchronisation.** Changes are invisible to other sessions until refresh, and two people editing one document will collide every save. Phase 2.
- **No business-rule extraction, no use cases or DTOs.** Phases 4 and 5.

The `app` struct is not a storage abstraction — it holds the connection pool so that tests can construct independent instances against one database, which acceptance criterion 6 requires.

---

## Course method (standing rules)

- **Socratic**: Andrew answers first, then gets corrected. The tutor does not hand out answers to open items.
- **Plain-language first**: he states his approach in plain language before any code is written.
- **No "self explanatory"**: every item gets articulated, even the obvious ones.
- Read the latest file in `handoffs/` before resuming — it carries his recurring error patterns.

## Workflow

Process decisions (how we work — not what we build; those go in the decision documents):

- **Own repo per project** — independent history, clean presentation.
- **Branch per phase** (`phase-0`, `phase-1`, …) cut from `main`, merged back when the phase completes. `main` only ever holds finished phases; the merge is a checkpoint matching the mentor-session cadence.
- **Decision logs record architecture decisions only**, one entry at the time the decision is made — reasoning written afterwards comes out revisionist and too tidy. These are the documents presented in mentor sessions.

## Files

| Path | What it is |
|---|---|
| `main.go`, `metrics.go` | The application |
| `main_test.go` | FR1-FR4 test suite |
| `cmd/loadtest/` | Load generator for acceptance criterion 5 |
| `Dockerfile`, `compose.yaml`, `deploy/` | Decision 4's reproducibility mechanism |
| `scripts/localpg.sh` | Userland PostgreSQL for development and tests |
| `decisions.md` | Phase 0 architecture decisions |
| `phase1-decisions.md` | Phase 1 architecture decisions, with the sizing math |
| `phases/` | Phase specifications |
| `worksheets/` | Worksheet answers |
| `glossary.md` | Acronyms and jargon |
| `handoffs/` | Dated session summaries |

Phase 0's file-based implementation is in git history, on `main` before the phase-1 merge.

## Status

- **Phase 0** — complete and merged. Files on disk, no database, no architecture, messy on purpose.
- **Phase 1** — decisions approved and revised after mentor review; implementation built and verified locally (see the table above).
- **Phase 2** — not started. Identity, real-time synchronisation, and operational visibility.

Still owed from worksheet 1: an Infrastructure one-liner, and the explanation of the Scrum/stubs/frameworks/plug-ins synthesis quote.
