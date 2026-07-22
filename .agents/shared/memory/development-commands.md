# Development Commands

One task runner — **moon** — across the Go and TypeScript worlds. It caches
build/test/check, so unchanged projects are skipped.

## Run (dev)

```bash
moon run backend:dev      # Go backend   -> http://localhost:8080/health
moon run extension:dev    # WXT dev; load unpacked in the browser
moon run website:dev      # Next.js       -> http://localhost:3000
```

## Quality (all projects, cached)

```bash
moon run :check           # go vet + tsc --noEmit
moon run :test            # go test + vitest
moon run :build           # go build + wxt build + next build
```

Target one project: `moon run <project>:<task>` (e.g. `moon run backend:test`).

## Contracts (shared types)

```bash
moon run contracts:generate   # contracts/types.go -> packages/shared/src/types.gen.ts
```

## Local services

```bash
docker compose -f infra/docker-compose.yml up -d   # postgres + redis
```

## First-time

```bash
pnpm install                  # TS deps + wxt prepare
# install Go 1.23+, moon (npm i -g @moonrepo/cli), tygo (optional)
```
