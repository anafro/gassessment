# Monorepo instructions

## Architecture

```text
.
├── backend/                 Go HTTP API, line generation, wallet/bet transactions, tests
│   ├── main.go              Routes, SSE, two simulated odds sources, Postgres access
│   └── Dockerfile           Go build and minimal runtime image
├── frontend/                Vue 3 + TypeScript application built with Vite
│   ├── App.vue              Match list, live connection, bet slip and wallet UI
│   ├── nginx.conf           Static hosting and /api reverse proxy
│   └── Dockerfile           Node build and nginx runtime
├── scripts/                 PowerShell validation orchestration
├── compose.yaml             Postgres, API and web; persistent database volume
├── package.json             Shared JavaScript lint/format tools (not npm workspaces)
└── README.md                Setup, design, limitations and AI contribution
```

The browser calls same-origin `/api` routes through nginx. Go owns odds and monetary validation; Postgres owns durable wallet and bet records. SSE delivers the in-memory line. There is one demo player and one API instance. Preserve this simple architecture unless a requirement needs more structure. Nested AGENTS.md files provide module-specific rules.

## Code style

- Use UTF-8, LF, final newline, four spaces and no tabs everywhere. This explicit project rule overrides stack defaults, including Go's usual tabs.
- Follow idiomatic Go, Vue Composition API, strict TypeScript, modern CSS and PowerShell conventions. Prefer the newest stable syntax supported by the versions pinned in this repository; do not introduce unsupported syntax or upgrade dependencies merely for syntax sugar.
- Use clear domain names, small functions, early returns and explicit error handling. Prefer standard-library/platform features before new dependencies.
- Prettier owns web/document formatting; ESLint and Stylelint own semantic rules. Go formatting is gofmt followed by expansion of tabs to four spaces. Follow the checked-in configurations.
- Keep monetary amounts as integer minor units, odds as integer hundredths, and preserve atomic, idempotent bet acceptance. Never trust client balance or odds validation.
- Keep README accurate about behavior, verification, limitations and AI authorship. Do not edit generated output, dependency directories or caches.

## Commands

Run from the repository root unless stated otherwise. Prerequisites: Go compatible with backend/go.mod, Node/npm, PowerShell 7 and a running Docker engine.

```powershell
npm ci
npm --prefix frontend ci
Save-Module -Name PSScriptAnalyzer -RequiredVersion 1.24.0 -Path .cache/psmodules -Force
pwsh -File scripts/lint.ps1
pwsh -File scripts/lint.ps1 -Fix
npm run lint
npm run format
docker compose up --build
docker compose down
```

Open http://localhost:8080. `docker compose down -v` also deletes player data; use only when a reset is explicitly intended.

After every change run `pwsh -File scripts/lint.ps1`, including documentation/configuration changes. After formatting, run it again. It checks PowerShell, Go formatting/vet/tests, ESLint, Stylelint, Prettier, TypeScript/build, Compose, Dockerfiles and nginx. Report environment blockers and never claim checks passed unless observed. Docker validation can require image downloads.

## Safety priority

- Money consistency outranks convenience and syntax sugar. Never acknowledge acceptance before commit. An uncertain response must retain the original idempotency key.
- Monetary changes require real Postgres integration tests covering concurrent requests, replay, rejection and rollback. Do not rely solely on mocks or formatting checks.

## Commit rules

- Strictly follow Conventional Commits for every commit, including documentation and tooling changes.
- Use `<type>[optional scope][!]: <description>`. Use lowercase types: `feat`, `fix`, `docs`, `style`, `refactor`, `perf`, `test`, `build`, `ci`, `chore`, or `revert`.
- Use `feat` for new behavior and `fix` for bug fixes. Choose a precise scope when useful, such as `backend`, `frontend`, or `scripts`.
- Write a concise imperative description in English, without a trailing period. Describe the actual change; avoid vague messages such as "updates" or "fix stuff".
- Mark breaking changes with `!` before the colon and explain them in a `BREAKING CHANGE: ...` footer, separated from the body by a blank line. Include migration steps when needed.
- Keep each commit focused on one logical change. Explain significant rationale and validation in the body, separated from the subject by a blank line.
- Before committing, inspect the staged diff and run the required checks. Never include unrelated user changes, secrets, caches, or generated build output.
- Examples: `fix(backend): prevent duplicate wallet debits`, `feat(frontend): restore pending bets after reload`, `docs: document transaction safety rules`.
