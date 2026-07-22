# Backend Scope (Go)

`apps/backend` — Go 1.24. **Gin** (HTTP) · **GORM** (Postgres) · **Viper**
(config) · **Zap** (logging) · **go-redis** · **golang-jwt**. Domain-Driven
module layout, mirroring `aloevol/e_commerce/backend`.

## Load With

- `.agents/shared/rules/go-patterns.md`
- `.agents/shared/rules/agent-safety.md` (always — this is a browser agent)

## Layout

```text
cmd/api/                 entrypoint (main) — load config+logger, run app
internal/
├── app/                 wiring: build Gin engine, middleware, mount modules
├── modules/<domain>/    one folder per domain (DDD / clean architecture)
│   ├── domain/          entity.go, repository.go (interface)
│   ├── application/     dto.go, service.go (business logic)
│   ├── infrastructure/  repository.go (GORM / external impls)
│   └── interfaces/      handler.go, routes.go (Gin)
├── shared/
│   ├── config/          Viper config
│   ├── middleware/       cors, auth, logger, rate_limit, security_headers
│   └── infrastructure/  db + cache clients
└── websocket/           the realtime agent gateway (goroutine per session)
pkg/                     reusable, domain-free: response, logger, jwt, pagination
```

Planned Charli modules: `chat`, `agent` (the ReAct loop), `tools` (skill
registry), `audit`. Cross-cutting: `safety` (policy engine) and the `llm`
client live in `shared/infrastructure`; `health` is the reference module.

## Defaults

- The LLM never executes. It selects a tool; the safety engine decides; the
  handler/service acts. Deterministic Go owns authorization (~80/20).
- Controllers (`interfaces`) only validate + delegate; logic lives in `application`.
- All HTTP responses use `pkg/response` (`{success,message,data,error}`).
- `context.Context` first arg on anything doing I/O. Wrap errors with `%w`.
- Sensitive fields redacted before any page content reaches the `llm` client.

## Verify

```bash
moon run backend:check   # go vet
moon run backend:lint    # golangci-lint
moon run backend:test    # go test ./... -race -cover
moon run backend:build
```
