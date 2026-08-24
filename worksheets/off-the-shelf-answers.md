# Off-the-Shelf Worksheet — Answers

Answers to the commercial-services worksheet. Written for discussion in mentor sessions.

---

## Question 2 — Build vs Buy

### a. When should you use commercial products instead of building yourself?

**1. Is it your differentiator?**
The primary test, and the one that decides most cases on its own. Build what makes the product good; buy everything else. This is worksheet 1's layer model applied to spending: value concentrates in the Domain and Generic Infrastructure layers, so engineering hours belong there and money belongs in the Off-the-Shelf layer. Nobody chooses a documents app because its password-reset flow is exceptional — so that is not where the four weeks go.

**2. Does the product absorb a legal or compliance burden?**
Buying often purchases an answer to a regulatory requirement rather than a feature. Payment processing carries **PCI DSS** (Payment Card Industry Data Security Standard); authentication vendors carry **SOC 2** (Service Organization Control 2) and **GDPR** (General Data Protection Regulation) tooling. The vendor's audit report answers a customer's security questionnaire; a hand-rolled equivalent means defending your own implementation. That shortens sales cycles, not just development.

**3. What does it genuinely cost to build?**
Developer time at a loaded rate, plus the salaries behind it, plus the **opportunity cost** of what those people are not building instead. For a solo developer opportunity cost dominates every other term: four weeks on authentication is four weeks the actual product does not exist.

**4. Would your version actually be better?**
An honest expertise check. A commercial service has absorbed years of edge cases, attacks, and production incidents you have not seen yet. Where that accumulated hardening matters — security, payments, deliverability — the ready-made product usually wins on quality, not merely on time.

**5. What does the alternative actually cost?**
"Buy" is not one thing. Free and open-source components (PostgreSQL, Redis) cost operations rather than licence fees; subscription services cost money that recurs and grows. The comparison is different in each case, and conflating them produces bad decisions in both directions.

### b. Hidden costs of commercial services

The subscription is not hidden — it is on the invoice. These are the ones that are not:

**1. Lock-in and switching cost.**
Integrate carelessly and the vendor's model spreads through the codebase: its identifiers in your tables, its assumptions in your control flow. Leaving then means a migration rather than a swap. The cost is invisible on the day you sign and enormous on the day you want out.

**2. Integration cost.**
Wiring the service in, learning its model, working around its edge cases, and maintaining that glue. Always non-zero, almost never budgeted, and paid again at every major version.

**3. Pricing that scales with your success.**
Per-user and per-request billing ties your cost curve to your growth curve. A service that is trivially cheap at MVP scale can be punishing at 10× or 100× — and by the time it hurts, cost 1 above has made leaving expensive. This is the hidden cost with real consequences.

**4. Their outage is your outage.**
You inherit the vendor's **SLA** (Service Level Agreement) and their incidents, with no ability to fix anything. Your users experience your outage, and you refresh a status page. This is the bill for having traded control for leverage.

**5. The deprecation treadmill.**
Vendors retire **API** (Application Programming Interface) versions on their schedule. Forced migrations arrive at times you did not choose and cannot defer.

### c. Calculating TCO for build vs buy

**TCO** (Total Cost of Ownership) means totalling both options over the same horizon, because the naive comparison — development cost versus monthly fee — is wrong on both sides. Building is also maintaining; renting is also expanding.

**Horizon: 3 years.** Beyond that, the technology, the vendor landscape and the requirements have all changed enough to make the number fiction.

**Build column**
- Initial development: `hours × loaded hourly rate`
- Maintenance: 10–20% of initial development *per year, indefinitely*
- Infrastructure to run and monitor it
- Opportunity cost of the work displaced

**Buy column**
- Subscription × horizon, **modelled at projected usage**, not today's
- One-time integration development
- Per-transaction fees at projected volume

Then risk-adjust both sides: the expected cost of an outage or breach under each option, and the exit cost if the decision turns out wrong.

**Worked example — authentication, at this worksheet's numbers**

Assumptions: $100/hour, 4 weeks to build (160 hours), 10% annual maintenance, $200/month for the service, 1 week of integration work.

| | Build | Buy |
|---|---|---|
| Up-front | $16,000 | $4,000 (integration) |
| Per year | $1,600 (maintenance) | $2,400 (subscription) |
| **Year 1** | **$17,600** | **$6,400** |
| **3-year total** | **$20,800** | **$11,200** |

Buy wins decisively, and only crosses over at roughly year seven — by which point the built version needs work the estimate never included: **MFA** (Multi-Factor Authentication), **SSO** (Single Sign-On), and continuous security patching.

**The number that flips it:** per-user pricing. If growth takes the subscription from $200 to $2,000/month, buy costs $24,000/year and the arithmetic reverses entirely. So the decision is not a single sum — it is a sum evaluated at *the scale you expect to reach*, which is exactly the term that gets forgotten.

---

## Question 3 — Authentication Services

### a. Why is authentication complex to build correctly?

