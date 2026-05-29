# Setup — from clone to running

This guide takes you from a fresh clone of **starterpack** to a running dev
environment. The stack is a Turborepo with a **Vite + TanStack Router** frontend
(two apps), a **Go hexagonal** backend, a **shadcn/ui** design system, and
feature-toggled integrations. Only **Clerk** (auth) and **PostgreSQL** are needed
for the full experience — everything else gracefully degrades when its key is
absent.

## 1. Prerequisites

| Tool | Version | Notes |
|------|---------|-------|
| [Bun](https://bun.sh) | ≥ 1.3 | Package manager + JS runtime |
| [Go](https://go.dev/dl/) | ≥ 1.24 | Backend (`go` must be on `PATH`) |
| [Docker](https://docs.docker.com/get-docker/) | any | Easiest way to run PostgreSQL locally (optional) |
| GNU Make | any | Orchestrates every command |

Verify:

```bash
bun --version
go version
make --version
```

> **WSL note:** if the project lives on a Windows drive (`/mnt/c/...`), enable
> filesystem metadata so Git/Go/Bun can set permissions. Either run
> `sudo mount -o remount,metadata /mnt/c` (until reboot) or add to `/etc/wsl.conf`:
> `[automount]\noptions = "metadata"` then `wsl --shutdown`. For best performance,
> prefer cloning into the Linux filesystem (`~/`).

## 2. Install dependencies and generate code

```bash
make setup
```

This runs, in order:
- `bun install` — all JS workspace dependencies
- `go mod download` — Go dependencies (in `apps/api`)
- `make tools` — installs the Go CLIs **sqlc**, **dbmate**, **swag** (into your
  `GOBIN`; ensure it is on `PATH`, e.g. `export PATH="$(go env GOPATH)/bin:$PATH"`)
- `make generate` — runs sqlc (DB types), swag (OpenAPI spec), and the typed TS
  API client generator

## 3. Configure environment

Each app ships an `.env.example`. Copy them to `.env.local` and fill in keys:

```bash
cp apps/api/.env.example   apps/api/.env.local
cp apps/app/.env.example   apps/app/.env.local
cp apps/web/.env.example   apps/web/.env.local
```

**Minimum to boot:** nothing — the API falls back to an in-memory store and auth
is bypassed in dev. **For the full experience**, set:

- **PostgreSQL** — in `apps/api/.env.local`:
  ```
  DATABASE_URL=postgres://postgres:postgres@localhost:5433/starterpack?sslmode=disable
  ```
- **Clerk** — create an app at [clerk.com](https://clerk.com), then:
  - `apps/api/.env.local`: `CLERK_SECRET_KEY=sk_test_...`
  - `apps/app/.env.local`: `VITE_CLERK_PUBLISHABLE_KEY=pk_test_...`

Every other key (Stripe, Sentry, Resend, Google Analytics, PostHog) is optional —
see [docs.md](./docs.md#feature-toggles) for the full toggle list.

## 4. Start PostgreSQL (optional but recommended)

```bash
docker run -d --name starterpack-pg \
  -e POSTGRES_PASSWORD=postgres -e POSTGRES_DB=starterpack \
  -p 5433:5432 postgres:16-alpine
```

## 5. Run migrations

```bash
make migrate
```

This applies `apps/api/db/migrations/*.sql` with dbmate and refreshes
`apps/api/db/schema.sql` (which sqlc reads). Re-run `make sqlc` if you changed the
schema.

## 6. Start everything

```bash
make dev
```

| URL | App |
|-----|-----|
| http://localhost:3000 | `app` — dashboard |
| http://localhost:3001 | `web` — marketing site |
| http://localhost:3002 | `api` — Go backend (`/health`, `/api/v1/...`, `/swagger/index.html`) |
| http://localhost:3003 | `email` — React Email preview |
| http://localhost:6006 | `storybook` — design system |

## 7. Verify it works

```bash
# API health
curl http://localhost:3002/health            # {"status":"ok"}

# Create + list a user (validation: username 2–6 chars)
curl -X POST http://localhost:3002/api/v1/users \
  -H 'Content-Type: application/json' \
  -d '{"username":"jdoe","email":"jdoe@example.com"}'
curl http://localhost:3002/api/v1/users
```

Open the dashboard at http://localhost:3000 — it lists users from the API. With
Clerk configured, you'll be redirected to `/sign-in`.

## Common commands

```bash
make help          # list all targets
make build         # build all JS apps + the Go binary
make lint          # ultracite (JS) + go vet
make test          # JS + Go tests
make migrate-new name=create_widgets
make generate      # re-run sqlc + openapi + client codegen
```
