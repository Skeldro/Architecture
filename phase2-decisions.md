# Phase 2 — Architecture Decisions

Phase 2 adds the three off-the-shelf capabilities CollabDocs cannot launch without: **identity**, **real-time coordination**, and **operational visibility**. Each is a deliberate build-versus-buy decision, made against a committed market hypothesis and the project's quality-attribute ranking.

**Status:** Decisions 1, 2 and 3 settled 2026-08-27. Decisions 4–5 pending.

---

## Requirement labels (for citation)

**Functional**
- **FR1** — users can register, log in and log out; all document operations require valid credentials
- **FR2** — each user can access only their own documents (cross-user sharing is Phase 4)
- **FR3** — edits in one session propagate to other sessions viewing the same document in near real-time, without a page refresh
- **FR4** — application errors and operational signals are observable without reading raw container logs

**Non-functional**
- **NFR1** — login latency p95 < 200 ms at MVP scale
- **NFR2** — edit-to-peer-visible latency p95 < 500 ms
- **NFR3** — an error rate of ≥ 1% over one minute is detectable within five minutes
- **NFR4** — supports 10× MVP load without re-architecture (re-tuning is fine)
- **NFR5** — combined end-to-end availability meets the target for the chosen market

**Quality-attribute ranking (tie-breaks):** Performance > Resilience > Scalability > Consistency > Maintainability

**Constraints in force:** no abstraction around auth, real-time or monitoring libraries — direct use is correct this phase (Phase 3); no business-rule extraction from handlers (Phase 4); no use cases, DTOs or input validators (Phase 5); no multi-user sharing, ownership is one user per document (Phase 4).

---

## Decision 1 — Market hypothesis

*Drives Decisions 2, 4 and 5: market context shapes authentication feature needs, monitoring rigour, and the availability target.*

**Decision: product-led B2B — bottom-up adoption by small teams, sold self-serve.**

Pricing is flat per organisation up to a seat threshold, then per seat beyond it: a single predictable price for a small team, growing with the account as it grows.

### Why

**B2C is dead on arrival for this product.** The incumbent is Google Docs, it is free, and it has distribution no new entrant can buy. Competing at $0 against that requires hundreds of thousands of users before revenue is meaningful, which is not a bet a solo developer wins.

**B2G is not feasible with a legal team of zero.** Government procurement demands compliance artefacts, accessibility mandates, data-residency guarantees and tender processes that are staffing problems before they are engineering problems. The cost of entry is measured in people, not code.

**B2B is what remains, and the economics are the reason it works.** Documents products monetise on teams rather than individuals — Notion, Confluence and Coda all do — so a seat is worth $10–15 a month against a consumer's $0. Deal sizes land in the low thousands per year, which is enough for a small number of customers to matter.

**But specifically *product-led* B2B, because enterprise sales is also unstaffable.** Traditional B2B means demos, procurement and security reviews conducted by people this project does not have. Bottom-up adoption moves the buying decision inside the product: individuals sign up free, teams convert, and the sales cycle collapses from months to minutes. That preserves B2B economics without requiring a sales function, and it is how Slack, Notion and Figma were built.

**Compliance posture follows from that shape.** SOC 2 Type II and a data-processing agreement become necessary at the point where the first serious customer's security questionnaire arrives — not before. That is a real future obligation, and it is deliberately not a day-one one.

### The five dimensions, stated for the record

| Dimension | Position |
|---|---|
| **Price point** | Flat per organisation up to a seat threshold, per seat beyond it |
| **Deal size** | Low thousands of dollars per year per organisation |
| **Sales cycle** | Minutes. Self-serve signup; no salesperson exists |
| **Deployment shape** | Single multi-tenant hosted service; no on-premise, no per-customer instances |
| **Compliance posture** | GDPR and a DPA now; SOC 2 when the first customer requires it, not before |

### Explicitly not optimising for: enterprise procurement

No on-premise deployment, no bespoke contracts, no security-questionnaire desk, no per-customer infrastructure. **Every deal that requires those will be lost, and that is accepted** — the capability is not merely unbuilt, it is currently unstaffable.

**With one qualification that has architectural consequences.** This is a sacrifice of the present, not a permanent product position: if revenue arrives, B2G-style contracts become worth pursuing. The requirement that follows is *reversibility* rather than readiness — decisions today should not make that path impossible, but neither should they pay for it now.