**1. The surface area is enormous, and none of it is the product.**
Signup, login, logout, password reset, email verification, session management, remember-me, account lockout, rate limiting, **MFA** (Multi-Factor Authentication) enrollment *and recovery*, social-login callbacks, **SSO** (Single Sign-On), device management, audit logs. Each is a separate flow with its own way of being subtly wrong, and not one of them is a reason anybody chooses the product.

**2. The failure mode is catastrophic and silent.**
An ordinary bug loses a paragraph and someone complains. An authentication bug exposes every user's data, produces no crash and no stack trace, and is typically discovered by an outsider. The severity class is different from ordinary defects: the exposure is legal, not merely operational.

**3. Correctness is a moving target.**
Password hashing parameters, session handling, **CSRF** (Cross-Site Request Forgery) defence, credential-stuffing mitigation and breach-list checking all shift over time. What was correct practice five years ago is weak now, and tracking that drift is a permanent maintenance obligation rather than a one-off implementation cost.

**4. It depends on infrastructure outside the application.**
Verification and password-reset messages have to actually arrive, which requires domain reputation and correct **SPF** (Sender Policy Framework), **DKIM** (DomainKeys Identified Mail) and **DMARC** (Domain-based Message Authentication, Reporting and Conformance) configuration. Misconfigured, reset emails land in spam and authentication is silently broken for real users. Federated login adds a second such dependency: being a registered relying party with each identity provider.

### b. Compliance benefits of commercial authentication services

Vendors carry attestations that would otherwise have to be earned: **SOC 2 Type II** (Service Organization Control 2), **ISO 27001**, GDPR tooling for data-subject access and deletion requests, and **HIPAA** (Health Insurance Portability and Accountability Act) alignment where health data is involved. A **DPA** (Data Processing Agreement) places part of the breach liability on the vendor contractually.

The mechanism matters more than the list: **their audit evidence becomes usable as your answer.** When a customer's security questionnaire asks how credentials are stored, the response is an attestation rather than a defence of one's own cryptography. The benefit is therefore commercial as much as legal — it unblocks enterprise sales. Vendors also track regulatory change continuously, which is ongoing work rather than a one-time cost.

### c. Auth0 / WorkOS at $200 per month versus four weeks of building

The financial comparison is Question 2's arithmetic: roughly $20,800 over three years to build against $11,200 to buy, crossing over only around year seven — and that crossover assumes the built version never grows MFA, SSO or continued security patching, which it must.

The decisive term is not financial. Building means **assuming the legal exposure personally**: the breach, the compliance gap and the questionnaire all become yours to answer. Buying transfers a meaningful portion of that.

Two conditions reverse the conclusion. **Per-user pricing at scale** — growth that takes the subscription from $200 to $2,000 per month makes buying cost $24,000 a year and flips the sum. And **build genuinely wins** when authentication is itself the product, when requirements exist that no vendor supports, or when data-residency rules forbid the vendor outright.

### d. Externalized authorization (OpenFGA, SpiceDB, Auth0 FGA) — the build-vs-buy trade-off

Authorization is a *different* decision from authentication, despite appearing to be the same shape.

**The options are not all subscriptions.** OpenFGA is open source (a **CNCF** — Cloud Native Computing Foundation — project, originating at Auth0/Okta) and self-hostable at no licence cost. SpiceDB is likewise open source, with an optional managed tier. Auth0 FGA is the hosted product. So "buy" here spans free-but-operated through to fully managed, and the free-versus-subscription distinction from Question 2 applies inside this single category.

**What these tools solve** is the permission *graph*: can this user edit this document, given membership of a team that was granted access to a parent folder? Google published the Zanzibar paper describing its own solution — built for Google Docs — and OpenFGA and SpiceDB are open implementations of that model. This is **ReBAC** (Relationship-Based Access Control), as opposed to simpler **RBAC** (Role-Based Access Control).

**Four reasons the calculus differs from authentication:**

1. **Authorization sits on the hot path.** Authentication happens once per session; authorization is evaluated on every request, often repeatedly. Externalizing it adds a network hop to the most frequently executed path in the system.
2. **The rules are domain knowledge.** Who may edit a document is a business rule — Domain Layer material — whereas authentication is commodity infrastructure. Applying the differentiator test therefore yields *different answers* for the two.
3. **There is a complexity threshold.** Owner/editor/viewer is a column and a `WHERE` clause, and that is the correct answer for most systems. Zanzibar-class tooling earns its place once the graph deepens — nested groups, inherited folder permissions, share links — or once several services must agree on the same rules.
4. **Failure coupling is more severe.** An unavailable authorization service denies everything, which is a harder outage than an unavailable login service.

**Position taken:** buy authentication, build authorization — until the permission graph outgrows a table. For CollabDocs specifically, sharing with people, teams and links is the exact use case Zanzibar was designed for, so the threshold is genuinely reachable; the plan is a `permissions` table in phase 2, with ReBAC tooling reconsidered when nested or inherited sharing arrives.
