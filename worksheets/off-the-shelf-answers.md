# Off-the-Shelf Worksheet — Answers

Reference answers to the commercial-services worksheet. Each question is reproduced, then answered. Written for discussion in mentor sessions.

Acronyms are expanded on first use throughout.

---

## Question 1 — Service Models & Managed Services

> - Explain the difference between IaaS, PaaS, and SaaS with examples.
> - When would you choose PaaS over IaaS?

### a. The three models

All three are rungs on a single ladder: **how much of the stack somebody else operates.** The stack runs `hardware → virtualisation → operating system → runtime → application code → data`, and each rung hands another slice to a vendor.

| Model | Vendor operates | You operate | Examples |
|---|---|---|---|
| **IaaS** (Infrastructure as a Service) | hardware, virtualisation | operating system upward | EC2, Google Compute Engine, DigitalOcean droplets |
| **PaaS** (Platform as a Service) | up through runtime and scaling | your code and data | Heroku, Google App Engine, Fly.io, Render, Elastic Beanstalk |
| **SaaS** (Software as a Service) | the entire application | your data only | Gmail, Salesforce, Auth0, Stripe, Forter |

**IaaS** hands over a machine with root access. Everything above the kernel — patching, runtime, deployment, scaling — remains the customer's problem.

**PaaS** removes the machine entirely. Code is pushed; the platform builds, runs, restarts and scales it. The useful test is whether there is anything to log into: if you can open a shell on it, it is infrastructure and its contents are yours to maintain; if there is only a deployed application, it is a platform.

**SaaS** makes the customer a user rather than an operator. Nothing is deployed.

An important qualification: *as a Service* means **the vendor operates it**. Software that is licensed for self-installation is a product, not a service, even when the same company sells a hosted version of it — running it yourself means taking on the operations that the service model exists to remove.

### b. Managed Services — the fourth term, which is not a rung

**Managed Services** are a modifier applied to a single component rather than a level of the ladder. **RDS** (Relational Database Service) is managed PostgreSQL; **GKE** (Google Kubernetes Engine) is managed Kubernetes; **MSK** (Managed Streaming for Kafka) is managed Kafka.

They sit between IaaS and PaaS: the engine, version, instance size and schema remain the customer's choices, while patching, backup, replication and failover become the vendor's. The knobs are kept; the pager is not.

### c. When to choose PaaS over IaaS

The governing question is whether **running servers is part of what makes the product good**. If it is not, the rung should be bought. PaaS is therefore the correct default, and the burden of proof lies with the decision to drop down to infrastructure.

Descend to IaaS when one of these applies:

- **Capability wall.** The platform simply will not run the workload — **GPUs** (Graphics Processing Units), kernel tuning, raw sockets, unusual protocols, very long-running jobs. This is a matter of impossibility rather than expense.
- **Cost at scale.** Platform margin becomes material once volume is large enough.
- **Compliance or data residency.** Regulation dictates where and how the workload runs, in ways the platform cannot guarantee.
- **Lock-in.** Leaving a platform means rebuilding the deployment story from nothing.

Each rung upward **buys leverage and spends control**: more is achieved per engineer, and fewer things remain adjustable.

---

## Question 2 — Build vs Buy Decisions

> - When should you use commercial products versus building yourself? (List 4-5 criteria)
> - What are the hidden costs of commercial services? (Name 3-4)
> - How do you calculate TCO (Total Cost of Ownership) for build vs buy?

### a. When to use commercial products instead of building

**1. Is it your differentiator?**
The primary test, and the one that decides most cases on its own. Build what makes the product good; buy everything else. This is the layer model applied to spending: value concentrates in the Domain and Generic Infrastructure layers, so engineering hours belong there and money belongs in the Off-the-Shelf layer. Nobody chooses a documents application because its password-reset flow is exceptional — so that is not where four weeks go.

**2. Does the product absorb a legal or compliance burden?**
Buying often purchases an answer to a regulatory requirement rather than a feature. Payment processing carries **PCI DSS** (Payment Card Industry Data Security Standard); authentication vendors carry **SOC 2** (Service Organization Control 2) and **GDPR** (General Data Protection Regulation) tooling. The vendor's audit report answers a customer's security questionnaire; a hand-rolled equivalent means defending one's own implementation. That shortens sales cycles, not merely development.

**3. What does it genuinely cost to build?**
Developer time at a loaded rate, the salaries behind it, and the **opportunity cost** of what those people are not building instead. For a solo developer, opportunity cost dominates every other term: four weeks spent on authentication is four weeks in which the actual product does not exist.

**4. Would the in-house version actually be better?**
An honest expertise check. A commercial service has absorbed years of edge cases, attacks and production incidents that a new implementation has not seen. Where accumulated hardening matters — security, payments, email deliverability — the ready-made product usually wins on quality, not merely on time.

**5. What does the alternative actually cost?**
"Buy" is not one thing. Free and open-source components (PostgreSQL, Redis) cost operations rather than licence fees; subscription services cost money that recurs and grows. The comparison differs in each case, and conflating them produces bad decisions in both directions.

### b. Hidden costs of commercial services

The subscription is not hidden — it appears on the invoice. These are the costs that do not:

**1. Lock-in and switching cost.** Careless integration spreads the vendor's model through the codebase: its identifiers in the tables, its assumptions in the control flow. Leaving then becomes a migration rather than a swap. The cost is invisible on the day of signing and enormous on the day of departure.

**2. Integration cost.** Wiring the service in, learning its model, working around its edge cases, and maintaining that glue. Always non-zero, almost never budgeted, and paid again at every major version.

