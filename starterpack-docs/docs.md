# Changes & Decisions — starterpack

This document records every change made to re-platform **next-forge** (Next.js +
Prisma) into **starterpack** (Vite + TanStack Router frontend, Go hexagonal
backend), how to reproduce each step manually, and the ongoing architectural
decisions adopted for this codebase.

| Commit | Phase |
|--------|-------|
| `scaffold next-forge baseline` | 0 — pristine `next-forge init` |
| `prune to target shape` | 1 — remove Next apps + unused integrations |
| `fresh shadcn/ui package + auth forms` | 2 — design system |
| `Go hexagonal backend` | 4 — `apps/api` |
| `Vite + TanStack Router frontends` | 3 — `apps/app`, `apps/web` |
| `typed OpenAPI client + Makefile` | 5 & 6 |
| `Uber Go Style Guide adoption` | 7 — architecture standards |

## Final structure

```
apps/
  api/         Go backend — hexagonal (Gin, zerolog, pgx+sqlc, Atlas, Clerk, OpenAPI)
  app/         Vite + TanStack Router dashboard (:3000)
  web/         Vite + TanStack Router marketing site (:3001)
  email/       React Email preview (:3003) — kept from next-forge
  storybook/   Design system workshop (:6006) — switched to Vite builder
packages/
  api-client/      Typed TS client generated from the API's OpenAPI spec
  auth/            Clerk wrapper (mounts only when keyed)
  design-system/   shadcn/ui + Tailwind v4 tokens + auth forms
  email/           Resend + React Email templates — kept, key-gated
  typescript-config/  Shared tsconfig presets — kept
.github/workflows/   CI/CD pipelines (ci, deploy-staging, deploy-prod)
Makefile           Single entrypoint (wraps bun/turbo + Go toolchain + CI/CD targets)
change.md          16-item prioritised Go architecture change plan
```

## Tooling decisions

| Concern | Choice |
|---------|--------|
| Package manager / runtime | Bun |
| Backend | Go, hexagonal (ports & adapters), **Gin**, **zerolog** |
| DB access | **sqlc + pgx**; migrations via **Atlas** |
| Go validation | `go-playground/validator` via `Validator` struct (DI, no `init()`); single source of truth in service layer |
| Frontend | **Vite + TanStack Router**, `app` + `web` split |
| API contract | Go **OpenAPI** (swag) → **openapi-typescript + openapi-fetch** |
| Design system | fresh **shadcn/ui** + Storybook + custom Clerk-wired auth forms |
| Go style guide | **Uber Go Style Guide** — full rules in `.agents/skills/starterpack/references/good-practices.md` |
| Go linting | **golangci-lint** with config at `apps/api/.golangci.yml` |

---

## Phase 0 — Scaffold the baseline

```bash
npx next-forge@latest init --name starterpack-scaffold --package-manager bun --disable-git
# move generated contents into the repo root, keep your own .git, commit
```
The generated `skills/next-forge/` directory documents the original template and
is kept for reference.

## Phase 1 — Prune to the target shape

**Removed apps:** `app`, `web`, `api` (rebuilt later), `docs`, `studio`.
**Removed packages:** `analytics`, `observability`, `security` (Arcjet),
`rate-limit`, `cms`, `seo`, `storage`, `webhooks`, `notifications`,
`collaboration`, `internationalization`, `ai`, `feature-flags`, `next-config`,
`auth`, `database`, `payments`, `design-system` (rebuilt fresh).
**Kept:** `apps/email`, `apps/storybook`, `packages/email`,
`packages/typescript-config`.

```bash
rm -rf apps/app apps/web apps/api apps/docs apps/studio
rm -rf packages/{ai,analytics,auth,cms,collaboration,database,feature-flags,\
internationalization,next-config,notifications,observability,payments,\
rate-limit,security,seo,storage,webhooks,design-system}
rm -rf scripts next-env.d.ts tsup.config.ts .autorc .cursorrules.example
```
Then rewrite root `package.json`: rename to `starterpack`, drop the next-forge
CLI bin/deps and the Prisma `migrate*` scripts; keep `turbo` tasks.

## Phase 2 — Design system (fresh shadcn/ui)

Rebuilt `packages/design-system` as a framework-agnostic React + **Tailwind v4**
package (no Next.js deps):

1. `package.json` with `exports` for `.`, `./components/*`, `./hooks/*`,
   `./lib/utils`, `./styles/globals.css`.
