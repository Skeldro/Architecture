# Phase 1 — Architecture Decisions

Phase 1 moves CollabDocs from Phase 0's local-files starter to a real platform: state outlives the runtime, concurrent writes do not silently lose content, the environment is reproducible, and an operator can see system health.

## The short version, in plain language

Phase 0 was one program on one laptop writing text files into a folder. That holds up right until you want two things: more than one copy of the program running, and confidence that turning the machine off doesn't lose anyone's work. Everything below is that change.

**1. How the app runs — containers, managed by Kubernetes, five copies.**
We wrap the app in a container: a sealed box holding the program and everything it needs, so it runs identically on any machine instead of only on the one where it was set up. Kubernetes is the thing that keeps those boxes alive — if one crashes it starts a replacement within seconds, with nobody getting woken up. We run five copies so losing one is a non-event rather than an outage. Being honest about it: five is more than this traffic needs. The real argument isn't today's load, it's that growing ten times later becomes a number change instead of a rewrite.

**2. Where documents live — a PostgreSQL database instead of files.**
The moment there are five copies of the app, a folder of text files stops working: each copy has its own folder, so your save lands on one machine and your next page load asks a different one that's never heard of it. Documents move into a database all five share. There's a second reason, and it's the one that actually matters: a database can check, at the instant of writing, whether someone else changed the document while you were typing — and tell you if they did. Files can't do that. They quietly overwrite, which is exactly the bug this phase exists to kill.

**3. What remembers things — the database, and nothing else.**
The app copies remember nothing between requests. All the truth sits in the database. That's what makes it safe to kill, restart, or replace any copy of the app at any moment: there's nothing inside worth keeping. The flip side is that the database becomes the one thing we genuinely cannot lose, so it gets its own storage, its own lifecycle, and its own backups.

**4. How you get it running — the instructions are code in the repo.**
Phase 0's deployment procedure was "however it got set up that day," which nobody can repeat, including the person who did it. Now the container recipe and the deployment files live in git, so a fresh clone becomes a running system via a build and one apply command, inside ten minutes. The one hand-made piece left is the cluster itself, set up once and documented — a known asterisk on the word "reproducible."

**5. How you know it's healthy — the app reports on itself.**
Today the only way to learn something broke is a user complaining. So the app writes structured logs, exposes a page of counters (how many requests, how slow, how many errors, how many save-collisions), and answers a health check that actually tries to talk to the database rather than just confirming it's awake. A health check that only says "I'm running" would cheerfully report all-clear while every single save was failing.

---

**Status:** All five decisions reviewed and approved, 2026-08-19.

---

## Requirement labels (for citation)

**Functional**
- **FR1** — deployable to a clean machine via a documented, reproducible procedure
- **FR2** — state survives recreation of the runtime (process, container, or VM)
- **FR3** — concurrent writes to the same document do not *silently* lose content
- **FR4** — an operator can determine system health without reading raw logs

**Non-functional**
- **NFR1** — read latency p95 < 200 ms at MVP scale
- **NFR2** — sustained MVP-level writes complete without queuing or request loss
- **NFR3** — recovery < 60 s after instance loss
- **NFR4** — supports 10× MVP without re-architecture (re-tuning fine; replacing the storage tier is not)
- **NFR5** — deployment < 10 minutes from fresh git clone to working instance

**Quality-attribute ranking (tie-breaks):** Performance > Resilience > Scalability > Consistency > Maintainability

---

## Standing assumptions (stated, not derived)

| # | Assumption | Basis |
|---|---|---|
| A1 | **MVP = 1,000 concurrent users** | Worksheet: ~1% of the year-1 goal of 100,000 concurrent |
| A2 | **"Concurrent user" = an open tab with activity in the last 5 minutes** | Andrew's definition. Note: no WebSockets this phase, so users hold no persistent connection — concurrency is a *request arrival rate*, not a connection count |
| A3 | **20% of concurrent users are actively typing at any instant** | Estimate. A2's bucket is looser than "typing now", so this is a separate assumption and the most load-bearing unverified number in this document |
| A4 | **Auto-save cadence = 2 seconds** | Andrew's call, upgraded from 5 s for perceived seamlessness |
| A5 | **Average document size ≈ 10 KB** | Estimate for storage-growth math |
| A6 | **~20 registered users per concurrent user; ~10 documents each** | Estimate for storage-growth math |

