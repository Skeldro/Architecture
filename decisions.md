# Decision Log

Every non-trivial decision gets an entry **when it's made**, not reconstructed later — the point is honest reasoning, including what was rejected. Written in plain language, presentation-ready for mentor sessions. Numbered globally (D1, D2, …), grouped by phase.

Entry format:

```
### D<n>: <short title>
- **Decision:** what we chose
- **Why:** the reasoning — and what we rejected or gave up
- **Revisit if:** (optional) what would change our mind
```

---

## Phase 0

### D1: Own repository per project
- **Decision:** The course gets its own repo (`Skeldro/Architecture`), not a folder in an existing one.
- **Why:** Independent history and clean presentation; course state shouldn't be tangled with unrelated work.

### D2: Branch per phase
- **Decision:** Each phase is developed on its own branch (`phase-N`) cut from `main`, merged back when the phase completes.
- **Why:** `main` only ever holds finished phases, so it's always presentable; the merge is a natural checkpoint matching the mentor-session cadence. Process/meta files (README, this log's scaffold) go straight to `main`.

### D3: Decision log (this file)
- **Decision:** Track decisions and reasoning in `decisions.md`, sectioned by phase.
- **Why:** Mentor sessions revolve around presenting and discussing decisions. Logging at decision time captures the real reasoning; writing it up afterwards produces revisionist, too-tidy justifications.
