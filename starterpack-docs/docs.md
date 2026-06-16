# Changes — starterpack vs. stock next-forge

This document records every change made to re-platform **next-forge** (Next.js +
Prisma) into **starterpack** (Vite + TanStack Router frontend, Go hexagonal
backend), and how to reproduce each step manually. It mirrors the git history:

| Commit | Phase |
|--------|-------|
| `scaffold next-forge baseline` | 0 — pristine `next-forge init` |
| `prune to target shape` | 1 — remove Next apps + unused integrations |
| `fresh shadcn/ui package + auth forms` | 2 — design system |
| `Go hexagonal backend` | 4 — `apps/api` |
| `Vite + TanStack Router frontends` | 3 — `apps/app`, `apps/web` |
| `typed OpenAPI client + Makefile` | 5 & 6 |

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
Makefile           Single entrypoint (wraps bun/turbo + Go toolchain)
```

## Tooling decisions

| Concern | Choice |
|--------|--------|
| Package manager / runtime | Bun |
| Backend | Go, hexagonal (ports & adapters), **Gin**, **zerolog** |
| DB access | **sqlc + pgx**; migrations via **Atlas** |
| Go validation | `go-playground/validator` + domain value-object constructors |
| Frontend | **Vite + TanStack Router**, `app` + `web` split |
| API contract | Go **OpenAPI** (swag) → **openapi-typescript + openapi-fetch** |
| Design system | fresh **shadcn/ui** + Storybook + custom Clerk-wired auth forms |

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

Ports & adapters layout:
```
cmd/api/main.go                  composition root (config, logger, wiring, server)
internal/domain/user/            entity + Username/Email value objects + Repository port
internal/application/user/       use cases (depend only on the port)
internal/adapters/http/          Gin handlers, DTOs (validator tags), middleware
internal/adapters/persistence/   postgres (pgx+sqlc) and memory implementations
internal/config/                 env-driven feature toggles
internal/platform/logger/        zerolog
db/{migrations,queries,schema.sql}  Atlas + sqlc inputs
docs/                            generated OpenAPI spec (swag)
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
- **Validation in two layers:** Gin binding tags (`binding:"min=2,max=6"`) at the
  edge, and authoritative domain constructors (`NewUsername`/`NewEmail`).
- **Feature toggles:** `internal/config` reads env; each optional service has an
  `Enabled()`; Clerk middleware is mounted only when keyed; without
  `DATABASE_URL` the API uses the in-memory repository so it always boots.
- **Errors:** domain errors are mapped to HTTP status in `adapters/http/response.go`.

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

---

## Feature toggles

Every integration is enabled only when its key(s) are present; otherwise it is
inert and the app still runs.

| Service | Status | Enable by setting | Where |
|---------|--------|-------------------|-------|
| **Clerk** (auth) | needed | `CLERK_SECRET_KEY`, `VITE_CLERK_PUBLISHABLE_KEY` | api + app |
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

## Deferred (intentionally not in this pass)

- Observability stack (Prometheus + Grafana)
- SSR/SEO for the marketing site (currently a client-rendered SPA)
- Stripe checkout, Resend sending, self-hosted PostHog wiring
- Deployment manifests (Docker images / Compose / hosting)

## Known follow-ups

- `make typecheck` (or `bun run typecheck`) needs the generated route trees; run
  `make dev` or `make build` once first (vite generates them).
- The app's main JS chunk is large; add route-level `manualChunks` later.
- Live DB path (Atlas → pgx) verified structurally via sqlc codegen; run
  `make db-apply && make dev` against a real PostgreSQL to exercise it end-to-end.