### The math

Writes/sec = `concurrent users × typing fraction ÷ auto-save interval`

| | MVP (1,000 concurrent) | 10× MVP (10,000 concurrent) |
|---|---|---|
| Writes/sec | `1000 × 0.20 ÷ 2` = **100/s** | **1,000/s** |
| Write bandwidth (full-document overwrite, A5) | ~1 MB/s | ~10 MB/s (roughly 2× with WAL) |
| Per-instance load at 5 replicas | 20 writes/s | 200 writes/s |
| Stored data (A5, A6) | 20,000 users × 10 docs × 10 KB ≈ **2 GB** | ≈ 20 GB; year-1 (100× concurrent) ≈ **200 GB** |

**Cost of the 2-second cadence:** 5 s → 2 s is **2.5× the write load**, spent entirely on the storage tier because saves are full-document overwrites. Backed by the QA ranking (Performance ranks first). A debounced save — 2 s after typing *stops* rather than every 2 s — would cut writes by roughly an order of magnitude for the same felt smoothness, and remains available if the write path becomes the bottleneck.

**Compute is not the constraint.** At 100 writes/sec, a single Go instance is at a few percent utilization; goroutines plus the netpoller make this load unremarkable. Instance count below is chosen for resilience and deploy behaviour, not throughput.

---

## Decision 1 — Compute model and horizontal scaling

*Affects FR1, NFR3, NFR4, Resilience.*

- **Compute shape:** container, on an orchestrator (Kubernetes).
- **Instance count at MVP:** 5 replicas, stateless.
- **Load balancing:** orchestrator-built-in (Kubernetes Service / Ingress). No self-run nginx or HAProxy.
- **Restart authority:** Kubernetes. Pod dies → rescheduled automatically, well inside NFR3's 60 s.

