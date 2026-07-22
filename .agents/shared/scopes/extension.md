# Extension Scope (WXT)

`apps/extension` — WXT (Manifest V3, Chrome + Firefox), React 19, TypeScript.
Side-panel chat UI + content scripts (perception + action).

## Load With

- `.agents/shared/rules/frontend-components.md` (same component style as website)
- `.agents/shared/rules/react-patterns.md`
- `.agents/shared/rules/agent-safety.md`

## Layout

```text
entrypoints/            WXT entrypoints — THIN, wire only
├── background.ts       message router + websocket client to the backend
├── content.ts          perception (a11y tree) + action (click/type)
└── sidepanel/          mounts <ChatApp/> from features
features/               FSD feature slices (chat, ...) — the real UI + logic
shared/                 generic ui / lib / messaging helpers
```

Same FSD discipline as the website: slices export via `index.ts`, use aliases
(`@features/*`, `@shared/*` — configured in `wxt.config.ts`), no cross-layer
relative imports.

## Rules specific to the extension

- Entrypoints stay thin; all UI/logic lives in `features/` and `shared/`.
- The content script NEVER acts without the backend safety engine's go-ahead.
- Redact sensitive fields (password inputs, card numbers) before sending any
  page content over the websocket.
- One kill switch (Esc) must always stop an in-progress action.

## Verify

```bash
moon run extension:check   # tsc --noEmit  (needs `wxt prepare` first)
moon run extension:build
```
