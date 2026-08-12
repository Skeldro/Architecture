# Handoff: Software Architecture Learning Session (2026-08-12)

Context for the assistant reading this: Andrew (C/C++ systems programmer, Infinity Labs graduate, in job placement) completed a Socratic-style learning session over a 5-item software architecture worksheet. This file summarizes what was covered, what he got right/wrong, and what remains open. Method used throughout: he answers first, then gets corrected — continue that pattern if resuming the session. Standing rule: he must state his plain-language approach before writing any code.

## Worksheet items covered

### 1. SDLC phases

White Paper → Domain Analysis → Requirements Gathering → Solution Analysis → Design → Implementation → Integration → Testing → Release → Retirement.

Key corrections he received:

* Domain Analysis = studying the problem domain (business concepts, rules, actors, vocabulary) — NOT technical scoping. He initially confused it with tech-stack analysis.
* Requirements split into functional (what the system does — the verbs) vs. non-functional (how well — the adverbs: performance, scalability, availability, security, maintainability, portability, plus external constraints).
* Solution Analysis vs. Design: solution analysis selects among competing approaches ("which approach?"); design elaborates the chosen one into a buildable spec ("how exactly?"). Distinguished by scale and reversibility, not by "deciding vs. not deciding."
* Integration is a phase because interfaces are where independently-tested modules break.

### 2. Where architecture sits

* Architecture begins after requirements, during solution analysis; before detailed design.
* Non-functional requirements drive architecture: almost any architecture satisfies the functionals; non-functionals eliminate candidate architectures ("architecturally significant requirements"). Cross-cutting, nearly impossible to retrofit.
* Architectural vs. detailed-design classification is relative to system scale — the test is blast radius: how much of the system's structure does the decision constrain? (Worked example: the Watchdog's signal-based IPC is arguably architectural in a two-process system, detail in a large one.)

### 3. Methodologies (one-word essence each)

Waterfall = sequence (right when requirements frozen/regulated). Prototyping = learning (prototype is disposable). Incremental = slices. Spiral = risk (retire the riskiest unknown first). Lean = waste-elimination in the process, not the binary ("decide as late as possible"). XP = engineering practices (TDD, pairing, CI) — orthogonal, combinable. Scrum = time-boxed sprints, team synchronization. Kanban = continuous flow with WIP limits.

Applied to his EHS SaaS (solo, pre-customer): XP + Lean + Kanban-solo fit; Waterfall is the actively wrong one; Spiral's lesson = the riskiest unknown is market risk (will anyone pay flat-per-org), attack that before serious code. Notable correction: he initially defined Lean as "thorough upfront planning to avoid needless stuff" — that inversion (Waterfall wearing Lean's jacket) was flagged as his most practically important mistake, given a documented tendency to over-architect before validating.

### 4. Architecture layers

Application (use cases, thin) / Domain (model of the business reality) / Generic Infrastructure (in-house reusable, domain-agnostic) / Off-the-Shelf (bought components) / Platform (OS, runtime, possibly cloud substrate — the cloud blurred OTS vs. Platform).

Key points he worked through:

* Layering = dependencies point only downward through defined interfaces.
* IP concentrates in Generic Infrastructure: domain knowledge is cheap to observe/reconstruct; hardened infrastructure is accumulated solved-hard-problems. (id Software engine licensing used as the case study — engine = id's infrastructure, licensee's OTS; layer classification is relative to whose stack you're in. Timeline verified: Doom/Quake/id Tech 3 were commercially licensed for years before GPL release.)
* Healthy migration flows downward (code sinks): application code proves reusable → generalized → sinks to infrastructure; infrastructure gets replaced by matured commercial products; categories sink to platform over time. Upward migration = specialization = usually a layering violation. Migration direction is the mechanism that concentrates IP in infrastructure.
* CRM exercise corrections: Application = use cases, not delivery channels ("website" is packaging); Domain = entities/rules a salesperson would recognize pre-software (Customer, Lead, Deal, pipeline states, invariants) — NOT business KPIs.

### 5. Core concepts (definitions settled)

* Stub: interface-honoring stand-in that decouples progress from dependencies.
* Architecture: the set of highest-level design decisions about a system's structure — major components, how they communicate, the constraining principles — the decisions hardest to change later.
* Framework vs. library: inversion of control ("don't call us, we'll call you"); his own phrasing "a library is me using tools, a framework is me writing code to be used by a tool" was approved.
* Design pattern: named solution template re-implemented per situation, not reusable code; the naming provides compressed communication between engineers.
* Plug-in: third-party-writable component loaded at runtime through a host-defined interface (dlopen/dlsym; he has built this).

## OPEN / UNFINISHED

1. Infrastructure one-liner — he never wrote his own definition (item 5 list). Should produce it from item 4 material.
2. The synthesis quote — "It's hard to achieve the full benefits of a Scrum heartbeat of sprints until you have basic infrastructure stubs that support frameworks and plug-ins." His explanation is still owed. Scaffold given to him: five devs, sprint 1, empty repo — what breaks, and which of the four terms fixes which part? (Intended answer direction: sprints must ship working increments every two weeks; without infrastructure there's nothing to stand on; without stubs teams can't work in parallel on unfinished dependencies; frameworks/plug-in seams give devs isolated extension points so work parallelizes without merge collisions.)
3. Item 2's earlier open question from a prior session — which methodologies fit the EHS SaaS — was answered this session (see item 3 above).

## His recurring error patterns (for the next tutor)

* Answers the interesting question rather than the one asked; reaches for mechanisms before stating properties.
* Dismisses items as "self explanatory" instead of articulating them (testing/release/retirement; Scrum/Kanban) — don't let it slide.
* Confuses problem-domain analysis with technical analysis (twice: domain analysis, CRM domain layer).
* Justifies correct answers with adjacent-but-wrong reasoning (migration direction justified by dependency rule; non-functionals justified by "functional = front end").
