# Phase 1 — Architecture Decisions

Phase 1 moves CollabDocs from Phase 0's local-files starter to a real platform: state outlives the runtime, concurrent writes do not silently lose content, the environment is reproducible, and an operator can see system health.

## The short version, in plain language

Phase 0 was one program on one laptop writing text files into a folder. That holds up right until you want two things: more than one copy of the program running, and confidence that turning the machine off doesn't lose anyone's work. Everything below is that change.

**1. How the app runs — a container, on a platform that runs containers for you, two copies.**
We wrap the app in a container: a sealed box holding the program and everything it needs, so it runs identically on any machine instead of only on the one where it was set up. A managed container platform keeps those boxes alive — if one dies it starts a replacement within seconds, with nobody getting woken up. Two copies, because one copy means every crash and every deploy is an outage, and the second copy removes that. Not five: at this traffic a single copy runs at under a third of its capacity, so the second is for survival, not for speed. The platform adds more automatically when traffic grows.

**2. Where documents live — a PostgreSQL database instead of files.**
The moment there are two copies of the app, a folder of text files stops working: each copy has its own folder, so your save lands on one machine and your next page load asks a different one that's never heard of it. Documents move into a database both copies share. There's a second reason, and it's the one that actually matters: a database can check, at the instant of writing, whether someone else changed the document while you were typing — and tell you if they did. Files can't do that. They quietly overwrite, which is exactly the bug this phase exists to kill.

**3. What remembers things — the database, and nothing else.**
The app copies remember nothing between requests. All the truth sits in the database. That's what makes it safe to kill, restart, or replace any copy of the app at any moment: there's nothing inside worth keeping. The flip side is that the database becomes the one thing we genuinely cannot lose, so it gets its own storage, its own lifecycle, and its own backups.

**4. How you get it running — the instructions are code in the repo.**
Phase 0's deployment procedure was "however it got set up that day," which nobody can repeat, including the person who did it. Now the container recipe, the service configuration and the infrastructure definition all live in git, so a fresh clone becomes a running system via a build and one deploy command. There is no cluster to hand-build, which is precisely why this is now reproducible rather than nearly reproducible.

**5. How you know it's healthy — the app reports on itself.**
Today the only way to learn something broke is a user complaining. So the app writes structured logs, exposes a page of counters (how many requests, how slow, how many errors, how many save-collisions), and answers a health check that actually tries to talk to the database rather than just confirming it's awake. A health check that only says "I'm running" would cheerfully report all-clear while every single save was failing.

---

**Status:** All five decisions reviewed and approved, 2026-08-19. **Revised 2026-08-26** following mentor review: the compute shape moved from self-managed Kubernetes to a managed container platform, the instance count is now derived rather than chosen, and connection-pool sizing and a read-latency budget have been added.

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
| A2 | **"Concurrent user" = an open tab with activity in the last 5 minutes** | No WebSockets this phase, so users hold no persistent connection — concurrency is a *request arrival rate*, not a connection count |
| A3 | **20% of concurrent users are actively typing at any instant** | Estimate. A2's bucket is looser than "typing now", so this is a separate assumption and the most load-bearing unverified number in this document |
| A4 | **Auto-save cadence = 2 seconds** | Chosen for perceived seamlessness, upgraded from 5 s |
| A5 | **Average document size ≈ 10 KB** | Estimate for storage and bandwidth math |
| A6 | **~20 registered users per concurrent user; ~10 documents each** | Estimate for storage-growth math |
| A7 | **Reads ≈ 50/sec at MVP** | Each concurrent user triggers roughly one page load per 20 s of activity (opening a document, returning to the list) |
| A8 | **One application instance sustains ≈ 500 req/s** | Conservative for a Go handler doing one indexed query plus template render; the real figure is higher, and using a low estimate makes the sizing argument safer rather than weaker |

### The math

Writes/sec = `concurrent users × typing fraction ÷ auto-save interval`

| | MVP (1,000 concurrent) | 10× MVP (10,000 concurrent) |
|---|---|---|
| Writes/sec | `1000 × 0.20 ÷ 2` = **100/s** | **1,000/s** |
| Reads/sec (A7) | ~50/s | ~500/s |
| **Total requests/sec** | **~150/s** | **~1,500/s** |
| Write bandwidth (full-document overwrite, A5) | ~1 MB/s | ~10 MB/s (roughly 2× with WAL) |
| Stored data (A5, A6) | 20,000 users × 10 docs × 10 KB ≈ **2 GB** | ≈ 20 GB; year-1 (100× concurrent) ≈ **200 GB** |