**3. Pricing that scales with success.** Per-user and per-request billing ties the cost curve to the growth curve. A service that is trivially cheap at MVP scale can be punishing at 10× or 100× — and by the time it hurts, cost 1 has made leaving expensive. This is the hidden cost with real consequences.

**4. Their outage is your outage.** The vendor's **SLA** (Service Level Agreement) and their incidents are inherited wholesale, with no ability to intervene. Users experience the outage as yours. This is the bill for having traded control for leverage.

**5. The deprecation treadmill.** Vendors retire **API** (Application Programming Interface) versions on their own schedule, producing forced migrations at times not of the customer's choosing.

### c. Calculating TCO for build vs buy

**TCO** means totalling both options over the same horizon, because the naive comparison — development cost against monthly fee — is wrong on both sides. Building is also maintaining; renting is also expanding.

**Horizon: 3 years.** Beyond that, the technology, the vendor landscape and the requirements have all changed enough to make the figure fiction.

**Build column**
- Initial development: `hours × loaded hourly rate`
- Maintenance: 10–20% of initial development *per year, indefinitely*
- Infrastructure to run and monitor it
- Opportunity cost of the work displaced

**Buy column**
- Subscription × horizon, **modelled at projected usage**, not today's
- One-time integration development
- Per-transaction fees at projected volume

Both sides are then risk-adjusted: the expected cost of an outage or breach under each option, and the exit cost should the decision prove wrong.

**Worked example — authentication.** Assumptions: $100/hour, four weeks to build (160 hours), 10% annual maintenance, $200/month for the service, one week of integration work.

| | Build | Buy |
|---|---|---|
| Up-front | $16,000 | $4,000 (integration) |
| Per year | $1,600 (maintenance) | $2,400 (subscription) |
| **Year 1** | **$17,600** | **$6,400** |
| **3-year total** | **$20,800** | **$11,200** |

Buying wins decisively and crosses over only around year seven — by which point the built version requires work the estimate never included: **MFA** (Multi-Factor Authentication), **SSO** (Single Sign-On) and continuous security patching.

**The term that flips it** is per-user pricing. Growth that takes a subscription from $200 to $2,000 per month makes buying cost $24,000 a year and reverses the arithmetic entirely. The decision is therefore not a single sum but a sum evaluated at *the scale expected to be reached* — precisely the term that gets forgotten.

---

## Question 3 — Authentication Services

> - Why is authentication complex to build correctly? (Give 3-4 reasons)
> - What compliance benefits do commercial auth services provide?
> - What's the trade-off between Auth0/WorkOS ($200/month) and building your own (4 weeks development)?
> - Authorization (who-can-do-what) can also be bought as a service — OpenFGA, SpiceDB, Auth0 FGA — instead of built. What's the build-vs-buy trade-off?

### a. Why authentication is complex to build correctly

**1. The surface area is enormous, and none of it is the product.**
Signup, login, logout, password reset, email verification, session management, remember-me, account lockout, rate limiting, MFA enrollment *and recovery*, social-login callbacks, SSO, device management, audit logs. Each is a separate flow with its own way of being subtly wrong, and not one of them is a reason anybody chooses the product.

**2. The failure mode is catastrophic and silent.**
An ordinary bug loses a paragraph and someone complains. An authentication bug exposes every user's data, produces no crash and no stack trace, and is typically discovered by an outsider. The severity class differs from ordinary defects: the exposure is legal, not merely operational.

**3. Correctness is a moving target.**
Password hashing parameters, session handling, **CSRF** (Cross-Site Request Forgery) defence, credential-stuffing mitigation and breach-list checking all shift over time. What was correct practice five years ago is weak now, and tracking that drift is a permanent maintenance obligation rather than a one-off implementation cost.

**4. It depends on infrastructure outside the application.**
Verification and password-reset messages have to actually arrive, which requires domain reputation and correct **SPF** (Sender Policy Framework), **DKIM** (DomainKeys Identified Mail) and **DMARC** (Domain-based Message Authentication, Reporting and Conformance) configuration. Misconfigured, reset emails land in spam and authentication is silently broken for real users. Federated login adds a second such dependency: being a registered relying party with each identity provider.

### b. Compliance benefits of commercial authentication services

Vendors carry attestations that would otherwise have to be earned: SOC 2 Type II, ISO 27001, GDPR tooling for data-subject access and deletion requests, and **HIPAA** (Health Insurance Portability and Accountability Act) alignment where health data is involved. A **DPA** (Data Processing Agreement) places part of the breach liability on the vendor contractually.

The mechanism matters more than the list: **their audit evidence becomes usable as the answer.** When a customer's security questionnaire asks how credentials are stored, the response is an attestation rather than a defence of one's own cryptography. The benefit is therefore commercial as much as legal — it unblocks enterprise sales. Vendors also track regulatory change continuously, which is ongoing work rather than a one-time cost.

### c. Auth0 / WorkOS at $200 per month versus four weeks of building

The financial comparison is Question 2's arithmetic: roughly $20,800 over three years to build against $11,200 to buy, crossing over only around year seven — and that crossover assumes the built version never grows MFA, SSO or continued security patching, which it must.

The decisive term is not financial. Building means **assuming the legal exposure personally**: the breach, the compliance gap and the questionnaire all become the builder's to answer. Buying transfers a meaningful portion of that.

Two conditions reverse the conclusion. **Per-user pricing at scale** — growth taking the subscription from $200 to $2,000 per month makes buying cost $24,000 a year and flips the sum. And **building genuinely wins** when authentication is itself the product, when requirements exist that no vendor supports, or when data-residency rules forbid the vendor outright.

### d. Externalized authorization — the build-vs-buy trade-off

Authorization is a *different* decision from authentication, despite appearing to be the same shape.

