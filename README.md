# Architecture Course

Learning software architecture through worksheets + a course project. This folder is the territory: worksheets, session handoffs, and (eventually) the project live here.

## Method (standing rules)

- **Socratic**: Andrew answers first, then gets corrected. The tutor does not hand out answers to open items.
- **Plain-language first**: he states his approach in plain language before writing any code.
- **No "self explanatory"**: every item gets articulated, even the ones that feel obvious.
- Tutor should read the latest handoff in `handoffs/` before resuming — it carries his recurring error patterns.

## Workflow

Project work happens in **phases**. Each phase gets its own branch (`phase-0`, `phase-1`, …) cut from `main`; when the phase is complete it merges back into `main`. Process/meta changes (like this README) can go straight to `main`.

## Files

- `BeforeYouStart.txt` — the original 5-item worksheet (SDLC, architecture placement, methodologies, layers, core concepts).
- `handoffs/` — session summaries, one per session, dated. Latest = current state of the course.

## Status (2026-08-12)

BeforeYouStart worksheet: items 1–5 covered (see `handoffs/2026-08-12-before-you-start.md`).

**Open — owed by Andrew, answer-first:**

1. **Infrastructure one-liner** — his own definition, built from the item-4 layer material.
2. **The synthesis quote** — explain: *"It's hard to achieve the full benefits of a Scrum heartbeat of sprints until you have basic infrastructure stubs that support frameworks and plug-ins."* Scaffold: five devs, sprint 1, empty repo — what breaks, and which of the four terms fixes which part?

Next after that: course project + further worksheets (not yet received).