**Cost of the 2-second cadence:** 5 s → 2 s is **2.5× the write load**, spent entirely on the storage tier because saves are full-document overwrites. Backed by the QA ranking (Performance ranks first). A debounced save — 2 s after typing *stops* rather than every 2 s — would cut writes by roughly an order of magnitude for the same felt smoothness, and remains available if the write path becomes the bottleneck.

**Compute is not the constraint.** At ~150 req/s against A8, a single instance runs at roughly 30% utilisation. Instance count below is therefore derived from survival requirements, not throughput.

---

## Decision 1 — Compute model and horizontal scaling

*Affects FR1, NFR3, NFR4, Resilience.*

- **Compute shape:** container, on a **managed container platform** — Google Cloud Run, or AWS ECS Fargate as the equivalent. No self-managed cluster.
- **Instance count at MVP:** **2**, minimum, always warm. Platform autoscaling above that on concurrency, ceiling of **8** (see Decision 2 — the ceiling is set by database connections, not by compute).
- **Load balancing:** provided by the platform. Nothing self-run.
- **Restart authority:** the platform. An unhealthy or dead instance is replaced automatically, well inside NFR3's 60 s.

### Deriving the instance count

The previous draft said "5 replicas" without derivation, which was the weakest sentence in this document. The derivation:

| Instances | Utilisation at MVP (150 req/s, A8) | Behaviour when one is lost |
|---|---|---|
| **1** | 30% | Full outage until a replacement starts. Every deploy is downtime. |
| **2** | 15% each | Survivor absorbs all 150 req/s at **30% utilisation** — comfortable. No outage, and deploys roll with no downtime. |
| **3+** | 10% each | Buys nothing at this load; the two-instance survivor case is already comfortable. |

**Therefore: 2.** The second instance removes the process as a single point of failure and makes zero-downtime rolling deploys possible. The third would need a reason that does not currently exist.

**What would justify more than 2:** a single instance no longer absorbing full load during a failure (utilisation above roughly 70% at N=1), or a requirement to *retain redundancy through the loss of an availability zone* — two instances across two zones survive a zone failure but are left with no redundancy, so tolerating that with redundancy intact needs three across three zones. NFR3 asks for recovery under 60 seconds, not survival-with-redundancy through a zone loss, so this is not required today.

**A note on the number 5, because it turns out not to have been random.** At 10× MVP the load is ~1,500 req/s; against A8 with a 70% utilisation target that is 4.3 instances, so five. **Five was the right answer to the wrong question** — it sized the 10× case, not the MVP case. That is exactly what platform autoscaling is for, and the ceiling of 8 comfortably contains it.

### Why this shape

- The container **image is the reproducibility artifact** FR1 asks for: it builds identically anywhere, which is precisely what Phase 0's "works only where it was set up" failure lacked.
- Restart supervision is a property of the platform rather than something hand-rolled, which is what serves NFR3.
- **NFR4 is satisfied without a cluster.** Ten times the load is a maximum-instance number change plus a larger database, not a redesign. This was the original argument for Kubernetes, and a managed container platform delivers it without the operational surface.
- **Sizing to the target.** The worksheet requires components sized to the scale target. Kubernetes at 150 req/s was not that: it added a control plane, manifests, and a hand-provisioned cluster to serve a load two containers carry at a third of capacity. The earlier justification — "Maintainability ranks last in the QA ranking" — was misapplied, and is withdrawn: **the ranking breaks ties between attributes that are in conflict, and nothing here was in conflict.** Complexity that buys nothing is not a trade-off, it is a cost.
- Serverless functions were rejected as structurally hostile: no long-lived process, cold starts fighting NFR1, and an entire codebase that assumes a server. A managed *container* platform is the middle position — a real long-lived process, with the platform doing the supervision. Instances are configured always-warm with a minimum of 2, so no request pays a cold start.

**Rejected:** self-managed Kubernetes (over-sized for the target, and requires a hand-built cluster that undermines FR1); a bare VM (reproducibility becomes a documentation problem rather than a mechanism); PaaS such as Fly.io or Render (a legitimate answer, and simpler still — rejected only because the container image and its portability are worth keeping, and a managed container platform gives the same operational relief without giving up the artifact).

### Trade-offs and things given up