**The options are not all subscriptions.** OpenFGA is open source (a **CNCF** — Cloud Native Computing Foundation — project, originating at Auth0/Okta) and self-hostable at no licence cost. SpiceDB is likewise open source, with an optional managed tier. Auth0 FGA is the hosted product. "Buy" here therefore spans free-but-operated through to fully managed, and the free-versus-subscription distinction from Question 2 applies inside this single category.

**What these tools solve** is the permission *graph*: may this user edit this document, given membership of a team that was granted access to a parent folder? Google published the Zanzibar paper describing its own solution — built for Google Docs — and OpenFGA and SpiceDB are open implementations of that model. This is **ReBAC** (Relationship-Based Access Control), as opposed to simpler **RBAC** (Role-Based Access Control).

**Four reasons the calculus differs from authentication:**

1. **Authorization sits on the hot path.** Authentication happens once per session; authorization is evaluated on every request, often repeatedly. Externalizing it adds a network hop to the most frequently executed path in the system.
2. **The rules are domain knowledge.** Who may edit a document is a business rule — Domain Layer material — whereas authentication is commodity infrastructure. Applying the differentiator test therefore yields *different answers* for the two.
3. **There is a complexity threshold.** Owner/editor/viewer is a column and a `WHERE` clause, and that is the correct answer for most systems. Zanzibar-class tooling earns its place once the graph deepens — nested groups, inherited folder permissions, share links — or once several services must agree on the same rules.
4. **Failure coupling is more severe.** An unavailable authorization service denies everything, which is a harder outage than an unavailable login service.

**Position taken:** buy authentication, build authorization — until the permission graph outgrows a table. For CollabDocs specifically, sharing with people, teams and links is the exact use case Zanzibar was designed for, so the threshold is genuinely reachable; the plan is a `permissions` table in phase 2, with ReBAC tooling reconsidered when nested or inherited sharing arrives.

---

## Question 4 — Database Services

> - What does a managed database (RDS) handle that you'd need to manage on EC2?
> - When would you NOT use a managed database?

### a. What RDS handles that self-management on EC2 would not

Running PostgreSQL on a bare instance means owning every one of the following:

- **Provisioning and installation** — instance sizing, storage type, engine installation and initial configuration.
- **Patching** — operating-system updates plus database engine security patches, and minor and major version upgrades with their rehearsal and rollback plans.
- **Backups and point-in-time recovery** — scheduling, retention, off-instance storage, and *testing that restores actually work*, which is the step most self-managed setups skip until the day it matters.
- **Replication and failover** — configuring read replicas, maintaining a standby, detecting primary failure and promoting the standby automatically. Managed multi-zone deployments make this a checkbox; by hand it is a project.
- **Storage management** — growing volumes before they fill, provisioning **IOPS** (Input/Output Operations Per Second), and monitoring for exhaustion.
- **Security plumbing** — encryption at rest and in transit, key management and rotation, network isolation, and identity integration.
- **Monitoring** — metrics collection, slow-query surfacing and performance dashboards.
- **Maintenance windows** — a defined, announced period for disruptive work rather than ad-hoc downtime.

What remains the customer's responsibility even under RDS is worth stating explicitly, because it is routinely misunderstood: **schema design, indexing, query performance, connection pooling, capacity choice, cost control and migrations.** Managed means the database stays up; it does not mean the database is fast. A missing index is still entirely the developer's problem.

### b. When not to use a managed database

- **Superuser or operating-system access is required.** Managed engines permit only allowlisted extensions and no filesystem access. Custom or compiled extensions, unusual storage tuning or direct log access rule managed services out.
- **Cost at large scale.** Managed carries a substantial premium over raw compute. Beyond a certain fleet size, dedicated database expertise becomes cheaper than the markup — but this crossover arrives far later than most teams assume, and the comparison must include the salary, not merely the instance price.
- **Data residency or compliance** requiring on-premises hosting or a specific physical location the provider does not offer.
- **Portability and lock-in avoidance**, particularly where a multi-cloud or exit strategy is a genuine requirement rather than an aspiration.
- **Unsupported engine or version** — an engine the provider does not offer, or a version needed sooner than the provider ships it.
- **Local development and trivial scale**, where a container on a laptop is simpler and free.

**The decision in one line:** managed databases buy operational competence that is expensive to hire and easy to get wrong. Decline them when the workload needs control the managed offering forbids, or when scale makes the premium exceed the cost of the expertise.

---

## Question 5 — In-Memory Databases & Caching

> - What are in-memory databases and how do they differ from traditional databases?
> - When should you use Redis or Valkey for caching versus storing data in a traditional database?
> - What are common use cases for in-memory databases? (Name 3-4)

### a. What they are, and how they differ

An in-memory database keeps its working data in **RAM** (Random Access Memory) as the primary store, rather than on disk with memory used as a cache. **Valkey** is a Linux Foundation fork of Redis created in 2024 after a licence change, and is API-compatible with it.

The differences that matter:

- **Latency.** Memory access is measured in nanoseconds and a local Redis round-trip in tens of microseconds, against milliseconds for a disk-backed query. Typically one to two orders of magnitude faster.
- **Durability.** Disk-backed databases treat durability as the default guarantee. In-memory stores treat it as an option — Redis offers snapshotting and an append-only log, but both trade performance for safety, and the default posture is that data can be lost on restart.
- **Capacity and cost.** The dataset must fit in RAM, which costs roughly an order of magnitude more per gigabyte than disk. Datasets are therefore bounded deliberately, usually with eviction policies.
- **Data model and query power.** Simple key-value access plus purpose-built structures (lists, sets, sorted sets, hashes, streams) rather than a general query language. No joins, no ad-hoc querying, and transactional guarantees far weaker than **ACID** (Atomicity, Consistency, Isolation, Durability).

