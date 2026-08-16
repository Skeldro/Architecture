# Phase 0 — Starter Kit

**Goal:** The simplest possible version of Collab Docs that proves the product idea works in a browser. There is no architecture yet — messy code is fine, on purpose. Phase 1 onwards solves the problems this minimal version creates; starting messy is what makes those later decisions feel earned.

## Functional requirements

1. A user can create a new document with a title.
2. A user can open a document and edit its content.
3. A user can list existing documents and reopen one of them.

## Non-functional requirements

- The application starts via a single command (`npm run dev` or equivalent for the chosen runtime).
- Usable in a browser at a single URL.
- Documents and their content survive an application restart: stopping the process and starting it again must not lose data. In-memory-only storage is not enough — the data must live on disk in some form.

## Constraints

- No databases
- No authentication
- No real-time synchronization
- No abstractions
- No tests
- No Docker / deployment configuration

## Acceptance criteria

- Running the single start command brings the application up, usable at a known URL in a browser.
- All three functional behaviors — create, edit, list — work end to end.
- Edits to a document persist across a page refresh.
- After stopping and restarting the application, all previously created documents are still present and editable.