1. **The SPOF relocated rather than disappeared.** Two stateless instances do not remove the single point of failure — they move it downstream to the shared storage tier (Decision 2), and to the region if deployed to only one.
2. **Local debuggability is gone.** Phase 0 let you `cat docs/whatever.txt` to see the truth. Logs are now spread across instances and state lives behind a network hop.
3. **In-process state of any kind becomes illegal.** A Go mutex protects one instance, not two — so FR3's concurrency control *must* live in the storage tier. Likewise no in-process caching or session state. **This decision pre-constrains Decision 2.**
4. **Phase 0's local files are dead on arrival.** Two instances means two filesystems: a save lands on one, the next read hits the other, and the document appears to revert. Shared storage is now mandatory.
5. **Platform lock-in.** Cloud Run and Fargate are not interchangeable by configuration; moving between them means rewriting the service definition. The container image itself remains portable, which bounds the exit cost to the deployment layer — a deliberate, known limit rather than an accident.

**Not given up by this decision (recorded to prevent a mis-citation):** real-time collaboration is absent because Constraint 3 forbids WebSockets this phase, not because of the compute choice. Relatedly, one user's paste being clobbered by another's auto-save is the **lost-update problem** — that is FR3, the thing Phase 1 must fix, and it is not a SPOF.

---

## Decision 2 — Storage type and engine, and read-side caching

*Affects FR3, NFR1, NFR2, NFR4.*

- **Primary storage:** relational — **PostgreSQL**, single managed instance. One table: `documents(id, title, content, version, updated_at)`.
- **Concurrency control:** optimistic, via the `version` column. Save issues `UPDATE documents SET content=$1, version=version+1 WHERE id=$2 AND version=$3`; zero rows affected means someone else wrote first, and the user is told (HTTP 409) rather than silently overwritten.
- **Read-side caching:** **none at MVP**, with explicit re-entry triggers stated below.
- **Invalidation policy:** not applicable while there is no cache. When one is added, invalidate **on write**, not by TTL — a TTL would serve a stale document to someone actively editing it.
- **Driver use:** `database/sql` with `pgx`, called directly in route handlers. Constraint 1 declares this correct for this phase; the storage abstraction is Phase 3's job to earn.

### Why

- **FR3 is the deciding requirement.** It demands writes not be lost *silently* — a floor, not full consistency (Consistency ranks fourth of five). Optimistic version-checking is the cheapest mechanism that converts a silent clobber into a visible conflict, and it works across instances because the check is atomic inside the database. A shared file volume offers no such primitive.
- **NFR2 at MVP is comfortable:** 100 writes/sec of ~10 KB rows is roughly 1 MB/s, well inside a single modest instance.
- **NFR4 is satisfied by tuning, not replacement:** 1,000 writes/sec at 10× needs a larger instance and WAL tuning — allowed explicitly ("re-tuning is fine; replacing the storage tier is not").
- No cache at MVP is the Lean answer: at this scale a cache buys nothing measurable while adding an invalidation problem that FR3 makes strictly harder.

### Connection-pool sizing

This is where "stateless instances plus one database" designs actually fall over, so the numbers are stated rather than assumed.

PostgreSQL allocates **one backend process per connection**, so throughput peaks at a small multiple of core count and *degrades* beyond it — more connections do not mean more throughput. Vanilla `max_connections` defaults to **100** (managed providers scale it with instance memory; the real value must be read from the running instance, not assumed).

**The landmine:** Go's `database/sql` defaults `SetMaxOpenConns` to **unlimited**. Left unset, each instance opens connections without bound under load and exhausts the database for everyone. It must be set explicitly.

| | Instances | Pool per instance | Total connections | Against `max_connections` = 100 |
|---|---|---|---|---|
| MVP | 2 | 10 | **20** | 20% — comfortable |
| 10× MVP | 6 | 10 | **60** | 60% — fits |
| Autoscale ceiling | **8** | 10 | **80** | 80% — the hard limit |

Twenty connections are reserved for migrations, monitoring, administrative sessions and the superuser reservation, which is why the ceiling is 80 rather than 100. **This is what sets Decision 1's maximum instance count at 8** — the compute tier could scale further; the database is the binding constraint.

Also configured: `SetMaxIdleConns` equal to `SetMaxOpenConns` (the default of 2 causes constant connection churn), and `SetConnMaxLifetime` of 5 minutes so that pooled connections recycle — managed failover and proxies silently drop connections that a long-lived pool will otherwise keep handing out.

**Beyond 8 instances:** introduce **PgBouncer** in transaction-pooling mode, or the provider's built-in pooler. This is re-tuning rather than re-architecture, so NFR4 survives it.

### Read-latency budget (NFR1: p95 < 200 ms)