### b. Cache versus system of record

The governing rule: **a cache must be reconstructible.** If losing the store loses information, it is not a cache — it is a database, and it needs durability guarantees an in-memory store is not designed to provide.

**Use Redis or Valkey as a cache when** the workload is read-heavy, the data is expensive to compute or fetch, some staleness is tolerable, and the authoritative copy lives elsewhere.

**Keep data in the traditional database when** it is the system of record, must survive restarts, requires transactions or referential integrity, or must be queried by many different attributes.

The failure mode to avoid is the cache that quietly becomes a database because someone stored something there that exists nowhere else.

### c. Common use cases

1. **Read-side caching** — query results, computed aggregates, rendered fragments.
2. **Session storage** — shared session state across stateless application instances. This is the canonical use in a horizontally scaled deployment, since no single instance can hold the session.
3. **Rate limiting and counters** — atomic increment with expiry, which is exactly what the data structures are built for.
4. **Leaderboards and rankings** — sorted sets maintain ordering natively.
5. **Job queues and pub/sub** — lightweight asynchronous work distribution without a full message broker.

**Relevance to CollabDocs:** phase 1 deliberately chose no cache, which is correct at MVP scale where the database absorbs the load comfortably and a cache would add an invalidation problem. Use case 2 is the one that will actually arrive: once authentication lands in phase 2, five stateless replicas cannot hold sessions in process, so session state must move to a shared store — a signed token or a shared in-memory store being the two candidate answers.

---

## Question 6 — Async Communication

> - What is the difference between synchronous and asynchronous communication in distributed systems?
> - What are the benefits of using message queues in your architecture?
> - Compare Kafka, SQS, and RabbitMQ. When would you choose each?

### a. Synchronous versus asynchronous

**Synchronous** communication means the caller waits for the callee's response before continuing. Both parties must be available *at the same moment*, the caller's latency includes the callee's, and a failure downstream propagates upward immediately. This is temporal coupling, and it is the correct choice when the caller genuinely cannot proceed without the answer.

**Asynchronous** communication means the caller hands off a message and continues. The receiver processes it later, possibly much later. The caller needs only the broker to be available, not the consumer. Failures are absorbed by retries rather than propagated.

The trade is not latency for throughput but **immediacy for decoupling**: an asynchronous system gives up knowing that the work has been done in exchange for surviving the consumer being absent.

### b. Benefits of message queues

- **Temporal decoupling** — producer and consumer need not run simultaneously; the consumer can be redeployed, restarted or scaled without the producer noticing.
- **Load levelling** — a burst is absorbed by the queue and drained at the consumer's sustainable rate, protecting downstream systems from traffic they cannot handle.
- **Resilience** — retries, redelivery and **DLQs** (Dead-Letter Queues) for messages that repeatedly fail, instead of work simply being lost.
- **Responsiveness** — slow work leaves the request path. The user gets an immediate response while the email, the export or the thumbnail is produced afterwards.
- **Independent scalability** — throughput increases by adding consumers, with no change to the producer.
- **Fan-out** — one event consumed by several independent subscribers, each at its own pace.

The costs are real and should be stated alongside: eventual consistency, ordering that is guaranteed only within limits, duplicate delivery requiring **idempotent** consumers, and materially harder debugging because causality is spread across processes and time.

### c. Kafka, SQS and RabbitMQ compared

| | **RabbitMQ** | **Kafka** | **SQS** (Simple Queue Service) |
|---|---|---|---|
| Shape | Traditional broker | Distributed append-only log | Managed cloud queue |
| Message lifetime | Deleted once acknowledged | Retained for a configured period; re-readable | Deleted once acknowledged |
| Routing | Rich — exchanges, topics, fan-out, priorities | Topic and partition only | Minimal |
| Replay | No | **Yes** — consumers track offsets | No |
| Throughput | High | Very high | High, effectively unbounded |
| Operations | Self-hosted or managed | Heaviest to operate | Zero |

**Choose RabbitMQ** when the workload is task distribution with complex routing rules, per-message acknowledgement and priorities — a work queue where each message is a job to be done once.

**Choose Kafka** when the data is a *stream of events* rather than a set of tasks: when several independent consumers need the same events, when replaying history matters (rebuilding a projection, backfilling a new consumer), or when volume is extreme. The mental model is a durable log that consumers read at their own position, not a queue that drains.

**Choose SQS** when already on AWS and the requirement is simple decoupling with no operational burden at all. It is the pragmatic default in that environment; the cost is limited routing and no replay.

**The distinction worth carrying:** RabbitMQ and SQS deliver *work to be consumed*, whereas Kafka retains *facts to be read*. Choosing between them is really choosing which of those the data is.


### d. How each one actually works

The quickest way to hold all three is to ask **who remembers what has been consumed**; everything else follows from the answer.

**RabbitMQ — a broker that hands out work.** Producers publish to an **exchange**, not a queue; bindings and routing keys decide which queues receive a copy (`direct` for exact key match, `topic` for wildcard patterns, `fanout` for a copy to every bound queue). The broker then *pushes* to subscribed consumers and holds each message as unacknowledged until the consumer acks it — if that consumer dies first, the message is redelivered elsewhere. Ack means delete. Prefetch caps how many unacked messages a consumer may hold, which is the flow-control mechanism. The protocol is **AMQP** (Advanced Message Queuing Protocol) over a persistent connection. *Smart broker, dumb consumer:* per-message state lives in the broker.

