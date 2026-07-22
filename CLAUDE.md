# CLAUDE.md — Charli

Context for AI agents working in this repo.

## What Charli is

A flexible **browser agent**. The user talks to it in a side panel on any web
page; Charli perceives the page and helps do work — spanning a capability
spectrum (build in this order, each level ships as its own product):

- **L0** inline text assist (Grammarly-style: rewrite the selection)
- **L1** page understanding (summarize / extract / answer, read-only)
- **L2** single actions (fill a field, click) — gated by the safety engine
- **L3** multi-step tasks (a ReAct loop that drives the page)
- **L4** cross-app workflows (data from one app into another)

## Architecture (the non-negotiable rule)

The LLM never touches the database or executes anything directly. It only
**selects a tool**; the Go backend validates (permissions → risk → confirmation)
and the extension's content script performs the concrete DOM action. Roughly
**20% agent decisions / 80% deterministic backend code**.

```text
extension (perceive + act)  <--SSE + POST-->  Go backend (agent loop + safety)  -->  postgres/redis
```

**Realtime transport: SSE + POST, not WebSockets.** The agent needs to stream
steps DOWN (SSE `/events`) and receive user actions UP (`POST /chat`,
`/confirm`, `/interrupt`). SSE covers server→client push with built-in
reconnect; POST covers client→server. WebSockets were prototyped then dropped —
we don't need a full-duplex channel for this, and SSE+POST is simpler. Reach for
WebSockets only if a continuous two-way stream is ever required (L3/L4).

## Layout & ownership

- `apps/backend` (Go) — **Gin + GORM + Viper + Zap**, DDD module layout
  (`internal/modules/<domain>/{domain,application,infrastructure,interfaces}`,
  `internal/shared`, `internal/stream`, `pkg/`), mirroring
  `aloevol/e_commerce/backend`. Hosts the agent loop, tool registry, safety
  engine, memory, LLM client. Hot reload via `air`; lint via `golangci-lint`.
  Module: `github.com/levelaxis/charli/backend`. See
  `.agents/shared/scopes/backend.md`.
- `apps/extension` (WXT + React) — side panel chat, content scripts (perception
  via the accessibility tree; actions gated by the backend).
- `apps/website` (Next.js) — marketing + user console + audit-log viewer.
- `packages/shared` (TS) — generated types; **do not hand-edit** `types.gen.ts`.
- `contracts` (Go) — the **source of truth** for shared types. Edit `types.go`,
  then `moon run contracts:generate` (tygo) to update the TS side.

## Toolchain

- **moon** runs everything with caching: `moon run <project>:<task>` or
  `moon run :<task>` across all projects. Tasks: `dev`, `build`, `test`, `check`
  (backend also `lint` = golangci-lint).
- **pnpm** for the TS world only; **go.work** for the Go world. moon bridges them.
- Go runs from the system PATH (`platform: system` in each Go `moon.yml`).

## Conventions

- Keep the backend dependency-light; prefer the standard library until a real
  need appears (matches the "start simple, add later" philosophy).
- Sensitive fields (passwords, card numbers, tokens) are redacted before any
  page content is sent to the LLM.
- Every tool call the agent makes is audit-logged.
