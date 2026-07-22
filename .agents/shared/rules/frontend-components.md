---
paths:
  - 'apps/website/**/*.tsx'
  - 'apps/website/**/*.ts'
  - 'apps/extension/**/*.tsx'
---

# Frontend Component Standards

Design identity: **"Powerful Simplicity"** — clean, fast, trustworthy
(like Linear). Tailwind + shadcn/ui (new-york, base color `neutral`, `lucide`
icons). No inline styles; Tailwind classes only.

## Card

```tsx
<div className="bg-white rounded-xl border border-slate-200 p-6 shadow-sm hover:shadow-md transition-shadow">
  <div className="flex items-center justify-between mb-4">
    <h3 className="text-base font-semibold text-slate-800">Title</h3>
    <span className="text-xs text-slate-400">Meta</span>
  </div>
  {/* content */}
</div>
```

Always: `rounded-xl` · `p-6` · `border border-slate-200` · `shadow-sm hover:shadow-md transition-shadow`.

## Button

```tsx
// Primary
<button className="inline-flex items-center gap-2 px-4 py-2 bg-blue-600 text-white text-sm font-medium rounded-lg hover:bg-blue-700 active:bg-blue-800 transition-colors disabled:opacity-50 disabled:cursor-not-allowed">

// Secondary
<button className="inline-flex items-center gap-2 px-4 py-2 border border-slate-200 text-slate-700 text-sm font-medium rounded-lg hover:bg-slate-50 transition-colors">

// Icon button (36×36)
<button className="w-9 h-9 flex items-center justify-center text-slate-400 hover:text-slate-600 hover:bg-slate-100 rounded-lg transition-colors" title="Tooltip">
  <Icon className="w-4 h-4" />
</button>
```

Loading: swap the icon for `<Loader2 className="w-4 h-4 animate-spin" />` and disable.

## Form Fields

Label ALWAYS above the input — never use the placeholder as a label.

```tsx
<div className="space-y-1.5">
  <label className="text-sm font-medium text-slate-700">
    Label <span className="text-red-500">*</span>
  </label>
  <input className="w-full px-3 py-2 text-sm border border-slate-200 rounded-lg bg-white placeholder:text-slate-400 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent aria-[invalid=true]:border-red-400" />
  <p className="text-xs text-red-500 flex items-center gap-1">
    <AlertCircle className="w-3 h-3" /> Error message
  </p>
</div>
```

## Non-negotiables

- Every screen handles **loading / error / empty / permission-denied**.
- No `any`. Props typed. Server state via TanStack Query; forms via RHF + Zod.
- Responsive and valid at **375px** minimum.
- Prefer shared `shared/ui` primitives + `lucide` icons over ad-hoc markup.