**Kafka — an append-only log that consumers read.** A topic is divided into **partitions**, each an ordered, immutable sequence appended to disk; ordering holds within a partition only, and producers select a partition by hashing a key so that related records stay ordered. Consumers *pull* and each tracks its own **offset** into the log. Consumption deletes nothing — records expire on a retention policy — so offsets can be rewound to replay history, and independent consumer groups read the same partition at different positions simultaneously. Its throughput comes from sequential disk writes and zero-copy transfer rather than per-message bookkeeping. *Dumb broker, smart consumer.* The operational catch: **parallelism is capped by partition count** — a group cannot usefully employ more consumers than there are partitions.

**SQS — a queue as an HTTP endpoint.** No broker to operate and no persistent connection; consumers call `ReceiveMessage` over HTTP with long polling. The mechanism to understand is the **visibility timeout**: receiving a message hides it from other consumers for a period rather than removing it, and it is deleted only when the consumer explicitly calls `DeleteMessage`. A consumer that crashes first simply lets the timer expire, and the message reappears for someone else. At-least-once delivery therefore rests on **a timer rather than connection state**, which is precisely what allows the service to be stateless and scale without bound.

| | RabbitMQ | Kafka | SQS |
|---|---|---|---|
| Who tracks progress | Broker, per message | **Consumer**, one offset | Broker, via timer |
| After consumption | Deleted | **Retained** | Deleted |
| Delivery | Push | Pull | Pull (poll) |
| Parallelism limit | Any number of consumers | **Partition count** | Unlimited |
| Replay history | No | **Yes** | No |


---

## Question 7 — Monitoring & Observability

> - What is the difference between monitoring and observability?
> - How do commercial APM tools (Datadog, New Relic) differ from platform metrics (CloudWatch)?

### a. Monitoring versus observability

**Monitoring** is watching known signals for known failure modes. The questions are decided in advance — is error rate above a threshold, is the queue depth growing, is the disk filling — and dashboards and alerts are built to answer them. Monitoring tells you **that** something is wrong.

**Observability** is a property of the system: the ability to ask *new* questions about its internal state from its outputs, without shipping new code to answer them. It addresses failures nobody anticipated — why requests from one particular customer on one endpoint became slow after a specific deploy. Observability helps you find **why**.

The distinction is not which tool is used but whether the question had to be known in advance. Practically, observability requires rich, high-cardinality context attached to events — request identifiers, user identifiers, versions, feature flags — so that arbitrary slicing after the fact remains possible. The three commonly cited pillars are logs, metrics and traces.

### b. APM tools versus platform metrics

**Platform metrics** — CloudWatch and its equivalents — observe from the *infrastructure's* vantage point: processor and memory utilisation, disk activity, request counts and latencies at the load balancer, container restarts. They are cheap, built in, and know nothing whatsoever about the application's internals.

**APM** (Application Performance Monitoring) tools — Datadog, New Relic, Sentry — instrument from *inside* the process. They attribute time to individual functions and queries, follow a single request across service boundaries as a distributed trace, group errors with stack traces and correlate them with releases, and connect all of it back to specific users or endpoints.

**In one line:** platform metrics report that the machine is busy; APM reports which query in which handler made it busy.

The costs are a per-host or per-ingested-gigabyte bill that grows quickly, an agent inside the application, and telemetry — sometimes containing sensitive data — leaving for a third party.

**Relevance to CollabDocs:** phase 1 deliberately builds the minimum in-house — structured logs, a health endpoint that verifies storage reachability, and a metrics endpoint — while deferring commercial monitoring to phase 2. That is this exact trade taken consciously: the signals are produced now, and the decision about who watches them is postponed until there is something worth watching.

---

## Question 8 — Integration & Reliability

> - How do you handle commercial service failures? (Name 3 strategies)
> - How do you avoid vendor lock-in?

### a. Handling commercial service failures

The premise to internalise first: a service the code calls synchronously is a **runtime dependency that can fail mid-request**, and it will.

**1. Timeouts, bounded retries, and exponential backoff with jitter.**
Every outbound call gets an explicit timeout — the default in most clients is far too long or absent, and an unbounded wait converts a slow dependency into an outage as connections and threads accumulate. Retries apply only to idempotent operations, are capped, and are spaced by exponentially increasing delays with randomisation so that recovering services are not immediately overwhelmed by every client retrying in lockstep.

**2. Circuit breaker.**
After a threshold of consecutive failures the breaker opens and calls fail immediately without touching the network, for a cooldown period; it then admits a single trial request and closes again if that succeeds. This achieves two things: it stops the caller from exhausting its own resources waiting on something known to be broken, and it gives the failing dependency room to recover instead of being hammered throughout its outage.

**3. Graceful degradation and fallback.**
Decide in advance what the system does *without* each dependency. Serve a stale cached value; queue the work for later delivery; disable one feature while keeping the application up. A signup that fails because the email provider is unavailable is a design error — the account can be created and the message enqueued.

Two supporting practices: **bulkheads**, which give each dependency its own connection or thread pool so one slow service cannot starve the others, and **moving the dependency off the request path entirely** by making the interaction asynchronous, which is the strongest form of the same idea.

### b. Avoiding vendor lock-in

- **Wrap the vendor behind an interface of your own.** The application talks to a small internal abstraction; only the adapter behind it knows the vendor's model. This prevents vendor concepts leaking into the domain and confines a future migration to one file. (This is exactly what CollabDocs phase 3 exists to do for storage, which is why phase 1 forbids it — the abstraction has to be earned by a second implementation, not guessed at.)
- **Own the data model and keep exports current.** Vendor identifiers are stored as attributes, never as primary keys, and a working export is exercised periodically rather than assumed.
- **Prefer open standards and open-source engines.** OIDC rather than a proprietary login flow; PostgreSQL on a managed service rather than a proprietary engine; **OpenTelemetry** instrumentation rather than a vendor agent. The service can then be replaced without the interface changing.
- **Price the exit before signing.** A brief written answer to "what would migrating away cost" at evaluation time is worth more than any amount of architectural insurance afterwards.

