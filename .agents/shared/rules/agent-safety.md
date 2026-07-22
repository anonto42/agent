---
paths:
  - 'apps/backend/**/*.go'
  - 'apps/extension/**/*.ts'
  - 'apps/extension/**/*.tsx'
---

# Agent Safety (non-negotiable)

Charli runs on real web pages with the user's real accounts. These rules are
always at least High severity if violated.

## Authorization

- [ ] The model NEVER executes an action directly. It selects a tool; the Go
      `safety` engine decides if it runs.
- [ ] Authorization is decided by deterministic backend code, never by the
      model's claim (no "the model said the user is admin").
- [ ] Every action is checked against the current user + the site's permission.

## Risk tiers → confirmation

- [ ] read / summarize / extract → auto (read-only).
- [ ] fill a visible field → auto (reversible, visible).
- [ ] click submit / buy / send, delete, bulk edit → REQUIRE user confirmation.
- [ ] payment fields, password fields, auth flows → BLOCK.

## Site & data protection

- [ ] Charli is OFF by default; enabled per-site by explicit user action.
- [ ] Never auto-run on banking / auth / payment pages.
- [ ] Sensitive fields (passwords, card numbers, tokens) are REDACTED before any
      page content is sent to the LLM.
- [ ] Actions are visible (highlight the target before acting).
- [ ] A single kill switch (Esc) stops any in-progress action, always.

## Audit

- [ ] Every tool call (user, tool, args, result, timestamp) is written to the
      audit log. No silent actions.
