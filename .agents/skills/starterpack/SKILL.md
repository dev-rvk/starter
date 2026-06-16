---
name: starterpack
description: Expert assistance for the starterpack monorepo — a deployable Turborepo with a Vite + TanStack Router frontend, a Go hexagonal (ports & adapters) backend, a shadcn/ui design system, and feature-toggled SaaS integrations. Use this skill whenever the user is working in this repo or asks about its structure, apps, packages, the Go backend (Gin, zerolog, sqlc, pgx, Atlas, Clerk), the Vite apps, the design system, the typed API client, the Makefile workflow, feature toggles, environment variables, or how to add features and deploy — even if they don't name "starterpack" explicitly.
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

Only **Clerk** (auth) and **PostgreSQL** are needed for the full experience.
Without `DATABASE_URL` the API uses an in-memory store; without Clerk keys auth is
bypassed in dev. Run `make help` to see every target.

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
internal/domain/<x>/         entities, value objects (validation in constructors), ports
internal/application/<x>/    use cases — depend only on ports
internal/adapters/http/      Gin handlers (thin), DTOs, middleware, swag annotations
internal/adapters/persistence/  postgres (pgx + sqlc) and memory repositories
```

Handlers stay thin (decode → use case → encode). Validation happens twice: Gin
binding tags at the edge and authoritative domain value-object constructors.
Details and conventions in `references/architecture.md`.

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

## Deployment

Each frontend app builds to static assets (`apps/<app>/dist`) deployable to any
static host/CDN. The Go API compiles to a single binary (`make build-api`) for a
container or VM. Run migrations as a discrete CI/CD step (`atlas migrate apply` keyed off
`DATABASE_URL`). Provisioning manifests are intentionally left to the team.

## Reference files

- `references/architecture.md` — full structure, ports, hexagonal layout, turbo + Makefile pipeline
- `references/setup.md` — prerequisites, install, the complete environment-variable / feature-toggle matrix, DB setup, verification
- `references/packages.md` — every app and package, key files, and how they fit together
- `references/customization.md` — swapping providers, adding a domain to the backend, adding a route/feature to the frontend