**Why:**
- The container **image is the reproducibility artifact** FR1 asks for — it builds identically anywhere, which is precisely what Phase 0's "works on the machine where it was set up" failure lacked. A bare VM would leave reproducibility as a documentation problem rather than a mechanism.
- Restart supervision is a property of the orchestrator, not something to hand-roll — this is what serves NFR3.
- **NFR4 is the real case for Kubernetes:** 10× headroom becomes a `replicas` change plus a bigger database, not a redesign. (Rejected justification: "it works at any scale." The worksheet requires components *sized to the scale target*; "scales to anything" is over-architecting wearing a justification's jacket.)
- Serverless was rejected as structurally hostile: no long-lived process, no local disk, cold starts fighting NFR1, and the entire Phase 0 codebase assumes a server.

**Trade-offs and things given up:**
1. **The SPOF relocated rather than disappeared.** Five stateless replicas do not remove the single point of failure — they move it downstream to the shared storage tier (Decision 2), plus the cluster control plane and, in a single-AZ deployment, the zone itself.
2. **Operational complexity for a solo developer.** A cluster is now a thing that must be maintained, upgraded, and paid for (control plane + 5 replicas + managed database), to serve a load one instance could carry. Maintainability ranks last in the QA ranking, which is the ranking's own permission slip for this — but it is a real cost, not a free one.
3. **Local debuggability is gone.** Phase 0 let you `cat docs/whatever.txt` to see the truth. Logs are now spread across five pods and state lives behind a network hop.
4. **In-process state of any kind becomes illegal.** A Go mutex protects one pod, not five — so FR3's concurrency control *must* live in the storage tier. Likewise no in-process caching or session state. **This decision pre-constrains Decision 2.**
5. **Phase 0's local files are dead on arrival.** Five pods means five filesystems: a save lands on pod 3, the next read hits pod 1, and the document appears to revert. Shared storage is now mandatory, not optional.

**Known soft spots, accepted knowingly:**
- **The replica count is chosen, not derived.** At MVP, instance #2 buys removal of the process SPOF and rolling deploys with no downtime; instances #3–#5 buy nothing measurable at 100 writes/sec. Five stands as deliberate resilience headroom rather than the output of a load calculation — expect the question, and answer it as a resilience choice.
- **NFR5 depends on a pre-existing cluster.** The < 10-minute clock holds where the cluster is managed and already provisioned (GKE/EKS) and deploying means build image + `kubectl apply`. Cluster provisioning sits outside the clock; that reading carries into `phase1-spec.md`.

**Not given up by this decision (recorded to prevent a mis-citation):** real-time collaboration is absent because Constraint 3 forbids WebSockets this phase, not because of the compute choice. Relatedly, one user's paste being clobbered by another's auto-save is the **lost-update problem** — that is FR3, the thing Phase 1 must fix, and it is not a SPOF.

---

## Decision 2 — Storage type and engine, and read-side caching

*Affects FR3, NFR1, NFR2, NFR4.*

- **Primary storage:** relational — **PostgreSQL**, single managed instance. One table: `documents(id, title, content, version, updated_at)`.
- **Concurrency control:** optimistic, via the `version` column. Save issues `UPDATE documents SET content=$1, version=version+1 WHERE id=$2 AND version=$3`; zero rows affected means someone else wrote first, and the user is told (HTTP 409) rather than silently overwritten.
- **Read-side caching:** **none at MVP.**
- **Invalidation policy:** not applicable. If a cache is added later, invalidate **on write**, not by TTL — a TTL would serve a stale document to someone actively editing it.
- **Driver use:** `database/sql` with `pgx`, called directly in route handlers. Constraint 1 declares this correct for this phase; the storage abstraction is Phase 3's job to earn.

**Why:**
- **FR3 is the deciding requirement.** It demands writes not be lost *silently* — a floor, not full consistency (note Consistency ranks fourth of five). Optimistic version-checking is the cheapest mechanism that converts a silent clobber into a visible conflict, and it works across five pods because the check is atomic inside the database. A shared file volume offers no such primitive.
- **NFR1/NFR2 at MVP are comfortable:** 100 writes/sec and a working set of ~2 GB sit well inside a single modest Postgres instance with the hot data in RAM; p95 < 200 ms is dominated by network and rendering, not the query.
- **NFR4 is satisfied by tuning, not replacement:** 1,000 writes/sec at 10× needs a larger instance and WAL tuning — allowed explicitly ("re-tuning is fine; replacing the storage tier is not").
- No cache at MVP is the Lean answer: at this scale a cache buys nothing measurable while adding an invalidation problem that FR3 makes strictly harder. Deferred until a measurement justifies it.

**Rejected:** document store (Mongo) — the data is trivially relational and the conflict primitive is no better; shared file volume (ReadWriteMany) — no atomic compare-and-swap, so FR3 stays unsolved; key-value (Redis-with-persistence) — durability guarantees weaker than the requirement warrants.

**Trade-offs and things given up:**
1. **Documents stop being plain files.** The Phase 0 property of inspecting state with `cat` and `grep` is gone, and with it the "no database to operate" simplicity.
2. **A single Postgres instance is the new SPOF and the write ceiling.** No read replicas, no sharding, no multi-AZ failover at MVP — its loss is the slowest recovery path in the system.
3. **Auto-save every 2 seconds against optimistic locking means two people editing one document will generate constant conflicts.** That is *correct* behaviour under FR3 and Constraint 3 — the conflict surfaces instead of silently destroying work — but it is a visibly poor experience, and it is precisely the gap real-time collaboration closes in Phase 2.

---

## Decision 3 — State vs compute placement

*Affects FR2, NFR3.*

- **State lives** in the managed PostgreSQL instance, on its own volume with its own lifecycle, in a different failure domain from the application pods.
- **Compute runs** as stateless containers on Kubernetes nodes. A pod holds nothing durable — no local writes, no in-memory authority.
- **The boundary that makes state outlast compute** is the network plus the database's own managed storage: destroying every pod, node, or the whole Deployment loses nothing.
- **Recovery procedure:**
  - *Pod loss* — automatic, Kubernetes reschedules; seconds. Satisfies NFR3.
  - *Node loss* — automatic, rescheduled onto another node.
  - *Database loss* — restore from the managed automated backup. **Time-bounded but measured in minutes, not seconds.**

**Why:** FR2 asks for state to survive runtime recreation, which is exactly a demand that state and compute occupy separate lifecycles. Statelessness is also what makes Decision 1's five interchangeable replicas coherent at all.

**Trade-off stated honestly:** NFR3's 60-second clock is met for *compute* loss only. Database loss is the one path in this architecture that exceeds it, and closing that gap means replication and automated failover — deliberately out of scope at MVP scale.

---

## Decision 4 — Reproducibility mechanism

*Affects FR1, NFR5.*

- **Container image + Kubernetes manifests committed to the repository**, plus a documented deploy command. A clean machine becomes a working instance by: clone → build image → push → `kubectl apply -f k8s/` → verify `/healthz`.
- **Cluster provisioning is a documented one-time step, outside the 10-minute clock** (see Decision 1's open item). Infrastructure-as-Code for the cluster itself is deferred.
- README carries the procedure; no configuration-management tool (Ansible/Chef/Puppet).

**Why:** FR1 asks for a documented, reproducible procedure, and the image plus declarative manifests *are* that procedure — executable rather than prose, which is the failure mode Phase 0 exhibited. NFR5's < 10 minutes is achievable on this path because the only per-deploy work is an image build and an apply.

**What changes at 15 developers versus 1:** CI takes over image build and push on merge; the cluster moves to Terraform so it stops being a hand-made snowflake nobody can rebuild; per-developer preview namespaces appear; secrets move into a real secret manager rather than living beside the manifests. At one developer, every one of those is waste — which is the point of sizing to the current scale.

**Trade-off:** the cluster is the one hand-provisioned thing in the system, so "reproducible" has an asterisk on it until IaC lands.

---

## Decision 5 — Observability baseline

*Affects FR4, NFR3.*

- **Logging:** structured JSON to stdout, one object per request (method, path, status, duration_ms, doc id where relevant). Kubernetes collects stdout; JSON means the logs are queryable later without re-instrumenting the application.
- **Health endpoint:** `GET /healthz` **verifies storage reachability** — it executes `SELECT 1` against Postgres rather than merely proving the process is alive. Split into liveness (process up) and readiness (database reachable) so Kubernetes restarts a wedged pod but merely removes a database-blind pod from rotation.
- **Metrics:** in-process counters exposed at `GET /metrics` in Prometheus text format — request rate, p50/p95 latency, error rate, and **save-conflict count** (the direct FR3 signal: it shows whether users are actually colliding).

**Why:** FR4 requires health to be determinable without reading raw logs, which is exactly the difference between a status endpoint and a log stream. A liveness check that only proves the process is running would report "healthy" while every save fails — the specific failure FR4 exists to catch, hence the storage ping.

**Explicitly deferred to Phase 2:** third-party monitoring services (Sentry, Datadog), distributed tracing, alerting rules, and dashboards. Phase 1 builds the signals; Phase 2 decides who watches them.

---

## Decisions Made — digest

> **Format note:** the worksheet requires the 5-field digest format defined in `project-overview.md`, which is not in the repository. The five fields below are a stand-in and should be reformatted once that file is available.

| Decision | Choice | Why (primary requirement) | Traded away | Revisit if |
|---|---|---|---|---|
| 1. Compute | Containers on Kubernetes, 5 stateless replicas, built-in Service LB | NFR4 — 10× is a `replicas` change | Simplicity, cost, local debuggability; SPOF moves to storage | The replica count is never justified by load, or cluster upkeep outweighs its benefit |
| 2. Storage | Single managed PostgreSQL; optimistic concurrency via `version` column; no cache | FR3 — conflicts surface instead of silently clobbering | Plain-file inspectability; single-instance write ceiling | Write rate approaches the instance ceiling, or read latency breaches NFR1 |
| 3. Placement | State in managed Postgres; compute fully stateless | FR2 — state outlives any runtime | Sub-60 s recovery for *database* loss specifically | Database recovery time becomes unacceptable → replication + failover |
| 4. Reproducibility | Image + committed K8s manifests + README; cluster provisioned once, by hand | FR1, NFR5 — executable procedure, < 10 min deploy | Cluster itself is not yet reproducible (no IaC) | A second environment is needed, or the team grows past 1 |
| 5. Observability | JSON logs to stdout; `/healthz` pings the database; `/metrics` with conflict counter | FR4 — health without reading logs | External monitoring, tracing, alerting (Phase 2) | Incidents are being found by users rather than by signals |

---

## Still outstanding

Not open decisions — the five above are settled. These are gaps in the surrounding paperwork:

1. **`project-overview.md` is missing** — needed for the required *Decisions Made* digest format and the *Working with AI agent* section.
2. **Assumption A3 (20% typing fraction)** is the most load-bearing unverified number in this document; every write figure scales linearly with it.
3. **`phase1-spec.md` not yet written** — the worksheet's sequence is decisions first, then translate them into the build spec.