Concretely, that means keeping the **exit cost bounded and known** rather than pursuing portability. Phase 1 already established the bound: the container image is portable, and lock-in is confined to the deployment layer. The live consequence is in **Decision 2** — a hosted-only identity provider cannot be deployed into an air-gapped or sovereign environment, whereas a self-hostable one can, so that choice should be made with the cost of reversal stated rather than discovered later.

The discipline here is the one the Phase 1 review already enforced: **do not build for a market you have not entered.** Record what would have to change; do not pay for it today.

### Consequences carried into the rest of this phase

1. **SSO/SAML becomes a tier gate rather than table stakes.** Traditional B2B would make it non-negotiable on day one; product-led B2B makes it the feature that unlocks the first large account. It is required *eventually*, which keeps Decision 2 genuinely open instead of pre-decided.
2. **The availability target is 99.9%, and the current architecture misses it.** Phase 1's components multiply to roughly 99.65% — about **30.6 hours of downtime a year against a budget of 8.76**, some 3.5× over. Decision 5 must therefore name architectural changes, not tuning.
3. **Monitoring rigour rises.** In a volume consumer business one broken user is noise; in B2B one broken *organisation* is a churn event, so signals need to attribute failures to a customer.

### Assumptions this decision introduces

| # | Assumption | Basis |
|---|---|---|
| B1 | **~20,000 registered users in year 1** | Inherited, not claimed: Phase 1's premise of 1,000 concurrent users at MVP with roughly 20 registered per concurrent user. The 100,000-concurrent year-one goal came with the brief |
| B2 | **~15 seats per organisation on average** | Small-team focus; implies roughly 1,300 organisations at B1 |
| B3 | **Pricing around $99/month up to 10 seats, then ~$9/seat** | Placeholder figures consistent with the agreed shape. The *shape* is decided; the exact numbers are not, and only Decision 2's cost comparison depends on them |
| B4 | **~15,000 monthly active users in year 1** (band 14,000–16,000) | Derived backwards from the inherited concurrency figure rather than forecast — see below |
| B5 | **Year-1 peak concurrent = 10,000** | NFR4 requires 10× MVP headroom, and MVP is 1,000 concurrent (A1). The brief's 100,000-concurrent goal is a stated ambition, not a requirement of this phase, and reaching it would need re-architecture rather than re-tuning |



### Deriving monthly active users (B4)

**MAU** (monthly active users) is the unit identity vendors bill on, so Decision 2's cost comparison rests entirely on it. No forecast exists for this product, and a top-down market estimate would be unfalsifiable — so the figure is derived backwards from the one number the brief already fixed: concurrency.

The chain, using the standard ratios:

```
concurrent / DAU  =  minutes of use per day  ÷  minutes in the active window
                  =  75 min ÷ 600 min (a ~10-hour spread of working timezones)
                  ≈  12.5%

DAU  =  1,000 concurrent ÷ 0.125          =  8,000        (A1)
MAU  =  8,000 DAU ÷ 0.50 stickiness       =  16,000       (50% DAU/MAU is mid-band for a B2B work tool)
     →  16,000 / 20,000 registered        =  80% active   (B1)
```

**Cross-check against A6.** Phase 1 assumed 20 registered users per concurrent user, i.e. concurrent/registered = 5%. Multiplying the chain independently gives `12.5% × 50% × 80% = 5.0%`. The two agree, which means A6 and the standard ratios are describing the same product rather than contradicting one another.

**Sensitivity.** The only soft input is daily usage. At 84 minutes per day the chain yields 14,000 MAU; at 75 minutes, 16,000. The answer is therefore stable in the 14,000–16,000 band, and **no decision below turns on where in that band it falls** — which is the reason the estimate is good enough to use.

**Falsification.** An assumption that contradicts the others is refuted for free. A 20% active rate would imply 4,000 MAU, 2,000 DAU, and every daily user online for five hours — absurd, and rejected without any market research. Real numbers replace this the moment the product has users; until then the derivation stands in for a measurement, and is labelled as such.

**A tension recorded honestly:** B1 combined with B3 implies revenue in the millions within a year, which no solo developer reaches without exceptional growth. The scale figure is inherited from the brief rather than forecast here, and product-led growth is the only distribution model that makes it even arguable. Where the two conflict, the scale premise governs the architecture and the revenue figure should be read as illustrative.

---

## Decision 2 — Authentication

*Affects FR1, NFR1, and Decision 1.*

