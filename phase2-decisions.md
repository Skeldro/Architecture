# Phase 2 — Architecture Decisions

Phase 2 adds the three off-the-shelf capabilities CollabDocs cannot launch without: **identity**, **real-time coordination**, and **operational visibility**. Each is a deliberate build-versus-buy decision, made against a committed market hypothesis and the project's quality-attribute ranking.

**Status:** Decision 1 settled 2026-08-27. Decisions 2–5 pending.

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