**The mature position:** total portability is expensive, and pursuing it usually forfeits the leverage that made buying attractive in the first place. The goal is not zero lock-in but **a known and bounded exit cost**, accepted deliberately.

---

## Question 9 — CI/CD Platforms

> - What is the difference between self-hosted (Jenkins) and SaaS CI/CD (GitHub Actions)?
> - When would you choose each?

### a. The difference

**CI/CD** stands for Continuous Integration and Continuous Delivery (or Deployment): every change is built and tested automatically, and the pipeline can carry it to production.

**Self-hosted (Jenkins)** means running the controller and its build agents yourself. Full control follows: any operating system, special hardware, licensed toolchains, and direct network access to internal systems. So does the operational burden — upgrades, a plugin ecosystem that is powerful but fragile, security patching of a system that by definition holds credentials to everything, and capacity management of the agent fleet. There is no per-minute bill, but there is a permanent maintenance cost.

**SaaS (GitHub Actions, GitLab CI, CircleCI)** means the vendor runs the infrastructure. Builds execute on ephemeral, clean runners, integration with pull requests is immediate, and a marketplace supplies pre-built steps. The constraints are fixed runner specifications, a per-minute bill that scales with build volume, credentials held by the vendor, and no route into a private network without deploying self-hosted runners after all.

### b. When to choose each

**Choose SaaS** as the default — for small teams, standard build shapes and open-source projects. It removes an entire category of infrastructure at a cost that is negligible until build volume becomes large.

**Choose self-hosted** when compliance or an air-gapped environment forbids external services; when builds need specialised hardware, licensed compilers or unusual platforms; when build volume is high enough that per-minute pricing exceeds the cost of owned capacity; or when the pipeline must reach systems inside a private network.

**The common answer in practice is hybrid:** a SaaS control plane orchestrating self-hosted runners, which keeps the vendor's integration and interface while placing execution where control or network access requires it.

---

## Question 10 — AI Services

> - List 3 examples each of Language AI, Vision AI, and ML Platform services.
> - When should you use AI APIs versus training your own models?
> - What are vector databases and why are they important for AI applications?
> - Managed RAG services. What do they handle for you, and when would you use one instead of calling an LLM API directly?
> - When would you connect an existing MCP server (GitHub, Slack, Postgres) to an AI agent instead of writing custom integration code?

### a. Three examples of each

**Language AI** (**LLM** — Large Language Model — APIs): the Anthropic Claude API, the OpenAI GPT API, and the Google Gemini API. Adjacent to these sit translation services (DeepL, Google Translate) and speech-to-text services (Whisper, Deepgram, AWS Transcribe).

**Vision AI:** AWS Rekognition, Google Cloud Vision, and Azure AI Vision — covering object detection, image classification, content moderation and **OCR** (Optical Character Recognition). Document-specific variants include AWS Textract and Google Document AI.

**ML Platforms** (**ML** — Machine Learning): AWS SageMaker, Google Vertex AI, and Azure Machine Learning — for training, hosting and serving models. AWS Bedrock and Azure AI Studio are a related category: gateways offering several vendors' foundation models behind one interface.

### b. AI APIs versus training your own models

**Use the API — the correct default.** Frontier models are trained at a scale no ordinary team can approach, and for general language and vision tasks the hosted model will outperform anything trained in-house while requiring no ML expertise, no training infrastructure and no maintenance. Iteration happens in the prompt rather than in a training run.

**Train or fine-tune your own** when at least one of the following holds: proprietary labelled data exists that provides a genuine advantage; the task is narrow, repetitive and high-volume enough that per-inference cost dominates; latency or offline operation requirements exclude a network call; or regulation forbids data leaving the environment.

The middle ground is where most real answers live and is worth naming explicitly: **prompt engineering, then retrieval, then fine-tuning a small open model** — in that order of escalation. Training a model from scratch is almost never the right answer for an application team.

### c. Vector databases

Text, images and audio can be converted by a model into **embeddings** — vectors of numbers positioned so that semantically similar inputs land near one another. A vector database stores those vectors and answers *nearest-neighbour* queries: given this vector, return the closest ones, using approximate search to remain fast at scale.

Their importance follows from a hard constraint: a language model has a fixed context window and a training cutoff, so a corpus cannot simply be pasted into the prompt. The pattern instead is to embed the corpus in advance, retrieve only the few most relevant fragments at query time, and place those in the prompt. That pattern is **RAG** (Retrieval-Augmented Generation), and the vector database is its retrieval half.

Examples include Pinecone, Weaviate, Qdrant, Milvus, Chroma, and **pgvector** — a PostgreSQL extension, which is frequently the right answer for a system that already runs PostgreSQL and does not need a separate service.

### d. Managed RAG services

Services such as Bedrock Knowledge Bases and Vertex AI Search take over the entire retrieval pipeline: ingesting and parsing source documents, chunking them sensibly, generating and storing embeddings, indexing them, retrieving and re-ranking results at query time, assembling the prompt with citations, and re-synchronising when source documents change.

**Use one instead of calling an LLM directly** when the model must answer over a private corpus that is large, changes over time, or carries per-user access restrictions — and when building and operating that pipeline is not itself the product. It removes a substantial amount of undifferentiated engineering.

**Call the model directly instead** when the corpus is small enough to fit in a modern long-context window, in which case retrieval adds failure modes for no benefit; when retrieval quality must be tuned closely, since chunking and ranking strategy *is* the quality of a RAG system and managed services expose limited control; or when volume makes the managed premium material.