**Decision: WorkOS.** Hosted identity, integrated directly in the route handlers and one session middleware, with no abstraction layer over it (Constraint 1).

### One-year cost at 15,000 MAU (B4)

Integration effort is included, because leaving it out is the classic way a build-versus-buy comparison lies. Custom's eight hours *is* its integration; the vendors need roughly a week each.

| | Licence, year 1 | Integration | SSO when demanded | **Year 1** | **Total once SSO exists** |
|---|---|---|---|---|---|
| **Auth0** | $4,500 (15,000 × $0.025) | ~$4,000 | +$2,880/yr add-on | **$8,500** | **$11,380** |
| **WorkOS** | **$0** (free to 1M users for SSO/SAML/Directory Sync) | ~$4,000 | included | **$4,000** | **$4,000** |
| **Custom JWT + bcrypt** | $0 | $800 (8 hr) | +$12,000–16,000 to build | **$800** | **$12,800–16,800** |

**Auth0 is dominated and eliminated first.** It costs more than WorkOS on every line and buys nothing WorkOS lacks at this scale — not portability, not features, not reversibility.

**The remaining comparison has a shape worth stating plainly: custom is the cheapest option right up until the moment the market hypothesis pays off.** Without SSO it wins on cash by a wide margin. With SSO — which Decision 1 says arrives with the first serious customer — it becomes the most expensive option available, by a factor of three.

### Why cost is not the discriminator

The three options span an order of magnitude, and **the cheapest is a vendor while the middle one is building it yourself**. That inverts the usual build-versus-buy intuition and it means price has stopped separating the candidates. Three other factors decide it.

**1. Decision 1 makes SSO a near-term obligation.** Product-led B2B means the enterprise gate arrives on a customer's schedule, not ours. WorkOS supplies it years before it could be justified as a build. This is the deciding argument.

**2. Reversibility is real but hypothetical.** Decision 1 asked that B2G-style contracts stay reachable, and WorkOS — being hosted-only — cannot deploy into an air-gapped or sovereign environment. But that is *possible future revenue* weighed against *near-term actual revenue*, and paying a certain cost today for an uncertain benefit later is the exact pattern the Phase 1 review already corrected once. Reversibility is therefore preserved by cheap measures rather than by the implementation choice (below).

**3. Security ownership.** The worksheet's own phrase for custom is "security maintenance is yours", and that term never shrinks. Eight hours buys password hashing and token issue/verify — not password reset, not MFA, not session revocation, not breach-list checking, not account lockout, not audit logs.

### NFR1: login p95 < 200 ms constrains both options

Neither candidate satisfies the latency budget for free, in opposite directions:

- **Hosted providers** using a redirect flow cross the vendor's domain over several round trips, commonly 200–500 ms end to end. WorkOS must therefore be integrated in its **embedded/API mode rather than a hosted redirect**, which is a binding implementation requirement rather than a preference.
- **Custom bcrypt is deliberately slow** — that is its purpose. A work factor of 10 costs roughly 60–100 ms; 12 costs 250–400 ms and breaches the budget outright. Choosing custom would mean **capping the work factor for latency reasons**, trading offline-cracking resistance for response time, and defending where that line sits.

Verification is by measurement at the endpoint, not by assumption.

### Considered and rejected: an adapter layer with swappable auth implementations

The most attractive rejected option, recorded because the reasoning matters more than the conclusion: wrap authentication behind an internal interface, run WorkOS today, and deploy a self-hosted implementation for B2G later.

It was rejected for three independent reasons, any one of which is sufficient.

**It is forbidden this phase.** Constraint 1 bans abstraction over auth, real-time and monitoring libraries, deferring it to Phase 3 — the same treatment Phase 1 gave storage.

**An abstraction designed against one implementation is a guess.** The interface would be shaped by WorkOS's session model, token format, callback flow and user object, and the self-hosted alternative would not fit it. The wrapper leaks, the swap costs what it always would have, and a wrapper that misrepresented itself as ready is now also maintained. **The second implementation is what discovers the interface**, which is precisely why Phase 3 follows Phase 2 rather than preceding it.

**Two implementations is a fork, not a swap.** Maintaining parallel auth builds means every feature built twice, tested twice, and a permanent class of divergence bug — a large standing commitment made today against revenue that does not exist. The realistic path is not a build flag but a funded migration: win the contract, spend two to three weeks moving identity. Pricing that migration is the correct posture, and it matches Phase 1's conclusion about deployment lock-in — a bounded, known exit cost rather than portability purchased up front.

