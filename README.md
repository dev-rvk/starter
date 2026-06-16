# starterpack

A deployable Turborepo starter: **Vite + TanStack Router** frontend, a **Go
hexagonal** backend, a **shadcn/ui** design system, and feature-toggled
integrations (Clerk, Stripe, analytics, email, error tracking). A re-platform of
[next-forge](https://www.next-forge.com) onto Vite + Go.

## Quick start

```bash
make setup      # install deps + tools, generate code
make deps-up    # start local PostgreSQL (Docker)
make migrate    # apply database migrations
make dev        # run everything concurrently (bootstrapping)
```

> [!TIP]
> While `make dev` boots the entire stack in one terminal, it is highly recommended to run them separately in two windows/panes for clean log visibility:
> - `make client` runs all JS applications under the Turborepo TUI.
> - `make server` runs the Go backend API with clean, un-obscured stdout logging.

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

## Local dependencies (Docker)

Backing services run via Docker Compose. The flow is **local ↔ hosted by URL**:
run a container locally and point the env var at `localhost`, or skip the
container and point the same var at a managed provider — no code changes.

```bash
make deps-up        # start core services (postgres)
make deps-up-all    # also start redis + mailpit (opt-in profiles)
make deps-down      # stop (keep data)
make deps-reset     # stop and delete data volumes
make deps-logs      # tail logs
```

| Service | Local (compose) | Hosted equivalent | Status |
|---------|-----------------|-------------------|--------|
| PostgreSQL | `postgres` :5433 | Neon · Supabase · RDS | wired today |
| Redis | `redis` :6379 (profile `cache`) | Upstash · Elasticache | scaffold |
| Mailpit (SMTP) | `mailpit` :8025 UI (profile `mail`) | Resend (HTTP API) | scaffold |

To go hosted, e.g. swap Postgres: stop the local one (or just ignore it) and set
`DATABASE_URL` in `apps/api/.env.local` to your Neon/Supabase connection string.

## Configure features

Every integration is a feature toggle — set its key to enable it, leave it blank
to keep it off. Copy the templates first:

```bash
cp apps/api/.env.example apps/api/.env.local
cp apps/app/.env.example apps/app/.env.local
cp apps/web/.env.example apps/web/.env.local
```

| Feature | What it does | Get keys | Env var(s) → file |
|---------|--------------|----------|-------------------|
| **PostgreSQL** *(needed)* | App database | `make deps-up` (local) or [Neon](https://neon.tech) | `DATABASE_URL` → `apps/api/.env.local` |
| **Clerk** *(needed)* | Authentication | [clerk.com](https://clerk.com) → API Keys | `CLERK_SECRET_KEY` → api · `VITE_CLERK_PUBLISHABLE_KEY` → app |
| **Stripe** | Payments (planned) | [dashboard.stripe.com](https://dashboard.stripe.com) | `STRIPE_SECRET_KEY`, `STRIPE_WEBHOOK_SECRET` → api |
| **Sentry** | Error tracking | [sentry.io](https://sentry.io) | `SENTRY_DSN` → api · `VITE_SENTRY_DSN` → app |
| **Resend** | Transactional email | [resend.com](https://resend.com) | `RESEND_TOKEN`, `RESEND_FROM` → api |
| **Google Analytics** | Web analytics (free tier) | GA4 property | `VITE_GA_MEASUREMENT_ID` → app/web |
| **PostHog** | Product analytics (self-host) | your PostHog instance | `VITE_POSTHOG_KEY`, `VITE_POSTHOG_HOST` → app/web |

**Minimum to boot:** nothing — the API falls back to in-memory and auth is
bypassed in dev. For the full experience, set `DATABASE_URL` + the two Clerk keys.

**Removed from next-forge** (re-add as packages if you need them): Arcjet
(security), BetterStack (use Prometheus/Grafana later), BaseHub (CMS).

See [starterpack-docs/setup.md](./starterpack-docs/setup.md#environment-variables--feature-toggles)
for the full matrix and per-service detail.

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
make dev          # run Go API + all JS apps concurrently (bootstrapping)
make client       # run JS apps via turbo (TUI mode)
make server       # run the Go API backend (clean logs)
make build        # build all JS apps + the Go binary
make migrate      # apply dbmate migrations (needs DATABASE_URL)
make generate     # regenerate sqlc + OpenAPI + typed client
make lint         # ultracite (JS) + go vet
make test         # bun test + go test
make help         # list every target
```
