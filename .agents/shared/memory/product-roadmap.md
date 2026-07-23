# Product Roadmap — Backlog from Competitive Research

Ideas pulled from MindStudio, Composio, Traycer, and Base (Base MCP), filtered
to what fits Charli's model (LLM selects a tool; Go backend validates;
extension executes) and mapped onto gaps found in the current implementation
(as of commit `3a23283`, L1 shipped, no L2/L3 yet).

## Backlog

### 1. Tool registry (foundational — blocks 2 through 5)
Today `contracts.Action.Kind` is a hardcoded string literal switched on in
`internal/safety/policy.go`. Need a real registry: name, category, risk tier,
param schema, per-tool validator — modeled on Composio's intent-based
resolution (LLM expresses intent, backend resolves to a tool) rather than a
fixed switch. Required before L2 grows past fill/click.

### 2. Graded risk tiers in the safety engine
`.agents/shared/rules/agent-safety.md` already specifies the intended tiers
(read/fill = auto, submit/buy/delete = confirm, payment/password/auth =
block), but `safety.Evaluate` (`policy.go`) currently only does a binary
substring blocklist. Implement the tiers the rules doc already promises,
using the tool registry's per-tool risk metadata (item 1) to drive the
decision instead of a blocklist.

### 3. PII/sensitive-field redaction before LLM calls
`agent-safety.md` requires redaction of passwords/card numbers/tokens before
page content reaches the LLM. `apps/extension/shared/lib/page.ts` currently
sends raw `document.body.innerText`, truncated but not scrubbed. Close this
gap. Reference: MindStudio's PII block uses Microsoft Presidio.

### 4. Persisted audit log / run history
`agent-safety.md` requires every tool call logged (user, tool, args, result,
timestamp). Today this is `zap` structured logging only — no DB, no table,
not queryable. Needs Postgres + GORM (already in `infra/docker-compose.yml`)
and a schema, once the website's audit-log viewer needs real data. Reference:
MindStudio's named run history.

### 5. Plan-preview UX for the L3 ReAct loop
When the multi-step loop is built, don't gate one action at a time blindly —
generate the full plan up front, let the user review/edit/approve it, then
execute step-by-step with interrupt (the Esc kill switch already required by
`agent-safety.md`). Reference: Traycer's plan → review → execute flow.

### 6. (Future / L4) Spend-guardrail pattern for agent-initiated payments
Only relevant if/when Charli automates a checkout or payment flow on a page.
Reference: Base MCP's "agent wallets with spend guardrails" — same
confirm/risk-tier concept applied to money instead of DOM actions. Not
scheduled; noted for when L4 scope is defined.

## Not adopted
- MindStudio's drag-drop workflow builder / 1,000+ SaaS integrations — wrong
  product shape; Charli is an embedded page agent, not an agent-authoring
  tool.
- Azumo / The AI Automation Agency — agency service catalogs, no product
  mechanics to borrow. (Note: Azumo's own lineup includes an unrelated
  product also named "Charli" — a naming collision, not a technical
  concern.)
- Prospecta / Sperkline / Kasier Webflow templates — no AI-agent substance,
  pure marketing templates.

## Status
- **Item 1 (tool registry): done.** New `apps/backend/internal/tools` package
  (`Registry`, `Tool{Kind, Risk, PromptExample, Validate}`, `Default()` with
  `fill`/`click`). `internal/safety.Evaluate` is now a method on
  `safety.Engine{registry}` — same behavior (still gates every action to
  confirmation), plus real per-kind arg validation that didn't exist before
  (e.g. an empty-target click is now denied). The chat system prompt is built
  from the registry instead of a hardcoded const. `internal/app` wires
  `tools.Default()` + `safety.NewEngine(reg)` once. Extension's
  `performAction.ts` if-chain replaced with a `Record<string, handler>`
  dispatch map for the same reason. All existing tests updated; new tests
  added for the malformed-args case. Verified via `moon run backend:check/
  lint/test` and `moon run extension:check/test` — all green.
- Items 2–5 not started. Item 2 (graded risk tiers) is unblocked now that
  `Tool.Risk` metadata exists — it just needs `safety.Engine.Evaluate` wired
  to read it instead of gating everything unconditionally.