### Reversibility measures taken instead

Free, and none of them an abstraction, so Constraint 1 holds:

1. **Vendor identifiers never become primary keys.** The application keeps its own `users.id`; WorkOS's subject identifier is an ordinary attribute beside it, and `documents.owner_id` references ours. Changing provider then repopulates one column instead of rewriting every foreign key. This is the highest-value item on the list and it is a data-model choice, not a layer.
2. **Session validation lives in one middleware**, not scattered through handlers — which is how it would be written regardless.
3. **OIDC rather than a proprietary SDK flow**, so the protocol survives a change of vendor even though the vendor does not.

### Things given up

1. **The air-gapped deployment path, until a migration is funded.** B2G and sovereign contracts stay closed. Reopening them is a two-to-three-week identity migration — priced, not impossible, and explicitly not paid for now.
2. **Control of the pricing floor.** Free-to-1M is a land-grab term, and land-grab pricing gets revised. If it changes, the negotiation happens from inside a dependency every user authenticates through.

### Revisit if

- WorkOS reprices or restricts the free tier, at which point the migration cost above becomes a live number rather than a hypothetical one.
- A B2G or sovereign-deployment contract is actually won and funded.
- Measured login p95 exceeds 200 ms even in embedded mode, which would make the vendor a direct NFR1 failure rather than a cost question.

---

## Decision 3 — Real-time mechanism and conflict semantics

*Affects FR3, NFR2, NFR4.*

**Decision: a self-hosted WebSocket endpoint inside the Go application**, using a maintained Go library (`coder/websocket` or `gorilla/websocket`), with PostgreSQL `LISTEN`/`NOTIFY` as the cross-instance backplane. The server relays opaque operations; it does not apply them. Optimistic concurrency on the `version` column is retained as the persistence authority.

### Mechanism

**Rejected: Pusher / Ably — on fit, not on price.** These are **fan-out** products: one publish, many deliveries, which is what makes live scores, chat rooms and dashboards worth paying for. CollabDocs has a fan-out ratio of **1** — Constraint 4 gives each document a single user, and even after Phase 4 introduces sharing a document has a handful of editors, never thousands. More decisively, for a collaborative editor **the transport is the easy part**: a managed pipe delivers bytes and leaves ordering, convergence and persistence entirely with us. Buying it would solve perhaps a fifth of the problem while adding a fourth multiplicand to Decision 5's availability product.

*(Noted for the record: the managed category that would actually address the hard part is sync engines — Liveblocks, Yjs with a hosted backend, Automerge, ShareDB — not pub/sub. None appear in the options offered, and the mature implementations are JavaScript and Rust, which our stack argues against.)*

**Rejected: Socket.IO — because it is a JavaScript library and this is a Go service.** There is no first-party Go server; adopting it means depending on a community reimplementation of a protocol whose reference implementation lives in another language and versions independently. The failure mode is a routine client upgrade breaking the handshake, with the fix outside our control. Its genuinely useful features are small in Go: reconnection is about thirty lines of client script, rooms are a map, and heartbeats ship with the library. Only long-polling fallback is non-trivial, and its value has declined as WSS became universally proxy-passable.

**Chosen: raw WebSocket.** First-class in Go, actively maintained, and natively supported by the browser at zero client-bundle cost — which matters for an application that currently ships no JavaScript at all. The worksheet prices this at ~40 hours because it assumes message replay after disconnect; **we do not need replay.** On reconnect the client simply re-fetches the document over the existing HTTP endpoint, which removes sequence numbers, gap detection and buffering, and brings the realistic build to one or two days.

**Cross-instance delivery** uses PostgreSQL `LISTEN`/`NOTIFY` rather than adding Redis. Instances are stateless in every respect that matters (Phase 1, Decision 3): a connection is held by one instance, and a change on another reaches it through the database we already run. Payloads are capped at 8000 bytes, which is ample for "document N changed to version M", and durability is not required because the document itself is durable.

### Conflict semantics

**Retained: optimistic concurrency via the `version` column, with the server as relay rather than authority.**

**LWW is explicitly rejected as a regression.** The worksheet lists it as the simplest option, but the system already does better: the version predicate *detects* a collision and returns the losing writer's text to them. Last-write-wins would silently discard it, undoing the FR3 property built and verified in Phase 1.

