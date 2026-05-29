# starterpack

A deployable Turborepo starter: **Vite + TanStack Router** frontend, a **Go
hexagonal** backend, a **shadcn/ui** design system, and feature-toggled
integrations (Clerk, Stripe, analytics, email, error tracking). A re-platform of
[next-forge](https://www.next-forge.com) onto Vite + Go.

## Quick start

```bash
make setup      # install deps + tools, generate code
make dev        # run everything
```

| URL | App |
|-----|-----|
| http://localhost:3000 | `app` — dashboard (Vite + TanStack Router) |
| http://localhost:3001 | `web` — marketing site |
| http://localhost:3002 | `api` — Go backend (Gin, hexagonal) |
| http://localhost:3003 | `email` — React Email preview |
| http://localhost:6006 | `storybook` — design system |

**Prerequisites:** [Bun](https://bun.sh) ≥ 1.3, [Go](https://go.dev/dl/) ≥ 1.24,
GNU Make, and (optional) Docker for local PostgreSQL. `make setup` installs the
Go CLIs (sqlc, dbmate, swag) for you. Run `make help` to list every target.

Only **Clerk** (auth) and **PostgreSQL** are needed for the full experience —
every other integration is a feature toggle that stays inert until its key is
set. Without `DATABASE_URL` the API uses an in-memory store; without Clerk keys
auth is bypassed in dev, so the stack always boots from a fresh clone.

## Documentation

- **[starterpack-docs/setup.md](./starterpack-docs/setup.md)** — clone-to-running guide
- **[starterpack-docs/docs.md](./starterpack-docs/docs.md)** — every change vs. next-forge, tooling decisions, feature-toggle matrix
- **[skills/starterpack/](./skills/starterpack/)** — Claude skill: architecture, packages, and customization references for this repo

## Stack

- **Frontend:** Vite, TanStack Router, TanStack Query, Tailwind v4, shadcn/ui
- **Backend:** Go, hexagonal (ports & adapters), Gin, zerolog, pgx + sqlc, dbmate
- **Auth:** Clerk (graceful when unconfigured)
- **Contract:** Go OpenAPI (swag) → typed TS client (openapi-typescript + openapi-fetch)
- **Tooling:** Bun, Turborepo, a single `Makefile` entrypoint (`make help`)

## Layout

```
apps/       api (Go) · app · web · email · storybook
packages/   api-client · auth · design-system · email · typescript-config
```

## Common commands

```bash
make dev          # Go API + all JS apps concurrently
make build        # build all JS apps + the Go binary
make migrate      # apply dbmate migrations (needs DATABASE_URL)
make generate     # regenerate sqlc + OpenAPI + typed client
make lint         # ultracite (JS) + go vet
make test         # bun test + go test
make help         # list every target
```
