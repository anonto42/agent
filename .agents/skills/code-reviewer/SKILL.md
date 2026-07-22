---
name: code-reviewer
description: Review changed code against Charli's conventions before merge.
---

# Code Reviewer

Review the current diff against Charli's conventions. Report findings most-severe
first; a convention violation is at least Medium.

## Passes

1. **Correctness** — bugs, unhandled errors, race conditions, missing states
   (loading/error/empty/permission-denied on the frontend).
2. **Scope conventions** — load the scope for each changed area and check it:
   - `apps/backend`, `contracts` → `shared/scopes/backend.md` + `rules/go-patterns.md`
   - `apps/website` → `shared/scopes/website.md` + `rules/frontend-components.md`
   - `apps/extension` → `shared/scopes/extension.md`
3. **Agent safety** — always run `shared/rules/agent-safety.md` for any backend
   or extension change.
4. **Charli conventions** — `references/charli-conventions.md`.

## Verify before approving

```bash
moon run :check && moon run :test
```
