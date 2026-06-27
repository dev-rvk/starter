---
name: starterpack
description: Expert assistance for the starterpack monorepo — a deployable Turborepo with a Vite + TanStack Router frontend, a Go hexagonal (ports & adapters) backend, a shadcn/ui design system, and feature-toggled SaaS integrations. Use this skill whenever the user is working in this repo or asks about its structure, apps, packages, the Go backend (Gin, zerolog, sqlc, pgx, Atlas, Clerk), the Vite apps, the design system, the typed API client, the Makefile workflow, feature toggles, environment variables, how to add features, how to deploy, CI/CD pipelines, GitHub Actions workflows, Docker builds, Cloud Run, Cloudflare Pages, database migrations in CI, branch strategy, staging/production environments, Go best practices, or the Uber Go Style Guide as applied to this codebase — even if they don't name "starterpack" explicitly.
---

# starterpack

starterpack is a production-oriented Turborepo for building SaaS apps. It keeps
the feature set of [next-forge](https://www.next-forge.com) but swaps the stack:
a **Vite + TanStack Router** frontend (two apps), a **Go hexagonal** backend, a
fresh **shadcn/ui** design system, and integrations that are **feature-toggled by
environment variables** — anything without a key is inert, so the app always
boots. The whole monorepo is driven through a single **Makefile**.

## Quick Start

```bash
make setup     # install JS + Go deps, install Go CLIs, run code generators
make deps-up   # start local backing services (postgres) via Docker Compose
make migrate   # apply Atlas migrations (needs DATABASE_URL)
make migrate-new name=create_widgets # generate an Atlas migration
make db-status # show migration status
make db-reset  # reset database and re-apply migrations
make dev       # run the Go API + all JS apps concurrently (bootstrapping)
make client    # run the JS apps via turbo (interactive TUI)
make server    # run the Go API backend (clean stdout logs)
```

Only **PostgreSQL** is needed for the full experience. Without `DATABASE_URL` the API
uses an in-memory store; without Clerk keys the app defaults to local username/password
authentication. Run `make help` to see every target.

## Architecture Overview

Turborepo manages JS/TS workspaces (bun); the Go backend is its own module driven
by the Makefile alongside turbo.

**Apps** (in `/apps/`):

| App | Port | Stack | Purpose |
|-----|------|-------|---------|
| `app` | 3000 | Vite + TanStack Router | Authenticated dashboard |
| `web` | 3001 | Vite + TanStack Router | Marketing site |
| `api` | 3002 | Go, Gin, hexagonal | Backend (REST + OpenAPI) |
| `email` | 3003 | React Email | Email template preview |
| `storybook` | 6006 | Storybook (Vite) | Design system workshop |

**Packages** (in `/packages/`, imported as `@repo/<name>`): `api-client`,
`auth`, `design-system`, `email`, `typescript-config`.

For the full tree, ports, and turbo pipeline, read `references/architecture.md`.

## Key Concepts

### Feature toggles (the core pattern)

Every integration is enabled only when its key(s) are present, and degrades
gracefully otherwise. This keeps the starter runnable from a fresh clone and lets
teams adopt services incrementally.

- **Backend**: `apps/api/internal/config/config.go` reads env into a typed
  `Config`; each optional service exposes `Enabled()`. Adapters are wired only
  when configured (e.g. Clerk middleware mounts only with `CLERK_SECRET_KEY`; the
  Postgres repo is used only with `DATABASE_URL`, else an in-memory repo).
- **Frontend**: `apps/<app>/src/features.ts` maps `VITE_*` env to booleans;
  providers (analytics, auth) mount conditionally.

The full toggle matrix (Clerk, Stripe, Sentry, Resend, Google Analytics,
PostHog) is in `references/setup.md`.

### Hexagonal backend (ports & adapters)

The Go API isolates business logic from frameworks. Dependencies point inward:

```
cmd/api/main.go              composition root (config, logger, wiring, server)
internal/domain/errors.go    shared structured error type + sentinels
internal/domain/<x>/         entities (validate: struct tags), Repository port interface
                             (user, todo, account)
internal/application/<x>/    use cases + XService interface — depend only on ports
                             (user, todo, auth)
internal/adapters/http/      Gin handlers (thin), DTOs (json tags only), middleware
internal/adapters/persistence/  postgres (pgx + sqlc) and memory repositories
internal/platform/           logger (zerolog), validator (New()), jwtutil (local-auth JWTs)
```

Handlers stay thin (decode → use case → encode) in a flat structure using the
naming convention `{resource}_handler.go` and `{resource}_dto.go`. DTOs are pure
data shuttles with only `json:` tags — no `binding:` validation tags. Validation
is the **single source of truth** in the application/service layer, which uses
the platform validator (`internal/platform/validator`) to check `validate:` struct
tags on domain entities.

**Error handling**: shared structured `domain.Error` type in
`internal/domain/errors.go`. Constructors: `domain.NotFound("user")`,
`domain.AlreadyExists("user")`, `domain.ValidationError("user","Field","reason")`.
Each carries a `Kind` (KindUnknown=0, KindNotFound=1, KindAlreadyExists=2,
KindValidation=3) and wraps the corresponding sentinel so both `errors.Is()` and
`errors.As()` work. `response.go` imports only the shared `internal/domain` package
and maps error kinds to HTTP status codes.

**Interface compliance**: every port implementation carries a compile-time
assertion (`var _ userdomain.Repository = (*UserRepository)(nil)`). Every handler
takes the application-layer **interface** (`userapp.UserService`), not the
concrete `*Service` type.

**No `init()`**: the platform validator uses `New()` constructor, instantiated in
`cmd/api/main.go` and injected into services. The only `init()` in the codebase
is the blank import for the Swagger docs registry.

Details and conventions in `references/architecture.md`.
Full Uber Style Guide rules in `references/good-practices.md`.

### Authentication (dual mode)

Auth is always on; the mode is chosen by config in `NewRouter`
(`internal/adapters/http/server.go`):

- **Clerk** — set `CLERK_SECRET_KEY` (+ `VITE_CLERK_PUBLISHABLE_KEY` on the
  frontend). All `/api/v1` routes are guarded by Clerk JWT; no local-auth code is
  wired.
- **Local (default fallback)** — bcrypt + HMAC JWTs. `application/auth` orchestrates
  a credentials `domain/account` (email + password hash) and the profile
  `domain/user`, signed via `platform/jwtutil`. Public `/api/v1/auth/register` and
  `/auth/login`; everything else is behind `middleware.LocalAuth`. On the frontend,
  `@repo/auth` mounts `LocalAuthProvider` instead of `ClerkProvider`.

With neither configured, `/api/v1` is left unprotected and the server logs a
warning — acceptable for a fresh clone, not for production. See
`references/architecture.md` (Authentication) and `references/packages.md`.

### Design system

`@repo/design-system` is a framework-agnostic shadcn/ui package (Tailwind v4
tokens, dark mode via next-themes, ~56 components) plus bespoke auth forms
(login/sign-up/forgot-password) wired to Clerk's headless hooks. Components are
imported by package subpath, e.g.
`@repo/design-system/components/ui/button`. See `references/packages.md`.

### Typed API contract

The Go backend emits an OpenAPI spec (swag). `@repo/api-client` regenerates a
typed TS client via `swag → swagger2openapi → openapi-typescript` and wraps
`openapi-fetch`, injecting a Clerk bearer token. The frontend consumes it with
TanStack Query. Regenerate with `make generate`.

## Common Tasks

### Development

```bash
make dev        # everything (Go API + JS apps) concurrently (bootstrapping)
make client     # JS apps via turbo (TUI mode)
make server     # Go API backend (clean stdout logs)
make dev-js     # alias for client
make dev-api    # alias for server
```

### Local dependencies (Docker Compose)

Backing services are defined in `docker-compose.yml`, managed via the Makefile.
The flow is **local ↔ hosted by URL**: run a container and point the env var at
`localhost`, or skip it and point the same var at a managed provider.

```bash
make deps-up       # core (postgres); make deps-up-all adds redis + mailpit
make deps-down     # stop (keep data);  make deps-reset wipes volumes
```

### Database migrations (Atlas + sqlc)

The schema is owned by SQL migrations, which Atlas generates automatically by diffing `db/schema.sql` (your desired state) against your migration history:

```bash
make db-diff name=create_widgets   # generate a new migration file
make db-apply                      # apply pending migrations to database
make db-status                     # check migrations status
make db-reset                      # teardown and apply from scratch
make sqlc                          # regenerate type-safe Go from SQL
```

Atlas manages migration files in `apps/api/db/migrations/` and tracks them with an integrity file (`atlas.sum`). **sqlc reads the schema.sql file** to generate the Go persistence client.

### Code generation

```bash
make generate     # sqlc + openapi + typed client, in order
make openapi      # regenerate the OpenAPI spec from Go annotations (swag)
make gen-client   # regenerate the TS client from the spec
```

### Adding shadcn/ui components

```bash
bunx shadcn@latest add <component> -c packages/design-system
```

`components.json` is configured so generated imports use
`@repo/design-system/*` (resolvable by consuming apps).

### Quality and build

```bash
make lint    # ultracite (JS) + go vet
make test    # bun test + go test
make build   # all JS apps (turbo) + the Go binary
```

## Uber Go Style Guide — critical rules for this codebase

The applied rules are in `references/good-practices.md`. Non-negotiable rules:

| Rule | What it means here |
|------|-------------------|
| **No `init()`** | `platform/validator` uses `New()` + DI, not package globals |
| **Exit Once** | `main()` only calls `os.Exit(1)` once; startup logic lives in `run() error` |
| **Start Enums at 1** | `ErrorKind`: `KindUnknown=0`, `KindNotFound=1`, `KindAlreadyExists=2`, `KindValidation=3` |
| **Interface compliance** | `var _ Repository = (*Impl)(nil)` in every adapter file |
| **Depend on interfaces** | Handlers take `XService` interface, not `*Service` |
| **No fire-and-forget goroutines** | Server goroutine exposes error channel; graceful shutdown waits for it |
| **Avoid mutable globals** | Inject the validator; do not read package-level vars from service code |
| **Error once** | Either wrap-and-return OR log-and-degrade; never both |
| **Error wrapping** | `fmt.Errorf("context: %w", err)` — terse context prefix, no "failed to" |
| **3-group imports** | stdlib / external / internal, each separated by a blank line |
| **Pre-allocate slices** | `make([]T, 0, len(source))` when output size is known |
| **Table-driven tests** | Every function with >1 input case must have a `tests := []struct{...}` |
| **Functional Options** | Use `Option func(*T)` for constructors with optional deps |
| **Type assertions** | Always use comma-ok: `v, ok := x.(T)` |
| **`time.Time`/`time.Duration`** | Never raw `int` for timestamps or durations |
| **Field tags** | Every marshalled struct field must have explicit `json:` tag |

## Adding a domain — full checklist

```
internal/domain/<x>/
  <x>.go            entity with validate: struct tags
  port.go           Repository interface
  errors.go         domain.NotFound("x") / domain.ValidationError(...) constructors

internal/application/<x>/
  service.go        use cases + XService interface + var _ XService = (*Service)(nil)

internal/adapters/persistence/postgres/
  <x>_repository.go var _ xdomain.Repository = (*XRepository)(nil)

internal/adapters/persistence/memory/
  <x>_repository.go var _ xdomain.Repository = (*XRepository)(nil)

internal/adapters/http/
  <x>_handler.go    thin Gin handlers (decode → service → encode) + swag annotations
  <x>_dto.go        DTOs with only json: tags, no binding: tags

db/queries/<x>.sql   sqlc annotations
```

After all files exist: `make sqlc && make openapi && make gen-client`.

## Deployment & CI/CD

The pipeline uses **GitHub Actions** with three workflow files:

| Workflow | Trigger | What it does |
|----------|---------|--------------|
| `ci.yml` | PR to `main` or `release/prod` | generate-check, Go lint+test, JS lint+typecheck+test-affected |
| `deploy-staging.yml` | Push to `main` | Full tests → Docker build+push → DB migrate → Cloud Run staging → CF Pages staging |
| `deploy-prod.yml` | Push to `release/prod` | Full tests → Docker build+push → manual approval → DB migrate → Cloud Run prod → CF Pages prod |

**Deploy order is critical**: Migrations → API → Frontends. Never reverse.

**Branch strategy**: `main` is trunk (staging deploys on merge). `release/prod`
mirrors production. Feature branches PR into `main`; `main → release/prod` PR
triggers the production deploy.

**Infrastructure**:
- Go API → Docker image → **GCP Cloud Run** (multi-stage Dockerfile at `apps/api/Dockerfile`)
- Frontend apps → static `dist/` → **Cloudflare Pages** (via `wrangler-action`)
- Database → **Neon Postgres** with branch-per-environment (staging/prod)
- Migrations → `make db-migrate-prod` (Atlas via `npx @ariga/atlas@0.37.0`)

**CI-specific Makefile targets** (under `## CI / CD` section):
`test-api`, `test-js`, `test-js-affected`, `lint-api`, `lint-js`, `typecheck`,
`generate-check`, `docker-build`, `docker-push`, `db-migrate-prod`.

**Linting**: `apps/api/.golangci.yml` configures golangci-lint with goimports
(3-group ordering), errorlint (%w enforcement), prealloc, gosec, and gocritic.
CI uses `golangci-lint-action@v6` (with `version: latest` to match the Go
toolchain version) which reads this config automatically. `routeTree.gen.ts`
files are excluded from Biome linting via `biome.jsonc`.

For complete workflow YAML, secrets matrix, GCP IAM setup, and Cloudflare Pages
configuration, see `references/deployment-cloud.md`.

## Reference files

- `references/architecture.md` — full structure, ports, hexagonal layout, turbo + Makefile pipeline
- `references/setup.md` — prerequisites, install, the complete environment-variable / feature-toggle matrix, DB setup, verification
- `references/packages.md` — every app and package, key files, and how they fit together
- `references/customization.md` — swapping providers, adding a domain to the backend, adding a route/feature to the frontend
- `references/good-practices.md` — Uber Go Style Guide rules applied to this codebase (errors, interfaces, init(), goroutines, testing, linting)
- `references/deployment-cloud.md` — CI/CD pipeline, GitHub Actions workflows, Docker, Cloud Run, Cloudflare Pages, secrets, GCP IAM, Neon DB branching
