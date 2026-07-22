# .agents — Charli knowledge base

Shared conventions and reviewer skills for AI agents (and humans) working in
this repo. Mirrors the structure used across LevelAxis projects.

```text
.agents/
├── shared/
│   ├── scopes/     what stack + rules apply where (backend / website / extension)
│   ├── rules/      concrete, enforceable patterns (frontend components, go, safety)
│   └── memory/     commands + facts worth keeping
└── skills/
    └── code-reviewer/   review checklist tuned to Charli's conventions
```

## How to use

1. Pick the **scope** for the area you're touching (`shared/scopes/*`).
2. Load the **rules** that scope references.
3. Before opening a PR, run the reviewer skill's checklist
   (`skills/code-reviewer/references/charli-conventions.md`).

Root context lives in [`../CLAUDE.md`](../CLAUDE.md).