| Component | Expected |
|---|---|
| Platform load balancer / ingress | 1–2 ms |
| Handler and template render | 2–5 ms |
| Connection acquisition from pool | < 1 ms (unsaturated) |
| PostgreSQL primary-key lookup, ~10 KB row in shared buffers | 1–3 ms |
| Application ↔ database round trip (same region) | 0.5–1 ms |
| Serialisation and 10 KB response | 1–2 ms |
| **Server-side p50** | **~10 ms** |
| **Server-side p95** (GC pauses, occasional disk read, pool contention) | **~30–50 ms** |

**Budget: ≤ 50 ms server-side p95**, leaving roughly 150 ms for internet round trip and browser render — which is where most of a real user's 200 ms is actually spent. The requirement is therefore met with substantial margin, and the margin is the point: it is what allows the cache to be deferred honestly rather than optimistically.

**Cache re-entry triggers.** The deferred cache returns when any of these holds:
- server-side read p95 sustained above **100 ms** (half the budget consumed);
- database CPU sustained above **60%**;
- connection-pool wait time above zero at p95 (saturation);
- a hot-document access pattern emerges — the same documents read repeatedly, giving a cache something to actually hit.

**Rejected:** document store (the data is trivially relational and the conflict primitive is no better); shared file volume (no atomic compare-and-swap, so FR3 stays unsolved); key-value with persistence (durability weaker than the requirement warrants).

### Trade-offs and things given up

1. **Documents stop being plain files.** Inspecting state with `cat` and `grep` is gone, and with it the "no database to operate" simplicity.
2. **A single PostgreSQL instance is the new SPOF and the write ceiling.** No read replicas, no sharding, no multi-zone failover at MVP — its loss is the slowest recovery path in the system.
3. **Auto-save every 2 seconds against optimistic locking means two people editing one document generate constant conflicts.** That is *correct* behaviour under FR3 and Constraint 3 — the conflict surfaces instead of silently destroying work — but it is a visibly poor experience, and it is precisely the gap real-time collaboration closes in Phase 2.

---

## Decision 3 — State vs compute placement

*Affects FR2, NFR3.*

- **State lives** in the managed PostgreSQL instance, on its own volume with its own lifecycle, in a different failure domain from the application instances.
- **Compute runs** as stateless containers on the managed platform. An instance holds nothing durable — no local writes, no in-memory authority, no session state.
- **The boundary that makes state outlast compute** is the network plus the database's own managed storage: destroying every instance, or the whole service, loses nothing.
- **Recovery procedure:**
  - *Instance loss* — automatic; the platform starts a replacement in seconds. Satisfies NFR3.
  - *Whole-service loss* — redeploy from the committed image and service definition; minutes, bounded by Decision 4's procedure.
  - *Database loss* — restore from the managed automated backup. **Time-bounded but measured in minutes, not seconds.**

**Why:** FR2 asks for state to survive runtime recreation, which is exactly a demand that state and compute occupy separate lifecycles. Statelessness is also what makes Decision 1's interchangeable instances coherent at all, and what forces the concurrency control in Decision 2 into the database where it belongs.

**Trade-off stated honestly:** NFR3's 60-second clock is met for *compute* loss only. Database loss is the one path in this architecture that exceeds it, and closing that gap means replication with automated failover — deliberately out of scope at MVP scale, and the first thing to buy when it stops being acceptable.

---

## Decision 4 — Reproducibility mechanism

*Affects FR1, NFR5.*

- **Everything needed to reach a running system is committed to the repository:** the `Dockerfile`, the service definition (Cloud Run YAML or an ECS task definition), and a **Terraform** configuration covering the container service and the managed PostgreSQL instance.
- **Procedure:** clone → build image → push to the registry → `terraform apply` → verify `/healthz`. One documented command sequence, no console clicking.
- No configuration-management tool (Ansible, Chef, Puppet); there are no servers to configure.

**Why:** FR1 asks for a documented, reproducible procedure, and an image plus declarative infrastructure *is* that procedure — executable rather than prose, which is exactly the failure mode Phase 0 exhibited. NFR5's ten minutes is achievable because the only per-deploy work is an image build and an apply.

**The hole from the previous draft is closed, and it closed by getting smaller rather than by adding tooling.** The earlier design left a hand-provisioned Kubernetes cluster sitting outside the deployment clock, which meant "reproducible" carried an asterisk. Removing the cluster removed the thing that was too large to declare: a container service and a database instance are roughly fifty lines of Terraform, so the honest answer became cheap. Simplifying the compute shape and closing the reproducibility gap turned out to be the same action.

**What changes at 15 developers versus 1:** CI takes over image build and push on merge; per-developer preview environments appear (trivial on a platform that creates services on demand); secrets move into a managed secret store rather than living beside the configuration; and Terraform state moves to a shared remote backend with locking, because concurrent applies are the failure mode of infrastructure-as-code in a team. At one developer, every one of those is waste — which is the point of sizing to current scale.