2. `components.json` (shadcn, style `new-york`, aliases pointing at
   `@repo/design-system/*` so generated components resolve when consumed by apps).
3. `src/styles/globals.css` — Tailwind v4 `@theme` with **hex design tokens**:
   shadcn semantic tokens (`--color-background`, `--color-primary`, …) +
   sidebar/chart tokens, a **brand** scale (`--color-brand-*`), neutrals,
   semantic colors (success/warning/danger/info) and **fonts** (Inter / Cal Sans
   / JetBrains Mono). Dark mode overrides the same `--color-*` names under
   `.dark` in `@layer base`. `@source` globs let the single stylesheet scan the
   package and `apps/**` (Tailwind v4 does not scan `node_modules`). Inter +
   JetBrains Mono load via Google Fonts `<link>`s in each app's `index.html` and
   Storybook's `preview-head.html`; add Cal Sans manually to use `font-display`.
4. `src/lib/utils.ts` (`cn`), `src/components/theme-provider.tsx` (next-themes,
   works in Vite SPAs), `src/components/mode-toggle.tsx`.
5. Generate all components: `bunx shadcn@latest add --all -c packages/design-system`.
6. **Bespoke auth forms** in `src/components/auth/` (`LoginForm`, `SignUpForm`,
   `ForgotPasswordForm`) — presentational, props-driven, zod + react-hook-form;
   field rules mirror the Go domain (username 2–6 chars).

**Gotchas reproduced & fixed:**
- shadcn generates `@/...` imports that break for external consumers → rewrite to
  `@repo/design-system/...` and map that alias in the package's `tsconfig`
  (`sed -i 's#from "@/#from "@repo/design-system/#g' src/**`).
- `calendar.tsx` uses react-day-picker v10 → key is `month_grid`, not `table`.

**Storybook → Vite:** `apps/storybook` switched from `@storybook/nextjs` to
`@storybook/react-vite` + `@tailwindcss/vite` (`.storybook/main.ts`), removed
`next.config.ts`/`postcss.config.mjs`, and replaced the one `next/image` usage in
`aspect-ratio.stories.tsx` with a plain `<img>`.

## Phase 4 — Go hexagonal backend (`apps/api`)

Ports & adapters layout (target state — see `change.md` for in-progress items):

```
cmd/api/main.go                  run() error + single os.Exit in main()
internal/domain/errors.go        structured domain.Error type + KindUnknown/NotFound/AlreadyExists/Validation
internal/domain/user/            entity (validate: struct tags) + Repository port
internal/domain/todo/            (same pattern per resource)
internal/application/user/       use cases + UserService interface + compile-time check
internal/application/todo/       use cases + TodoService interface + compile-time check
internal/adapters/http/          flat: {resource}_handler.go, {resource}_dto.go, response.go, middleware/
internal/adapters/persistence/   postgres (pgx+sqlc) and memory implementations, each with interface compliance
internal/config/                 env-driven feature toggles (Enabled() per service)
internal/platform/logger/        zerolog
internal/platform/validator/     Validator struct with New() constructor — no init(), no global var
db/{migrations,queries,schema.sql}  Atlas + sqlc inputs
docs/                            generated OpenAPI spec (swag)
apps/api/.golangci.yml           golangci-lint config
```

Reproduce:

```bash
cd apps/api && go mod init github.com/starterpack/api
go get github.com/gin-gonic/gin github.com/rs/zerolog github.com/jackc/pgx/v5 \
  github.com/go-playground/validator/v10 github.com/clerk/clerk-sdk-go/v2 \
  github.com/joho/godotenv github.com/swaggo/gin-swagger github.com/swaggo/files \
  github.com/google/uuid
sqlc generate                    # db/schema.sql + db/queries → typed Go
swag init -g cmd/api/main.go -o docs   # OpenAPI from annotations
```

Key behaviours:

- **Validation (single source of truth):** validation happens **only** in the
  application/service layer using the injected `*validator.Validator`. Domain
  entities carry `validate:` struct tags. HTTP DTOs are pure data shuttles with
  only `json:` tags — no `binding:` tags.

- **Structured domain errors:** `internal/domain/errors.go` defines a
  `domain.Error` struct with `Kind` (KindUnknown=0, KindNotFound=1,
  KindAlreadyExists=2, KindValidation=3) and wraps sentinel errors. Use the
  constructors: `domain.NotFound("user")`, `domain.AlreadyExists("user")`,
  `domain.ValidationError("user","Email","reason")`. `response.go` maps
  `domErr.Kind` → HTTP status; `KindUnknown` → 500.

