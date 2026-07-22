# Charli Convention Checklist

Non-negotiable rules. A violation is at least Medium severity; safety/auth
violations are High or Critical.

## Agent authorization (Critical)

- [ ] No path where the model's output directly triggers a side effect without
      the `safety` engine validating it.
- [ ] No authorization decision derived from model text (`if agentSaysAdmin`).
- [ ] Confirmation required for destructive/irreversible actions; payment/
      password/auth flows blocked.
- [ ] Sensitive fields redacted before any page content reaches the LLM.
- [ ] Every tool call is audit-logged.

## Boundaries (High)

- [ ] The agent chooses tools; deterministic Go executes and enforces rules
      (~20% agent / 80% backend).
- [ ] Tool executors re-validate their arguments server-side.
- [ ] Loops are bounded (max turns + deadline + kill switch).

## Frontend / FSD (Medium)

- [ ] Correct FSD layer; no upward/sideways or relative cross-layer imports.
- [ ] Slice exported via `index.ts`; imports use aliases (`@shared/*` etc.).
- [ ] Loading / error / empty / permission-denied all handled.
- [ ] No `any`; Tailwind only (no inline styles); responsive at 375px.
- [ ] Access token memory-only; never `localStorage`.

## Contracts / types (Medium)

- [ ] Shared shapes changed in `contracts/types.go`, then regenerated — never
      hand-edit `packages/shared/src/types.gen.ts`.

## Go (Medium)

- [ ] `go vet` clean; errors wrapped with `%w`; `context.Context` threaded.
- [ ] No goroutine leaks; every long-lived goroutine exits on context cancel.
