# Website Scope (Next.js console)

`apps/website` — Next.js App Router, React 19, TypeScript, Tailwind, shadcn/ui.
Marketing site + user console + audit-log viewer.

State: **Zustand** (domain stores) + **TanStack Query** (server state).
Forms: **React Hook Form + Zod**. HTTP: `fetch` wrapper in `shared/api`.

## Load With

- `.agents/shared/rules/frontend-components.md`
- `.agents/shared/rules/react-patterns.md`

## FSD Layers

```text
app -> views -> widgets -> features -> entities -> shared
```

- `app/`: routes only; thin pages (Server Components by default).
- `views/`: full page compositions.
- `widgets/`: composed page sections.
- `features/`: user actions — API hooks, forms, modals (`'use client'`).
- `entities/`: domain types/state (e.g. `session`, `site`, `audit`).
- `shared/`: generic UI, lib, api. No domain knowledge.

Rules:

- No upward or sideways imports. No relative cross-layer imports.
- Use aliases: `@shared/ui`, `@entities/session`, `@features/audit-log`.
- Every slice exports through `index.ts`; no deep imports.
- Access token is memory-only in `shared/api/client.ts` — never `localStorage`.

## Client Boundaries

- `page.tsx` stays a Server Component unless it truly needs the client.
- `features/*/ui/*.tsx` and interactive `widgets` use `'use client'`.
- `shared/ui|lib|api` must not force `'use client'`.

## File Checklist

- Correct FSD layer, exported via `index.ts`, aliases (not relative).
- Handles loading, error, empty, and permission-denied states.
- No `any`; props typed. Tailwind only; no inline styles. Responsive at 375px.

## Verify

```bash
moon run website:check   # tsc --noEmit
pnpm --filter @charli/website lint
```