- **Interface compliance:** every adapter file carries
  `var _ xdomain.Repository = (*XRepository)(nil)`. Every handler takes
  the `XService` interface, not `*Service`.

- **No `init()` in application code:** `platform/validator` exposes a `Validator`
  struct instantiated via `New()` in `cmd/api/main.go` and injected into services.

- **Exit Once:** all startup logic is in `run() error`; `main()` calls
  `os.Exit(1)` at most once, on `run()` returning an error.

- **Goroutine lifecycle:** the server goroutine sends on a buffered error channel;
  `run()` selects on that channel and the OS quit signal so no goroutine is
  left unobserved.

- **Flat handler structure:** `{resource}_handler.go` and `{resource}_dto.go` in
  the flat `internal/adapters/http/` package.

- **Feature toggles:** `internal/config` reads env; each optional service has an
  `Enabled()`. Clerk middleware is mounted when keyed; otherwise local JWT auth
  is wired as the default. Without `DATABASE_URL` the API uses the in-memory
  repository so it always boots.

## Phase 3 — Frontend apps (Vite + TanStack Router)

`packages/auth` — thin Clerk wrapper: `AuthProvider` mounts `ClerkProvider` only
when `VITE_CLERK_PUBLISHABLE_KEY` is set; re-exports Clerk hooks.

`apps/app` (:3000) — dashboard. `vite.config.ts` uses `@tanstack/router-plugin`,
`@vitejs/plugin-react`, `@tailwindcss/vite`. File-based routes in `src/routes/`:
`__root`, `index` (guarded dashboard + users query), `sign-in`, `sign-up`,
`forgot-password` (design-system forms wired to Clerk headless hooks).
`src/features.ts` is the central toggle map.

`apps/web` (:3001) — marketing `index` + `pricing` using the design system.

**Gotchas reproduced & fixed:**
- The router plugin generates `src/routeTree.gen.ts` during `vite build`/`dev`, so
  the build script is `vite build && tsc --noEmit` (vite first, then typecheck).
  The generated files are committed to version control and excluded from Biome
  linting (`!**/routeTree.gen.ts` in `biome.jsonc`).
- `tsc` in consuming apps can't resolve the design-system source subpaths via
  `exports` alone → add `"@repo/design-system/*": ["../../packages/design-system/src/*"]`
  to each app's `tsconfig` `paths`.

## Phases 5 & 6 — Typed client + Makefile

`packages/api-client` — `bun run generate` chains
**swag → swagger2openapi → openapi-typescript** (swag emits Swagger 2.0; the
converter upgrades it to OpenAPI 3 for `openapi-typescript`). `createApiClient`
wraps `openapi-fetch` and injects a Clerk bearer token. `apps/app` consumes it via
an `ApiProvider` (token-aware when Clerk is on, plain otherwise).

Root **Makefile** is the single entrypoint — see `make help`. It wraps bun/turbo
for JS and the Go toolchain for the backend; `make dev` runs the Go API and all
JS apps concurrently.

## Phase 7 — Uber Go Style Guide adoption

The Go backend now targets the **Uber Go Style Guide** as its canonical code
standard. Applied rules and rationale are in
`.agents/skills/starterpack/references/good-practices.md`.

Prioritised changes are tracked in `change.md` (16 items, 5 waves). Summary:

| Wave | Theme | Key changes |
|------|-------|-------------|
| 1 | Correctness | Remove `init()` from validator; `run() error` in main; `KindUnknown=0` enum fix |
| 2 | Interface hygiene | Compile-time checks; service interfaces; wire TodoHandler |
| 3 | Style & performance | 3-group imports; slice pre-allocation; `.golangci.yml` |
| 4 | Testing | Table-driven tests for all services + handlers |
| 5 | Observability | Request-ID middleware; `go.uber.org/atomic`; functional options; Prometheus metrics |

---

## Feature toggles

Every integration is enabled only when its key(s) are present; otherwise it is
inert and the app still runs.

