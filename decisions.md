# Decision Log — architecture decisions only

Decisions about the **system being built** — its structure, technology, storage, interfaces. Process/workflow decisions (repo layout, branching, course method) live in README.md instead.

Every entry is written **when the decision is made**, not reconstructed later — the point is honest reasoning, including what was rejected. Plain language, presentation-ready for mentor sessions. Numbered globally (D1, D2, …), grouped by phase.

Entry format:

```
### D<n>: <short title>
- **Decision:** what we chose
- **Why:** the reasoning — and what we rejected or gave up
- **Revisit if:** (optional) what would change our mind
```

---

## Phase 0

### D1: Go, standard library only
- **Decision:** Backend in Go (`net/http` + `os`, no framework); frontend is server-rendered HTML forms, no client framework, near-zero JS. Single command = `go run .`.
- **Why:** The language should be boring so the architecture is loud — Go is small, reads like structured C (home turf), and its stdlib covers all of phase 0, which honors the "no abstractions" constraint. Later phases point at real-time collaboration, which lands in Go's core strength (goroutines/channels). Career-wise, C/C++ → Go is the natural systems-to-backend bridge. **Rejected:** another JS project (nothing new learned); C89 (fails the course, not the task — we'd spend every phase on HTTP plumbing instead of architecture).

### D2: Documents are plain files; the title is the filename
- **Decision:** One folder (`docs/`). Each document = `docs/<title>.txt`, raw plain text, newlines are `\n`. Save = full overwrite of the file. List = scan the folder at request time. No metadata file, no index, no dates.
- **Why:** The disk is the single source of truth — nothing cached or duplicated that can drift. The spec didn't ask for dates or metadata, so none exist (Lean). **Rejected:** sidecar JSON per doc (metadata with nothing to hold yet) and a central `index.json` (a second source of truth that can disagree with the folder).
- **Accepted costs (future-phase fodder):** titles must be filename-safe and unique; a crash mid-overwrite can lose a document; every list request is a directory scan. **Revisit if:** any of these actually hurt.
