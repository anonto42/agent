# Charli

A flexible browser agent: talk to it on any web page (or apps like Sheets/Excel)
and it helps you get work done — from inline text rewriting up to multi-step tasks.

## Monorepo layout

```text
agent/
├── apps/
│   ├── backend/     Go — the agent brain, HTTP + SSE (/events) + POST (/chat)
│   ├── extension/   TypeScript — WXT browser extension (chat side panel + content scripts)
│   └── website/     TypeScript — Next.js marketing site + user console / audit log
├── packages/
│   └── shared/      TS types generated from /contracts (imported by extension + website)
├── contracts/       Go — SINGLE SOURCE OF TRUTH for shared types (tygo -> shared)
└── infra/           docker-compose (postgres + redis)
```

Two build worlds — **Go** (`go.work`) and **TypeScript** (`pnpm-workspace.yaml`) —
tied together by one task runner: **moon** (caches build/test/check across both).

## Prerequisites

- Go 1.24+            (`https://go.dev/dl/`)
- Node 20+ and pnpm  (`corepack enable`)
- moon               (`https://moonrepo.dev` — `proto install moon` or `npm i -g @moonrepo/cli`)
- Docker (for postgres + redis)
- Go dev tools (optional): `air` (hot reload), `golangci-lint` (lint),
  `tygo` (type gen) — all installable via `go install …`

## First-time setup

```bash
docker compose -f infra/docker-compose.yml up -d
pnpm install                                   # TS deps (+ wxt prepare)
```

## Everyday commands (via moon)

```bash
moon run backend:dev        # Go backend (air reload)  -> http://localhost:8080/api/v1/health
moon run extension:dev      # WXT dev (load unpacked in the browser)
moon run website:dev        # Next.js console         -> http://localhost:3000

moon run :build             # build everything (cached)
moon run :test              # unit tests everywhere (cached)
moon run :check             # vet (Go) + typecheck (TS)
moon run :e2e                # full end-to-end (mock LLM + real backend + real panel)

moon run contracts:generate # regenerate shared TS types from contracts/types.go
```

## Build phases (see CLAUDE.md for the full roadmap)

L0 inline assist → L1 page understanding → L2 single actions → L3 multi-step
tasks → L4 cross-app workflows. Start at L0–L1; each phase is a usable product.

Current: **L0 (chat) ✅ · L1 (page perception) ✅ · L2 (single actions, gated by
confirmation) ✅** — Charli can propose filling a field or clicking something,
but nothing runs until you approve it.