| Service | Status | Enable by setting | Where |
|---------|--------|-------------------|-------|
| **Clerk** (auth) | optional | `CLERK_SECRET_KEY`, `VITE_CLERK_PUBLISHABLE_KEY` | api + app |
| **Local Auth** | default | `JWT_SECRET` (falls back to dev secret) | api + app |
| **PostgreSQL** | needed | `DATABASE_URL` (else in-memory) | api |
| **Stripe** | later | `STRIPE_SECRET_KEY`, `STRIPE_WEBHOOK_SECRET` | api |
| **Google Analytics** | free tier | `VITE_GA_MEASUREMENT_ID` | app/web |
| **PostHog** (self-host) | optional | `VITE_POSTHOG_KEY`, `VITE_POSTHOG_HOST` | app/web |
| **Sentry** | optional | `SENTRY_DSN` / `VITE_SENTRY_DSN` | api / app |
| **Resend** (email) | later | `RESEND_TOKEN`, `RESEND_FROM` | api + packages/email |
| **Arcjet** | removed | — | (re-add a `security` package if needed) |
| **BetterStack** | removed | — | (use Prometheus/Grafana later) |
| **BaseHub CMS** | removed | — | — |

Backend toggles live in `apps/api/internal/config/config.go` (each has
`Enabled()`); frontend toggles in `apps/<app>/src/features.ts`.

## Local dependencies (Docker)

`docker-compose.yml` defines local backing services, managed via the Makefile
(`make deps-up` / `deps-up-all` / `deps-down` / `deps-reset` / `deps-logs`). The
design is **local ↔ hosted by URL**: run a container and point the env var at
`localhost`, or skip it and point the same var at a managed provider — no code
change, since the apps only read from env.

| Service | Local (compose) | Profile | Hosted equivalent | Status |
|---------|-----------------|---------|-------------------|--------|
| PostgreSQL | `postgres` :5433 | (core) | Neon · Supabase · RDS | wired today |
| Redis | `redis` :6379 | `cache` | Upstash · Elasticache | scaffold for caching/rate-limit |
| Mailpit | `mailpit` :8025 (UI) | `mail` | Resend (HTTP API) | scaffold for local SMTP testing |

Core (`postgres`) starts with `docker compose up`; the others are behind profiles
so they only run when asked. To add another dependency, add a service (behind a
profile if optional), document its local URL and hosted equivalent, and read its
URL from env in the relevant config — mirroring the Postgres pattern.

## Phase 8 — CI/CD pipeline & cloud deployment

Full CI/CD pipeline using **GitHub Actions**, deploying the Go API to **GCP Cloud
Run** and the frontend apps to **Cloudflare Pages**.

### Branch strategy

| Branch | Purpose | Deploy trigger |
|--------|---------|----------------|
| `main` | Trunk development | Push → staging deploy |
| `release/prod` | Production mirror | Push → production deploy (with manual approval gate) |
| `feature/*` | Feature work | PR into `main` (CI only) |

### What was created

| File | Purpose |
|------|---------|
| `apps/api/Dockerfile` | Multi-stage build: `golang:1.26-alpine` → `gcr.io/distroless/static:nonroot`. ~10–15 MB final image. |
| `apps/api/.dockerignore` | Excludes bin/, *.md, .env*, and linter config from Docker context |
| `apps/api/.golangci.yml` | golangci-lint config: goimports (3-group), errorlint (%w), prealloc, gosec, gocritic. Excludes generated sqlc and test files from noisy linters. |
| Root `Makefile` (CI/CD section) | New targets: `test-api`, `test-js`, `test-js-affected`, `lint-api`, `lint-js`, `typecheck`, `generate-check`, `docker-build`, `docker-push`, `db-migrate-prod` |

### Workflow files (`.github/workflows/`)

**`ci.yml`** — runs on every PR to `main` or `release/prod`:
- `generate-check`: runs `make generate`, asserts `git diff --exit-code` on all
  generated paths (sqlc output, OpenAPI docs, `schema.d.ts`). Catches stale
  generated code.
- `api`: golangci-lint + `make test-api` (race detector enabled)
- `js`: `make lint-js` + `make typecheck` + `make test-js-affected` (turbo
  affected filter against `origin/main`)

**`deploy-staging.yml`** — runs on push to `main`:
- Full test suite (Go + JS)
- Build + push Docker image tagged with `$GITHUB_SHA` to GCP Artifact Registry
- `make db-migrate-prod` against `STAGING_DATABASE_URL` (Neon staging branch)
- Deploy API to Cloud Run staging (`starterpack-api-staging`)
- Build + deploy `app` and `web` to Cloudflare Pages (`--branch=staging`)

