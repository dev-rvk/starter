# Setup

## Prerequisites

| Tool | Version | Notes |
|------|---------|-------|
| Bun | ≥ 1.3 | Package manager + JS runtime |
| Go | ≥ 1.24 | Backend; `go` must be on `PATH` |
| Docker | any | Easiest local PostgreSQL (optional) |
| GNU Make | any | Single entrypoint for every command |

Supported OS: macOS, Linux, Windows 11 (WSL2). On WSL with the repo on a Windows
drive (`/mnt/c/...`), enable filesystem metadata so Git/Go/Bun can set
permissions: `sudo mount -o remount,metadata /mnt/c` (until reboot) or add
`[automount]\noptions = "metadata"` to `/etc/wsl.conf` then `wsl --shutdown`. For
best performance, prefer cloning into the Linux filesystem.

## Installation

```bash
make setup
```

Runs, in order: `bun install` → `go mod download` → `make tools` (installs sqlc,
swag into `GOBIN` — ensure `$(go env GOPATH)/bin` is on `PATH`) →
`make generate` (sqlc, OpenAPI, typed client).

Then copy env templates:

```bash
cp apps/api/.env.example apps/api/.env.local
cp apps/app/.env.example apps/app/.env.local
cp apps/web/.env.example apps/web/.env.local
```

## Environment variables & feature toggles

Every integration is enabled **only when its key(s) are present**; otherwise it is
inert and the app still runs. Backend toggles live in
`apps/api/internal/config/config.go` (`Enabled()` per service); frontend toggles
in `apps/<app>/src/features.ts`. Vite only exposes `VITE_`-prefixed variables to
the browser.

| Service | Status | Variables | File(s) |
|---------|--------|-----------|---------|
| **PostgreSQL** | needed (else in-memory) | `DATABASE_URL` | `apps/api/.env.local` |
| **Clerk** (auth) | needed | `CLERK_SECRET_KEY` · `VITE_CLERK_PUBLISHABLE_KEY` | api · app |
| **Stripe** | later | `STRIPE_SECRET_KEY` · `STRIPE_WEBHOOK_SECRET` | api |
| **Sentry** | optional | `SENTRY_DSN` · `VITE_SENTRY_DSN` | api · app |
| **Resend** (email) | later | `RESEND_TOKEN` · `RESEND_FROM` | api · `packages/email` |
| **Google Analytics** | free tier | `VITE_GA_MEASUREMENT_ID` | app · web |
| **PostHog** (self-host) | optional | `VITE_POSTHOG_KEY` · `VITE_POSTHOG_HOST` | app · web |

Removed from the original next-forge (re-add as packages if needed): Arcjet
(security), BetterStack (use Prometheus/Grafana later), BaseHub (CMS).

### Minimum to boot

Nothing — the API falls back to in-memory and auth is bypassed in dev. For the
full experience set `DATABASE_URL` and the Clerk keys above.

### Other env vars

`apps/api/.env.local` also accepts `APP_ENV` (development|production), `PORT`
(default 3002), `LOG_LEVEL`, and `CORS_ORIGINS` (comma-separated; defaults to the
app/web dev URLs). `apps/app/.env.local` has `VITE_API_URL` (default
`http://localhost:3002`); `apps/web/.env.local` has `VITE_APP_URL`.

## Local dependencies (Docker Compose)

`docker-compose.yml` defines local backing services, driven by the Makefile:

```bash
make deps-up        # core (postgres :5433)
make deps-up-all    # also redis (:6379, profile cache) + mailpit (:8025, profile mail)
make deps-down      # stop, keep data
make deps-reset     # stop and delete data volumes
make deps-logs      # tail logs
```

**Local ↔ hosted flow:** each dependency is a local container *or* a managed
provider — whichever its env var points at. Run `make deps-up` for local, or set
the var to a hosted URL (Neon/Supabase for DB, Upstash for Redis, Resend for
email) and skip the container. No code changes; the apps only read env. Add a new
dependency by adding a service (behind a profile if optional) and reading its URL
from env — mirroring the Postgres pattern.

| Service | Local | Profile | Hosted equivalent |
|---------|-------|---------|-------------------|
| PostgreSQL | `postgres` :5433 | core | Neon · Supabase · RDS |
| Redis | `redis` :6379 | `cache` | Upstash |
| Mailpit (SMTP) | `mailpit` :8025 UI | `mail` | Resend (HTTP API) |

## Database setup

Set `DATABASE_URL` in `apps/api/.env.local`, e.g.
`postgres://postgres:postgres@localhost:5433/starterpack?sslmode=disable`, then:

```bash
make db-diff name=create_widgets # generate migration from db/schema.sql changes
make db-apply                    # apply pending migrations to database (Atlas)
make sqlc                        # regenerate type-safe Go if the schema changed
```

The schema is owned by SQL migrations, which Atlas generates automatically by diffing `apps/api/db/schema.sql` (your desired state) against your migration history. **sqlc reads schema.sql** to generate the typed query code.

## Running development

```bash
make dev 
```

| URL | App |
|-----|-----|
| http://localhost:3000 | `app` — dashboard |
| http://localhost:3001 | `web` — marketing |
| http://localhost:3002 | `api` — Go backend (`/health`, `/api/v1/...`, `/swagger/index.html`) |
| http://localhost:3003 | `email` — preview |
| http://localhost:6006 | `storybook` |

Run pieces individually with `make client` / `make server` (or their aliases `make dev-js` / `make dev-api`). Running them separately is recommended for clean log visibility from the Go API.

## Verify

```bash
curl http://localhost:3002/health        # {"status":"ok"}
curl -X POST http://localhost:3002/api/v1/users \
  -H 'Content-Type: application/json' \
  -d '{"username":"jdoe","email":"jdoe@example.com"}'   # 201
curl http://localhost:3002/api/v1/users
```

The dashboard at http://localhost:3000 lists those users. Invalid input (e.g. a
username outside 2–6 chars) returns `422` from the validator.

## Validation model

Validation is a **single source of truth** in the backend:
- **Backend**: domain entities carry `validate:` struct tags (e.g.
  `validate:"required,min=2,max=6"`). Validation is performed **only** in the
  application/service layer using the platform validator
  (`internal/platform/validator`). HTTP DTOs are pure data shuttles with only
  `json:` tags — no `binding:` validation tags. Validation errors are returned
  as `domain.ErrValidation` (→ 422).
- **Frontend (UX)**: zod schemas in the design-system auth forms mirror the same
  rules for instant feedback.

## Notes / gotchas

- `make typecheck` (or `bun run typecheck`) needs the generated TanStack route
  trees; run `make dev` or `make build` once first (Vite generates them).
- swag emits **Swagger 2.0**; `@repo/api-client` converts it to OpenAPI 3 via
  `swagger2openapi` before `openapi-typescript`. Just run `make gen-client`.
- The Go API runs migrations as a separate step (via Atlas in CI/CD), not on boot — the
  robust production pattern.