**OT and CRDT are deferred, because there is nothing yet to converge.** Constraint 4 permits one user per document, so the only concurrency available is one person with two tabs or a second device. Both are weeks of work for a scenario that cannot presently occur.

**What real-time changes is the role of the version column, not its existence.** The broadcast carries the new version alongside the operation, so a peer that receives it updates its baseline and is no longer stale. Real-time makes conflicts *rare*; the version check catches those that slip through — a dropped broadcast, a sleeping tab, a device on a dead connection — and the existing conflict page handles them. **Optimistic concurrency stops being the primary mechanism and becomes the safety net.** Nothing built in Phase 1 is retired.

Keeping the server a **relay** rather than an authority preserves Phase 1's Decision 3: the real-time layer holds no state worth keeping, so dropping every connection loses nothing, and PostgreSQL remains the single source of truth.

### Save cadence

**Broadcast immediately and independently; persist on a bounded, edit-triggered schedule.** Propagation and persistence are different problems with different deadlines: peers need the edit inside NFR2's 500 ms, whereas the database only needs the document eventually.

Persistence is **triggered by changes, never by a clock.** A clean document runs no timers, counts nothing and issues no writes — there is nothing to save, so nothing happens. The schedule exists only while unsaved changes do.

```
CLEAN ──first change──► DIRTY
                          │  start inactivity timer (2 s)
                          │  start ceiling timer   (5 s)   ← not reset by later changes
                          │  start byte counter
                          │
        further change ───┤  reset inactivity timer; accumulate bytes
                          │
        persist when ─────┤  2 s of inactivity
                          │  OR 5 s since the first unsaved change
                          │  OR 1 KB of accumulated change
                          │  OR the page is unloaded
                          ▼
                       CLEAN   cancel timers, reset counter
```

**The one implementation trap, stated because it is easy to get wrong:** the inactivity timer resets on every change, and the **ceiling timer must not**. Reset both and continuous typing never reaches the ceiling, which collapses the design back to pure debounce and restores the unbounded loss window the ceiling exists to close.

**Dirty means the content differs from the last persisted version, not that events occurred.** Typing a character and deleting it again leaves the document byte-identical to what is stored, so the flush is skipped. The comparison happens once at flush time rather than per keystroke.

**Why 1 KB rather than a smaller threshold.** The volume trigger exists to catch bulk changes — a paste, a large deletion — where waiting out the ceiling would leave a lot of work unsaved. It must therefore sit above what ordinary typing produces within one ceiling interval. At 40 words per minute a typist generates roughly 4 characters per second, and a fast typist about 8, so five seconds of continuous typing is at most ~40 characters. A threshold of 20 bytes would fire at 2.5–5 seconds — either duplicating the ceiling exactly or beating it and doubling the write rate — while contributing no additional safety. At 1 KB it fires only for changes that are not typing.

**Data-loss exposure, stated as a choice rather than left implicit.** Because Decision 3 makes the server a relay that never assembles the document, it cannot flush on a client's behalf when a connection drops. The worst case is therefore five seconds of one user's own typing, roughly 20–40 characters, and the page-unload flush removes the most common instance of it (closing a tab mid-sentence). This is more permissive than Google Docs, which persists far more aggressively, and it is accepted deliberately in exchange for the write reduction below.

**The write-load result.** Phase 1's assumption A4 saved every two seconds regardless, giving 100 writes/sec at MVP and 1,000 at 10×. A five-second ceiling bounds this at **40 writes/sec at MVP and 400 at 10×** — a 60% reduction, and a genuine upper bound rather than an average, since pauses persist earlier and then stop. **Adding real-time makes the storage tier cheaper**, which is the write-amplification lever Phase 1 identified and deferred.

### Channel authentication

**Session cookie on the handshake**, validated before the connection is accepted. The WebSocket upgrade is an ordinary HTTP request, so the browser attaches the WorkOS session cookie automatically. This is forced as much as chosen: **the browser WebSocket API cannot set custom headers**, so `Authorization: Bearer` is unavailable, and a token in the query string would be written to every access log along the path.

Two checks happen at connection time: the session must be valid (acceptance criterion 3), and the subscriber must own the requested document (FR2, acceptance criterion 4). Subscriptions are per document, so broadcasts never reach sessions viewing anything else (acceptance criterion 6).

### Scaling to 10× MVP (sketch only — not implemented)