**Trade-off:** Terraform is one more tool to learn and keep current, and its state file becomes something that itself must not be lost. That cost is accepted because the alternative is an undocumented environment, which is the specific problem this phase exists to solve.

---

## Decision 5 — Observability baseline

*Affects FR4, NFR3.*

- **Logging:** structured JSON to stdout, one object per request (method, path, status, duration_ms, document id where relevant). The platform collects stdout automatically; JSON means the logs are queryable later without re-instrumenting the application.
- **Health endpoint:** `GET /healthz` **verifies storage reachability** — it executes `SELECT 1` against PostgreSQL rather than merely proving the process is alive. Split into liveness (process up) and readiness (database reachable) so that the platform restarts a wedged instance but merely removes a database-blind one from rotation.
- **Metrics:** in-process counters exposed at `GET /metrics` in Prometheus text format — request rate, p50/p95 latency, error rate, **save-conflict count** (the direct FR3 signal: it shows whether users are actually colliding), and **connection-pool wait time** (the direct signal for Decision 2's ceiling, and the trigger for the deferred cache).

**Why:** FR4 requires health to be determinable without reading raw logs, which is exactly the difference between a status endpoint and a log stream. A liveness check that only proved the process was running would report "healthy" while every save failed — the specific failure FR4 exists to catch, hence the storage ping.

**Explicitly deferred to Phase 2:** third-party monitoring services (Sentry, Datadog), distributed tracing, alerting rules, and dashboards. Phase 1 builds the signals; Phase 2 decides who watches them.

---

## Decisions Made — digest

> **Format note:** reproduced in the five-field format with the `Considered` field included. Field names and order still need confirming against `project-overview.md`.

| Decision | Considered | Choice | Why | Trade-offs |
|---|---|---|---|---|
| **1. Compute** | Bare VM; self-managed Kubernetes; PaaS (Fly.io, Render); serverless functions | Container on a managed container platform (Cloud Run / ECS Fargate); **2 instances** at MVP, autoscaling to 8 | NFR4 — 10× is an instance-count change, with no cluster to operate. 2 is derived: one instance runs at 30% utilisation, so the second exists to survive losing one and to allow zero-downtime deploys | SPOF moves to the database; no local debuggability; in-process state now illegal; platform lock-in at the deployment layer |
| **2. Storage** | Document store; shared file volume; key-value with persistence | Single managed PostgreSQL; optimistic concurrency via `version` column; no cache, with stated re-entry triggers; pool of 10 per instance, ceiling 80 connections | FR3 — conflicts surface as 409 instead of silently clobbering. NFR1 met with margin (~50 ms server-side p95 against a 200 ms budget) | Plain-file inspectability gone; single-instance write ceiling; frequent visible conflicts when two people edit one document |
| **3. Placement** | State on an attached volume; sticky sessions with instance-local state | State in managed PostgreSQL; compute fully stateless | FR2 — state outlives any runtime; also what makes interchangeable instances coherent | Sub-60 s recovery is met for compute loss only; database loss is measured in minutes |
| **4. Reproducibility** | Shell scripts plus README; configuration management (Ansible); hand-provisioned infrastructure | Image + service definition + Terraform, all committed; one deploy command | FR1, NFR5 — executable procedure, under 10 minutes, nothing hand-built | Terraform is another tool, and its state becomes something that must not be lost |
| **5. Observability** | No metrics; plain-text logs; third-party APM from the start | JSON logs to stdout; `/healthz` pings the database; `/metrics` with save-conflict and pool-wait counters | FR4 — health determinable without reading logs; a process-only liveness check would report healthy while every save failed | External monitoring, tracing and alerting all deferred to Phase 2 |

---

## Still outstanding

Not open decisions — the five above are settled. These are gaps in the surrounding paperwork:

1. **`project-overview.md` is still not in the repository.** The mentor review confirms it specifies the five-field digest format; the digest above is a best reconstruction and the field names need checking against the real file.
2. **Assumption A3 (20% typing fraction)** remains the most load-bearing unverified number in this document; every write figure scales linearly with it.
3. **Cloud provider not chosen.** Cloud Run with Cloud SQL and ECS Fargate with RDS are equivalent for every argument made here, so the decision is deferred to deployment time rather than being architectural.
4. **Cloud deployment is unverified.** `deploy/main.tf` has never been applied and the container image has never been built — both need cloud credentials and a container runtime the development environment lacks. Everything below the deployment layer is verified; see the README's verification table.
