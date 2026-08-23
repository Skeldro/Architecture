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