### e. Existing MCP servers versus custom integration

**MCP** (Model Context Protocol) is an open protocol standardising how AI applications connect to external tools and data. A server exposes a system's capabilities in a uniform way, so any compatible agent can use them without bespoke glue written per pairing.

**Connect an existing server** when the target is a common system with a maintained implementation — GitHub, Slack, PostgreSQL. The tool definitions, authentication handling and schemas arrive already built and maintained by someone else, and the same agent can be pointed at a different server later without code changes. This is the buy side of the same build-versus-buy question the rest of this worksheet asks.

**Write custom integration code** when the internal system has no server; when several underlying calls need to be composed into one domain-meaningful operation the agent should perform; when latency or payload size demands a tighter interface; or — most importantly — when **the surface must be deliberately narrow.** A general server may expose far more capability than an agent should be trusted with, and a purpose-built integration that permits exactly three operations is a security control, not merely an optimisation.

---

# Exercises

## Exercise 1 — Online Bookstore: Service Selection

> You are building an online bookstore as a solo developer with limited time and budget.
> **Requirements:** user accounts and login; credit card payments; order confirmation emails; search books by title/author; monitor website uptime; 5,000 active users/month.
> **Task:** match each requirement to a service type, and name one commercial provider for each.

| Requirement | Service type | Provider | Alternatives |
|---|---|---|---|
| User accounts and login | Authentication | **Auth0** | Clerk, WorkOS, AWS Cognito |
| Credit card payments | Payment Processing | **Stripe** | PayPal, Square |
| Order confirmation emails | Communication | **SendGrid** | Postmark, AWS SES |
| Search books by title/author | Search | **Algolia** | Elasticsearch, Typesense |
| Monitor website uptime | Monitoring | **Datadog** | New Relic, Sentry, Better Stack |

### [Advanced] Build vs Buy Analysis

> **Given costs:** Auth0 $200/month; Stripe 2.9% + $0.30 per transaction; SendGrid $20/month; Algolia $50/month; Datadog $30/month.
> **Assumptions:** $100/hour developer time; maintenance 10% of development time per year; monthly revenue $10,000.
> **Tasks:** assume each service takes 4 weeks to build. For each service type, list pros and cons of buying versus building, calculate the 1-year cost of each option, and make a decision.

**Build cost, identical for every service:**

```
4 weeks × 40 hours = 160 hours
160 hours × $100/hour = $16,000  (initial development)
Maintenance: 10% of 160 hours = 16 hours/year = $1,600/year
1-year build cost = $16,000 + $1,600 = $17,600
```

**Stripe requires a transaction-volume assumption.** At $10,000 monthly revenue and an assumed average order of $40, that is 250 orders per month:

```
2.9% × $10,000            = $290.00
250 transactions × $0.30  = $ 75.00
Monthly                   = $365.00   →  1-year = $4,380
```

**One-year comparison, per service**

| Service | Buy (1 yr) | Build (1 yr) | Verdict | Break-even |
|---|---|---|---|---|
| Auth0 | $2,400 | $17,600 | Buy | ~20 years |
| Stripe | $4,380 | $17,600 | Buy | ~5.8 years — but see below |
| SendGrid | $240 | $17,600 | Buy | **never** |
| Algolia | $600 | $17,600 | Buy | **never** |
| Datadog | $360 | $17,600 | Buy | **never** |
| **Total** | **$7,980** | **$88,000** | | |

**The result worth presenting:** for SendGrid, Algolia and Datadog the build *never* breaks even at any horizon — because annual maintenance alone ($1,600) exceeds the annual subscription. The built version is more expensive forever, before the initial $16,000 is even considered. This is the sharpest illustration on the worksheet of why "it's just a subscription" is the wrong frame.

**Pros and cons, by service**

- **Authentication.** Buying transfers legal exposure and supplies MFA, SSO and compliance attestations. Building means owning every flow, every future vulnerability class, and the breach. Buy.
- **Payments.** Building means handling card data, which means PCI DSS compliance — an audit programme, not a feature. The four-week estimate is not merely optimistic, it is a category error. Buy; this one is not a genuine choice.
- **Email.** Building means SMTP plumbing plus the real work: domain reputation, SPF, DKIM and DMARC, bounce handling and spam-folder avoidance. $240 a year against that is not a contest. Buy.
- **Search.** Building means tokenisation, stemming, typo tolerance, ranking and relevance tuning — search relevance is a specialist discipline. Buy. (PostgreSQL full-text search is a legitimate free middle option if $50/month matters.)
- **Monitoring.** Building means becoming responsible for the system that tells you when the system is broken, which must be more reliable than the thing it watches. Buy.

**Decision: buy all five.**

Annual revenue is $120,000. Buying costs **$7,980 — 6.7% of revenue**. Building costs **$88,000 in year one — 73% of revenue** — and consumes **20 weeks of development**, roughly five months during which the bookstore itself does not exist. For a solo developer that opportunity cost, not the invoice, is the decisive figure.

---

## Exercise 2 — AI Service Matching

> Match each business need to the appropriate AI service type.

| Business need | AI service type | Why |
|---|---|---|
| Automatically respond to customer support emails | **Language Model API** | Open-ended text comprehension and generation |
| Verify uploaded ID documents are valid | **Computer Vision API** | Image analysis and OCR of a structured document |
| Convert podcast audio to searchable text | **Speech-to-Text API** | Audio transcription |
| Recommend similar products to customers | **Vector Database** | Similarity search over product embeddings |
| Filter inappropriate user-uploaded images | **Content Moderation API** | Purpose-trained classification against policy categories |
| Translate product descriptions to 10 languages | **Translation API** | High-volume, well-defined language conversion |

