# Frontend instructions

Inherit ../AGENTS.md. This module uses Vue 3.5, strict TypeScript 5.7, Vite 6, CSS and nginx.

## Code style

- Use Composition API with `<script setup lang="ts">`, typed props/events and PascalCase component filenames. Prefer `ref`, `computed` and declarative templates over manual DOM work.
- Prefer modern supported syntax: const, ES modules, async/await, optional chaining, nullish coalescing and TypeScript type inference. Vue 3.5 reactive props destructuring and defineModel are available when useful. Avoid unnecessary type assertions, explicit any and non-null assertions.
- Use four spaces, single quotes in scripts and Prettier formatting. Keep CSS readable and follow Stylelint. Use semantic HTML, accessible labels and explicit button behavior.
- Treat server responses as authoritative. Check HTTP status, display failures, expose connection status and prevent accidental duplicate submissions.
- Preserve the same idempotency key and payload when retrying an uncertain bet result. Do not silently accept a changed coefficient.
- Keep minor units integer throughout transport; format currency only for display. Avoid floating-point balance calculations.
- Keep frontend dependencies in this module; shared lint dependencies belong to the root. Update the appropriate lockfile when dependencies change.
- nginx serves built assets and proxies /api; preserve SSE buffering and timeout settings.

## Commands

From the repository root:

```powershell
npm --prefix frontend ci
npm --prefix frontend run build
npm exec --prefix frontend -- vite frontend --host 127.0.0.1
npm run lint
npm run format
pwsh -File scripts/lint.ps1
```

The Vite command is a UI development server; it has no API proxy configured. Use Docker Compose for the complete working application. Build runs vue-tsc before Vite. After every change, run the full root validation, not only the frontend build.

## Retry safety

- Persist the original pending request before sending. Recover it after reload; never generate a new key for an uncertain result.
- Only clear a request after confirmed acceptance or a definitive application rejection. Treat transport errors, malformed replies and server failures as uncertain.

## Tailwind CSS

- Use Tailwind CSS v4 utilities in Vue templates and the official Vite plugin. Keep complete class names in source so the scanner can detect them; do not construct names dynamically.
- Keep global base rules in style.css. Prefer utilities over component style blocks. Stylelint checks CSS files; the production build verifies Tailwind compilation.
