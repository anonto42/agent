# Product Roadmap — Backlog from Competitive Research

Ideas pulled from MindStudio, Composio, Traycer, and Base (Base MCP), filtered
to what fits Charli's model (LLM selects a tool; Go backend validates;
extension executes) and mapped onto gaps found in the current implementation
(as of commit `3a23283`, L1 shipped, no L2/L3 yet — see Status below for what
has since shipped).

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

### 5. ~~Plan-preview UX for the L3 ReAct loop~~ — superseded, see Status
Originally: don't gate one action at a time blindly, generate the full plan
up front (Traycer's plan → review → execute flow). Superseded once L3 was
actually built — `.agents/shared/rules/go-patterns.md`'s existing convention
("The loop emits ONE tool call per turn, applies the result, then
re-decides") took precedence over this idea, so L3 does NOT show an upfront
plan; it confirms one step at a time, same as L2, just repeating until the
model is done. The turn-by-turn kill switch this item also asked for
(`agent-safety.md`'s Esc requirement) did ship — see Status.

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
- **L3 (multi-step agent loop): done**, commit `f15d876`. `Reply` starts a
  per-session `taskState`; each confirmed action's result is reported back via
  new `POST /observe`, which appends an `Observation: ...` message and calls
  the model again — repeating until a plain-text final answer, a denial, or a
  6-turn cap (`defaultMaxTurns`). New `POST /interrupt` is the kill switch
  (Stop button + Escape in the panel), and cancels an in-flight LLM call via
  `context.CancelFunc`, not just future turns — a pointer-identity
  `isCurrent` guard stops a stale in-flight reply from reviving an
  interrupted task. Item 5's plan-preview idea was explicitly NOT adopted
  (see item 5). Covered by new Go tests (multi-turn continuation, max-turns,
  interrupt, interrupt-mid-flight-call) and an extended E2E test proving the
  loop continues past a single execute. Verified via `moon run backend:
  check/lint/test`, `moon run extension:check/test`, and `moon run :e2e` —
  all green.
- **Item 2 (graded risk tiers): done**, commit `683a1d2`. `safety.Decision`
  gained `RequiresConfirmation`; `Engine.Evaluate` sets it from each tool's
  `Risk` (`fill` = auto, `click` = confirm, per `agent-safety.md`'s table) and
  denies outright on `RiskBlock`. `service.step()` branches on it: an
  auto-tier action returns `"execute"` directly instead of `"action"`,
  skipping the confirm round-trip; the loop still continues via `Observe`
  afterward exactly as before. No extension changes were needed — the
  `"execute"` handling path already worked regardless of how it was reached.
  Covered by new Go tests per tier plus a `RiskBlock` test, and a new E2E
  test proving `fill` never shows the Approve/Reject card.
- **Item 3 (PII redaction): done.** New `apps/backend/internal/redact`
  package (`Text(string) string`) — regex-based, deliberately narrow to what
  `agent-safety.md` actually names (passwords, card numbers, tokens), not
  emails/phone numbers, so L1 stays useful for ordinary questions about a
  page. Card-number candidates are confirmed with a Luhn check before
  redacting, so ordinary long digit runs (order numbers, etc.) aren't
  false-positived. Wired into `chat/application/service.go`'s `Reply`, where
  page text is folded into the system prompt — `redact.Text(page)` instead of
  raw `page`. Covered by unit tests per pattern plus an HTTP-level test
  (`TestSendRedactsPageContext`) proving a card number never reaches the
  captured LLM prompt.
- **Item 4 (persisted audit log): done.** New `apps/backend/internal/modules/audit`
  module (`domain.Entry` + `Repository` interface, `application.Service`,
  `infrastructure.GormRepository`), plus `internal/shared/infrastructure/database`
  for the GORM/Postgres connection (`go.mod` now depends on `gorm.io/gorm` +
  `gorm.io/driver/postgres` — the backend's first real DB dependency).
  Graceful by design: `cfg.DatabaseURL == ""` or an unreachable database
  means `audit.Service` runs with a nil repo and `Record` no-ops (a `Warn`
  log, nothing more) — `moon run backend:dev` and `moon run :e2e` keep
  working unchanged with no database configured, exactly as confirmed by
  testing this repo's actual `.env` (no `DATABASE_URL` set) and E2E's bare
  binary (no DB either). `chat/application/service.go` now calls
  `s.audit.Record(...)` at every decision point (`step`, `Confirm`,
  `Interrupt`), plus a new one in `Observe` — whether the executed action
  actually succeeded on the page wasn't captured anywhere durable before,
  zap included; now it is (`observed_ok`/`observed_failed`). Verified two
  ways: unit/HTTP tests with a fake in-memory repo (asserting the exact
  outcome sequence a full propose→confirm→observe cycle produces), AND a
  real end-to-end manual run against `docker compose -f
  infra/docker-compose.yml up -d postgres` — confirmed 3 rows landed in the
  `entries` table via `psql`. No HTTP read endpoint yet (write path only,
  per this item's own original framing — a viewer API is separate,
  website-integration work). No integration test against a real Postgres in
  the automated suite (deliberate scope cut — the repository is two thin
  GORM calls; verified manually instead, see above).
- Items 5–6: not adopted / not scheduled, see their entries above — no change.
- Also shipped, not originally on this list: an E2E build-freshness guard
  (commit `3a49f00`) — `extension:dev` was found to dirty the same
  `.output` dir `extension:build` owns, and moon's cache doesn't re-verify
  existing output on a hit, so E2E could silently run against a stale
  dev-mode build. Added an uncached `extension:build-fresh` task and pointed
  `e2e`'s dependency at it.
- **L4 (first cross-app workflow): done.** Google Sheets connect + append —
  extract data from the current page, append it as a row in a Sheet the user
  connects via OAuth. New `apps/backend/internal/modules/google` module
  (`domain.Connection` with `UpdatedAt` meaningful — tokens refresh in place,
  unlike audit's append-only `Entry`; `application.Service` owns
  BeginConnect/CompleteConnect/IsConnected/AppendRow; `GormRepository`
  sharing the same Postgres connection as audit). New `sheets_append` tool
  (`RiskConfirm`) in the registry — no new "extraction" step needed, the
  model already gets page text every turn (L1) and just reasons over it like
  it already does for `fill`. Dependency: `golang.org/x/oauth2` only —
  deliberately not `google.golang.org/api/sheets/v4` (Google's generated
  client), a hand-rolled HTTP call to the one endpoint needed
  (`values.append`) matches `internal/shared/infrastructure/llm`'s existing
  style and CLAUDE.md's dependency-light convention.
  Two real design points: (1) surfaced that the extension had **no
  persistent identity at all** (`stream.ts`'s `session` is per-page-load,
  `chrome.storage` was granted but unused) — new `deviceId.ts` fixes that,
  kept deliberately separate from `session` (task/SSE routing untouched).
  (2) `sheets_append` is dispatched by the extension's **existing**
  `performAction.ts` handler map exactly like `fill`/`click` — "performing"
  it just means calling a backend API instead of touching the DOM, so
  `chat/application`'s loop (`step`/`Confirm`/`Observe`) needed **zero**
  changes; a not-connected failure flows back through the existing
  observe/loop machinery like any other failed action.
  Found and fixed a real bug during E2E verification: `appendRequest`'s
  `deviceId` had `binding:"required"`, so an empty/transient device id
  produced an opaque 400 instead of a clear "not connected" — removed the
  tag; `AppendRow`'s own not-found handling already covers it gracefully.
  Scope cuts (all deliberate, noted in the plan): tokens stored unencrypted
  in Postgres (same trust boundary as the rest of the DB — no KMS anywhere
  in this stack); one connection per device, not multi-user accounts; no
  spreadsheet-picker UI (user names the sheet by URL/ID in chat); no E2E
  test against real/fake-served Google APIs (Go-level tests with an
  httptest-server-backed oauth2.Config cover BeginConnect/CompleteConnect/
  token-refresh/AppendRow; the E2E "not connected" path needs no mocking
  since the test env has no Google credentials by construction).
  Verified: new Go tests (google/application + google/interfaces, including
  the token-refresh path against a fake token server), new extension tests
  (deviceId, performAction's sheets_append case, the connect button/hook),
  new E2E test, AND a real manual run — real Postgres migrated the new
  `connections` table cleanly, `/status` and `/append` both correctly
  reported unavailable with no `GOOGLE_CLIENT_ID` set. `moon run backend:
  check/lint/test`, `moon run extension:check/test`, `moon run :e2e` — all
  green. Not yet verified: an actual live OAuth round-trip (needs the user's
  real Google Cloud credentials in `.env`).