**`deploy-prod.yml`** — runs on push to `release/prod`:
- Same test + build pipeline
- GitHub Environment `production` gate (manual reviewer approval)
- Migrate against `PROD_DATABASE_URL` (Neon main branch)
- Deploy API to Cloud Run prod (`starterpack-api`)
- Deploy frontends to Cloudflare Pages (`--branch=production`)

### Deploy order (enforced via `needs:`)

```
Migrations → API → Frontends
```

Never reverse. The `@repo/api-client` types are generated from the API’s OpenAPI
spec — deploying frontends before the API creates a window where the client
references endpoints that don’t exist.

### Key decisions

- **Atlas via `npx @ariga/atlas@1.2.3`** in `db-migrate-prod` — matches the
  local Makefile to prevent version drift. No standalone Atlas CLI install.
- **Wrangler-action over CF git integration** — enforces deploy order via
  `needs:`. Disable CF git integration in the dashboard to avoid double deploys.
- **Neon branch-per-environment** — `staging` branch from `main`; each gets its
  own connection string. Migrations run independently against each branch.
- **Secrets via Google Secret Manager** — Cloud Run reads `CLERK_SECRET_KEY`,
  `JWT_SECRET`, `DATABASE_URL` from Secret Manager at runtime (never as plain
  `env_vars` in the revision config).

Full reference: `.agents/skills/starterpack/references/deployment-cloud.md`.

---

## Deferred (intentionally not in this pass)

- Observability stack (Prometheus + Grafana) — tracked as Wave 5 in `change.md`
- SSR/SEO for the marketing site (currently a client-rendered SPA)
- Stripe checkout, Resend sending, self-hosted PostHog wiring
- Request-ID middleware — tracked as Wave 5 in `change.md`
- Functional options pattern on services — tracked as Wave 5 in `change.md`

## Known follow-ups

- `routeTree.gen.ts` files are now committed to version control (un-ignored from
  `.gitignore`) and excluded from Biome linting (`biome.jsonc`). This allows
  `make typecheck` to work in CI without running a Vite build first.
- The app's main JS chunk is large; add route-level `manualChunks` later.
- Live DB path (Atlas → pgx) verified structurally via sqlc codegen; run
  `make db-apply && make dev` against a real PostgreSQL to exercise it end-to-end.
- **Wave 1–5 Go changes** are documented in `change.md`; tackle in order to avoid
  disruption between waves.
- `release/prod` branch needs to be created before first production deploy:
  `git checkout main && git checkout -b release/prod && git push -u origin release/prod`
- Set up GitHub Environments (`staging`, `production`) with required reviewers for
  the production approval gate.
- Create Docker Hub repository (`starterpack-api`) and generate an access token
  (see `DEPLOYMENT.md` Step 1).
- Create GCP project, service account, and IAM bindings (see `DEPLOYMENT.md`
  Step 2). No Artifact Registry needed — images are pushed to Docker Hub.
- Create Cloudflare Pages projects (`starterpack-app`, `starterpack-web`) without
  git integration (see `DEPLOYMENT.md` Step 3).
- Create Neon staging branch and store connection strings as GitHub secrets.
- Populate all GitHub secrets (see `DEPLOYMENT.md` Step 5 for the full matrix).

---

## Phase 9 — Vitest Frontend Testing

Added genuine TDD structure to the frontends using **Vitest**:
- `vitest`, `jsdom`, `@testing-library/react`, and `@testing-library/jest-dom` installed at the root (`package.json`).
- `packages/design-system` has `vitest.config.ts` and `vitest.setup.ts`. Example component tests cover the `Button` and the `cn` utility logic.
- `apps/app` has `test` configurations in `vite.config.ts` (with triple-slash reference for typings).
- `make test-js` / `make test` runs tests smoothly across all Turborepo frontend apps.

---

## Phase 10 — CD Pipeline (Staging)

Implemented `deploy-staging.yml` — the continuous deployment pipeline triggered
on push to `main`:

- **Docker Hub**: Images are built and pushed to Docker Hub (not GCP Artifact
  Registry), keeping the image registry provider-agnostic.
- **Cloud Run**: Pulls the public image from `docker.io/USERNAME/starterpack-api`
  at deploy time.
- **Deploy order enforced**: Migrations → API → Frontends (via `needs:` deps).
- **CI triggers updated**: `ci.yml` now only runs on PRs and manual dispatch.
  Push to `main` is handled by the CD pipeline which includes its own test gates.
- **`DEPLOYMENT.md`**: Step-by-step setup guide with both CLI and UI instructions
  for Docker Hub, GCP, Cloudflare, and Neon.
