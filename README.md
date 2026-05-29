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

See **[starterpack-docs/setup.md](./starterpack-docs/setup.md)** for the full
clone-to-running guide and **[starterpack-docs/docs.md](./starterpack-docs/docs.md)**
for the architecture, tooling decisions, and feature-toggle matrix.

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
