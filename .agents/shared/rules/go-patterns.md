---
paths:
  - 'apps/backend/**/*.go'
  - 'contracts/**/*.go'
---

# Go Patterns

Stack: Gin · GORM · Viper · Zap · go-redis. Lint with `golangci-lint`
(`.golangci.yml`); hot reload with `air` (`.air.toml`).

## General

- `gofmt -s` + `goimports` clean (local prefix `github.com/levelaxis/charli`).
  Import groups: stdlib, third-party, local.
- `context.Context` is the first parameter of anything doing I/O or cancellable.
  Never store it in a struct.
- Errors: wrap with `fmt.Errorf("doing X: %w", err)`. Handle every error;
  `errcheck` is enforced. Deliberate ignores get a comment.
- Accept interfaces, return concrete types. Interfaces defined by the consumer.
- Exported symbols have doc comments starting with the symbol name.
- No global mutable state; inject dependencies. All wiring in `internal/app`.

## Module architecture (DDD)

- `interfaces/` (Gin handlers) validate input, call `application`, return via
  `pkg/response`. No business logic here.
- `application/` holds services (domain logic) + DTOs.
- `domain/` holds entities + repository interfaces.
- `infrastructure/` implements repositories (GORM) + external clients.
- Register each module's routes via its `RegisterRoutes(group, handler)`.

## HTTP responses

- Always use `pkg/response`: `response.OK/Created/Error`.
- Never leak internal (5xx) error details, SQL, or stack traces to clients.

## Agent-loop specific

- The loop emits ONE tool call per turn, applies the result, then re-decides.
- Bound every loop: max turns + `context` deadline + user kill switch.
- SSE handlers block in `c.Stream` on the session channel; they exit when the
  request `context` is cancelled (client disconnect). No goroutine leaks.
- Tool executors re-validate their arguments server-side — never trust the
  model's arguments as safe.

## Verify

```bash
moon run backend:check && moon run backend:lint && moon run backend:test
```
