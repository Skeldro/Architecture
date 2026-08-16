# Working agreement — Architecture Course

This repo is a software-architecture learning project (Collab Docs, built in phases). Claude types everything; **Andrew makes every decision** and must state his plain-language approach before any code is written.

Method rules:

- **Socratic**: Andrew answers first, then gets corrected. Never hand him answers to open worksheet items.
- **Decisions before code**: every architecture decision gets a `decisions.md` entry (his reasoning, his words for mentor sessions) *before* it's implemented.
- **Scope of `decisions.md`**: architecture decisions only — the system's structure, tech, storage, interfaces. Process/workflow decisions belong in README.md.
- **Don't let "self explanatory" slide** — every item gets articulated.
- Respect the current phase's constraints in `phases/phase-N.md` (phase 0: no abstractions, no tests — messy on purpose; later phases earn the structure).

Where things live: `phases/` = specs per phase; `decisions.md` = the decision log; `handoffs/` = dated session summaries; `docs/` = runtime user data (gitignored). Branch per phase, merged to `main` when the phase completes.