**Two refinements worth adding in discussion:**

- **ID verification** in production rarely uses a raw vision API. Dedicated identity-verification vendors (Onfido, Persona, Jumio) combine OCR with document forensics, liveness checks and watchlist screening. A general vision API reads the text on the card; it does not tell you the card is genuine — which is the actual requirement.
- **Translation** can be done by a language model, and it is better where tone, brand voice or context matters. A dedicated translation API is cheaper and faster at volume. Ten languages of product copy is volume work, so the translation API is the right default.

---

## Exercise 3 — IoT Platform Decision

> Smart factory with 100,000 sensors sending temperature/pressure data every 30 seconds.
> **Message Broker:** A) AWS IoT Core (managed, $9,000/month for 100K devices) or B) self-hosted MQTT broker on EC2 ($100/month infrastructure).
> **Database:** A) AWS Timestream (managed time-series, $500/month) or B) self-managed InfluxDB on EC2 ($200/month infrastructure).
> **Processing:** A) AWS Lambda (serverless, $2,000/month estimated) or B) ECS containers (always running, $400/month).
> **Tasks:** calculate message volume; choose and justify each option; total both approaches; state when you would reconsider.

**IoT** = Internet of Things. **MQTT** = Message Queuing Telemetry Transport.

### Message volume

```
Each sensor: 1 message / 30 s = 2 messages / minute

Messages per minute  = 100,000 × 2                  = 200,000
Messages per second  = 200,000 ÷ 60                 ≈ 3,333
Messages per hour    = 200,000 × 60                 = 12,000,000
Messages per day     = 12,000,000 × 24              = 288,000,000
Messages per month   = 288,000,000 × 30             = 8,640,000,000   (8.64 billion)

Data per month @ 100 bytes = 8.64×10⁹ × 100 bytes   = 864 GB / month
```

### Decision 1 — Message broker: **self-hosted (Option B)**

The delta is **$8,900/month = $106,800/year**, which approaches the fully loaded cost of an engineer. That is large enough to justify taking on operational work, and it is the only one of the three decisions where that is true.

Two honest qualifications. First, **the $100/month figure is not realistic** for 100,000 concurrent MQTT connections with any redundancy: a three-node clustered broker (EMQX, VerneMQ) behind a load balancer, with **TLS** (Transport Layer Security) termination, is more like $500–1,000/month. The saving is still roughly $8,000/month. Second, **the hard part is not throughput but identity** — 3,333 messages/second is unremarkable for a modern broker, whereas issuing, rotating and revoking credentials for 100,000 devices is a genuine system that has to be built and operated.

*What would change this:* significantly fewer devices, absence of anyone able to operate a broker cluster, or a compliance requirement making managed device identity worth its premium.

### Decision 2 — Database: **managed Timestream (Option A)**

The delta is **$300/month = $3,600/year**, under 40 hours of engineer time annually. Operating a time-series database ingesting 864 GB per month — retention policies, downsampling, compaction, backup and restore, scaling — will consume far more than that in the first month alone, let alone every year after.

This is the instructive contrast with decision 1: **identical question, opposite answer, decided entirely by the size of the delta.** Self-hosting is not a philosophy, it is an arithmetic result.

*What would change this:* a much larger managed bill at higher retention or query volume, or a requirement Timestream cannot express.

### Decision 3 — Processing: **ECS containers (Option B)**

**ECS** = Elastic Container Service. The delta is **$1,600/month = $19,200/year**, but the deciding factor is workload *shape* rather than price. The load here is perfectly constant: 3,333 messages per second, continuously, with no idle periods. Serverless pricing rewards idleness — scale-to-zero, pay per invocation — so a workload with zero idle time is precisely the case where it loses. Always-on containers at steady high utilisation are the correct shape, and they are cheaper here as well.

**The general rule:** serverless wins on bursty, unpredictable or low-duty-cycle work; containers win on steady, high-utilisation work.

*What would change this:* genuinely spiky load, or a large increase in per-message processing complexity making elastic scaling worth its premium.

### Total cost

| Approach | Broker | Database | Processing | Monthly | Annual |
|---|---|---|---|---|---|
| **All managed** | $9,000 | $500 | $2,000 | **$11,500** | **$138,000** |
| **All self-hosted** | $100 | $200 | $400 | **$700** | **$8,400** |
| **Recommended mix** | ~$750 | $500 | $400 | **~$1,650** | **~$19,800** |

**Choice: the mix.** It captures roughly **$118,000 per year** of the available saving while taking on exactly **one** operational burden instead of three. All-self-hosted saves a further $11,000 a year and costs two additional systems to operate — a poor trade at these figures.

The general principle this exercise teaches: **"managed versus self-hosted" is not a posture to adopt system-wide.** It is decided per component, by comparing the delta against the operational burden it buys back. The same question produced three different answers here.

### When to reconsider

- **Below roughly 10,000 devices** — IoT Core's bill falls with device count while a broker cluster's floor cost does not; managed likely wins outright.
- **Above roughly 500,000 devices** — the saving grows, but so does the cluster, and it starts to require dedicated platform engineers rather than spare capacity.
- **Loss of the person who operates the broker.** Bus factor is a legitimate migration trigger and is usually noticed too late.
- **Rising incident frequency** — if the self-hosted component is generating outages, the saving was illusory.
- **Unpredictable growth** — managed services absorb bursts that a fixed cluster cannot, and elasticity has real value when the device count is not known in advance.
- **A new compliance requirement**, particularly around device identity and audit.
- **Provider price cuts**, which are frequent and asymmetric — the build-versus-buy line moves on its own without anything in the system changing.
