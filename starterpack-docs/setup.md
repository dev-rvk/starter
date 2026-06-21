# Setup — from clone to running

This guide takes you from a fresh clone of **starterpack** to a running dev
environment. The stack is a Turborepo with a **Vite + TanStack Router** frontend
(two apps), a **Go hexagonal** backend, a **shadcn/ui** design system, and
feature-toggled integrations. Only **PostgreSQL** is needed for the full
experience — everything else gracefully degrades when its key is absent. By
default, local username/password auth is used; Clerk is optional.

## 1. Prerequisites

| Tool | Version | Notes |
|------|---------|-------|
| [Bun](https://bun.sh) | ≥ 1.3 | Package manager + JS runtime |
| [Go](https://go.dev/dl/) | ≥ 1.24 | Backend (`go` must be on `PATH`) |
| [Docker](https://docs.docker.com/get-docker/) | any | Easiest way to run PostgreSQL locally (optional) |
| [golangci-lint](https://golangci-lint.run/usage/install/) | ≥ 1.57 | Go linter (required for `make lint-go`) |
| GNU Make | any | Orchestrates every command |

Verify:

```bash
bun --version
go version
golangci-lint version
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
- `make tools` — installs the Go CLIs **sqlc**, **swag**, **goimports** (into your
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

**Minimum to boot:** nothing — the API falls back to an in-memory store and uses
local username/password authentication by default. **For the full experience**, set:

- **PostgreSQL** — in `apps/api/.env.local`:
  ```
  DATABASE_URL=postgres://postgres:postgres@localhost:5433/starterpack?sslmode=disable
  ```
- **Clerk** (optional, replaces local auth) — create an app at [clerk.com](https://clerk.com), then:
  - `apps/api/.env.local`: `CLERK_SECRET_KEY=sk_test_...`
  - `apps/app/.env.local`: `VITE_CLERK_PUBLISHABLE_KEY=pk_test_...`

Every other key (Stripe, Sentry, Resend, Google Analytics, PostHog) is optional —
see [docs.md](./docs.md#feature-toggles) for the full toggle list.

## 4. Start local dependencies (Docker Compose)

Backing services are defined in `docker-compose.yml` and managed through the
Makefile:

```bash
make deps-up        # core services (postgres on :5433)
make deps-up-all    # also redis (:6379) + mailpit (:8025) — opt-in profiles
make deps-down      # stop (data kept)
make deps-reset     # stop and delete data volumes
make deps-logs      # tail logs
```

**Local ↔ hosted flow:** each dependency is a container *or* a managed provider —
whichever the env var points at. For local, `make deps-up` and set
`DATABASE_URL` to `postgres://postgres:postgres@localhost:5433/starterpack?sslmode=disable`.
To go hosted, set `DATABASE_URL` to a [Neon](https://neon.tech)/Supabase string
and skip the container. Same pattern for Redis (Upstash) and email (Resend); see
the table in [docs.md](./docs.md#local-dependencies--docker).

## 5. Run migrations

```bash
make db-apply
```

This applies the pre-generated database migrations via Atlas.
If you modify `apps/api/db/schema.sql` in development, generate the corresponding
SQL migrations by running:

```bash
make db-diff name=migration_name
```

Then apply them using `make db-apply` and run `make sqlc` to sync your Go client.

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

Open the dashboard at http://localhost:3000 — it lists users from the API.
You'll be redirected to `/sign-in` (either local auth or Clerk, depending on
your configuration).

## Common commands

```bash
make help          # list all targets
make build         # build all JS apps + the Go binary
make lint          # check JS (ultracite) + go vet
make lint-go       # golangci-lint on the Go backend (apps/api/.golangci.yml)
make lint-fix      # auto-fix JS/TS formatting and linting
make test          # JS + Go tests
make db-diff name=create_widgets # generate an Atlas migration from schema
make db-apply      # apply pending migrations
make db-status     # check migration status
make generate      # re-run sqlc + openapi + client codegen
```

## Go style guide

The Go backend follows the **Uber Go Style Guide**. The rules applied to this
codebase are in `.agents/skills/starterpack/references/good-practices.md`. Key mandates:

- No `init()` in application code — use `New()` constructors and dependency
  injection instead.
- All startup logic lives in `run() error`; `main()` has a single `os.Exit`.
- Every port implementation carries `var _ Interface = (*Impl)(nil)` for
  compile-time checking.
- Handlers depend on `XService` **interfaces**, not `*Service` concrete types.
- Errors use the structured `domain.Error` type with `Kind` for HTTP mapping.
- Imports are in three groups: stdlib / external / internal.

Run `make lint-go` before every PR to enforce these rules automatically.