At B5's 10,000 concurrent connections the current envelope **fails, and the arithmetic is worth stating**: Cloud Run permits at most 1,000 concurrent requests per instance and a held-open WebSocket counts as one for its entire life, so 10,000 connections need at least ten instances — while Decision 1 caps at eight, a ceiling set by database connections rather than by compute.

Three moves, in increasing order of change. **Re-tune first:** reduce the per-instance pool from ten connections to five, which doubles the permissible instance count to sixteen and carries 16,000 connections — re-tuning, so NFR4 survives. **Then decouple:** introduce PgBouncer so instance count stops being a function of database connections at all. **Then separate the tiers:** move connection-holding into a dedicated real-time service so that the request-serving tier scales on request concurrency and the connection tier scales on socket count, since the two signals stop agreeing the moment connections are long-lived.

A second, unrelated consequence: **Cloud Run terminates connections at sixty minutes**, so reconnect-and-refetch is mandatory rather than optional. That is already the design, which is convenient rather than accidental.

### The Phase 4 target architecture, recorded now

The mechanism for genuine multi-editor concurrency is settled in *shape* even though it is not built: **server-authoritative, clients transmit operations rather than documents, clients predict locally and hold a buffer of unacknowledged operations, and divergence is resolved by rolling back to server state and replaying the buffer.**

This is Quake 3's network model — usercmds, authoritative snapshots, client prediction, replay of unacknowledged input — and recognising the two problems as the same one is what makes the target concrete rather than aspirational. An OT client is structurally identical: a pending-operation buffer, rewound and re-applied when authoritative state arrives.

**One named open question decides which family it lands in.** Replay is only sound when operations are context-independent. Game inputs are — "move forward" means the same thing against any baseline. Text operations addressed by position are not: `insert(6, …)` refers to an index in a mutable array, so replaying it against a changed baseline corrupts the document. Two resolutions exist, and they are exactly the two options the worksheet offers:

- **Positional operations**, corrected before replay by a transform function — this is **OT**, and that transform is where its difficulty lives.
- **Identity-addressed operations** (`insert after character (site, counter)`), which are context-independent and replay natively with no transform — this is **CRDT**, paid for in per-character metadata and tombstones.

Deferred to Phase 4 deliberately: with zero concurrent editors and zero users, choosing now would mean designing against no implementations — the same error as wrapping authentication in an adapter before a second provider exists.

**Candidate carried forward:** bounding position arithmetic to document sections rather than the whole document, so that merge scope and blast radius stay small. Production editors do this (Notion syncs per block; Google Docs chunks internally). The caveat is that a section anchor must itself be stable — paragraph indices are not, since inserting a newline renumbers everything after it.

**Rejected outright, and worth recording because the reasoning generalises:** resolving concurrent edits inside a fixed time window — a "human reflex buffer" of a few hundred milliseconds, after which corrections are applied by offset. **Correctness may never depend on a wall-clock window in an asynchronous network**, because message delay is unbounded: a device on a poor connection delivers its operation seconds late, the window has closed, and the document is silently wrong — corruption rather than a detected conflict, which is precisely the FR3 failure. Timeouts are legitimate for liveness decisions ("assume it is gone and proceed") and never for correctness ones. The offset correction also generalises, on contact with multi-character deletes, pastes and out-of-order arrival, into a general position-remapping function — which is `transform()`, rebuilt without OT's correctness proofs, in a field where several published algorithms were later shown to be wrong.

### Things given up

1. **Managed availability for the real-time layer.** Self-hosting means an outage in the connection layer is ours to diagnose and fix, with no vendor SLA behind it. Accepted because it is also what *avoids* a fourth multiplicand in Decision 5 — the only choice in this phase that improves end-to-end availability rather than degrading it.
2. **Long-polling fallback.** Clients behind a proxy that blocks WebSocket get no real-time updates at all. They degrade to the Phase 1 experience — refresh to see changes — rather than failing, but in a B2B market corporate proxies are exactly where this risk lives.

### Revisit if

- Concurrent connections approach roughly 8,000, where the Cloud Run instance ceiling binds and the re-tuning sequence above must actually be executed.
- Phase 4 introduces multi-editor documents — at which point the positional-versus-identity question stops being deferrable.
- Measured edit-to-peer latency exceeds NFR2's 500 ms p95, which would indict `LISTEN`/`NOTIFY` as the backplane before it indicts the transport.
- Corporate-proxy WebSocket blocking is reported by real customers, which would revive the long-polling question.
