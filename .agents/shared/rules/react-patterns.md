---
paths:
  - 'apps/website/**/*.tsx'
  - 'apps/website/**/*.ts'
  - 'apps/extension/**/*.tsx'
  - 'apps/extension/**/*.ts'
---

# React Patterns (clean components)

Reference: `aloevol/omni.dns/frontend`. The goal: **components stay
presentational; logic lives in hooks; data access lives in an api layer.**

## Feature slice anatomy

```text
features/<name>/
├── api/        thin, typed functions over the shared client (one call each)
├── hooks/      use<Feature>() — owns ALL state, effects, handlers
├── ui/         presentational components (props in, JSX out)
└── index.ts    barrel: export the container + public types
```

## Container-hook split (the important one)

- **One custom hook per feature** (`useChat`, `useAuditLog`) owns state, data
  loading, and every handler. It returns a flat object of values + callbacks.
- The **container** calls that hook and composes presentational children,
  passing props down. No business logic in JSX.
- **Child components are dumb**: receive data + callbacks as props, render, call
  back on interaction. They never fetch or own domain state.

```tsx
export function RecordManager() {
  const { loading, entries, handleAdd, handleDelete } = useRecordManager();
  return <RecordTable loading={loading} entries={entries} onDelete={handleDelete} />;
}
```

## API layer

- `api/` functions are one-liners over the shared client (`@shared/api`). Every
  function is typed. On the website the client returns the backend
  `{ success, message, data, error }` envelope (mirrors Go `pkg/response`).
- **Never call `fetch` inside a component or hook body** — go through `api/`.

## Shared UI primitives

- Reusable pieces live in `shared/ui` (Button, Card, StatCard, PageHeader,
  ConfirmDialog…): typed props, `cn()` to merge classes, a `loading` prop that
  renders a Skeleton, slots like `actions`. Compose them; don't re-implement
  markup per feature.
- Icons: `lucide-react`. User feedback: `sonner` (`toast.success/error`).

## Every component handles

- **loading / error / empty / permission-denied** — no exceptions.
- No `any`; props typed via an `interface`. Tailwind only + `cn()`; no inline
  styles (except the extension skeleton until Tailwind is wired there).
