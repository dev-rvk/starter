# Architecture

## Monorepo Structure

Turborepo manages the JS/TS workspaces (bun). The Go backend (`apps/api`) is its
own Go module, orchestrated alongside turbo by the root `Makefile`.

```
starterpack/
├── apps/
│   ├── app/          # Dashboard — Vite + TanStack Router (port 3000)
│   ├── web/          # Marketing site — Vite + TanStack Router (port 3001)
│   ├── api/          # Backend — Go, Gin, hexagonal (port 3002)
│   ├── email/        # React Email preview (port 3003)
│   └── storybook/    # Design system workshop — Storybook + Vite (port 6006)
├── packages/
│   ├── api-client/        # Typed TS client generated from the API's OpenAPI spec
│   ├── auth/              # Clerk wrapper (mounts only when keyed)
│   ├── design-system/     # shadcn/ui + Tailwind v4 tokens + auth forms
│   ├── email/             # Resend client + React Email templates
│   └── typescript-config/ # Shared tsconfig presets
├── starterpack-docs/      # setup.md (clone→run) + docs.md (changes/decisions)
├── .agents/skills/        # Project skills (this one: starterpack)
├── docker-compose.yml     # Local backing services (postgres core; redis/mailpit via profiles)
├── Makefile               # Single entrypoint (wraps bun/turbo + Go toolchain + compose)
├── turbo.json
└── package.json
```

## Apps

### app (Port 3000)
The authenticated dashboard. Vite + React with **TanStack Router** (type-safe
file-based routing) and **TanStack Query**. Auth via `@repo/auth` (Clerk);
sign-in / sign-up / forgot-password routes render the design system's auth forms
wired to Clerk headless hooks. Data is fetched from the Go API through the typed
`@repo/api-client`. `src/features.ts` is the central feature-toggle map.

### web (Port 3001)
The marketing site (home + pricing). Vite + TanStack Router + the design system.
Client-rendered SPA today; SSR/SEO is a deliberate future step.

### api (Port 3002)
The Go backend — hexagonal (ports & adapters), Gin, zerolog. Serves REST under
`/api/v1`, a health check at `/health`, and Swagger UI at `/swagger/index.html`.
See "Hexagonal layout" below.

### email (Port 3003)
React Email preview server. Templates live in `@repo/email` (`packages/email/templates`).

### storybook (Port 6006)
Storybook on the **Vite** builder (`@storybook/react-vite` + `@tailwindcss/vite`)
for developing design-system components in light/dark.

## Hexagonal layout (apps/api)

Dependencies point inward — domain knows nothing of HTTP or SQL.

```
apps/api/
├── cmd/api/main.go                       # composition root: run() error + single os.Exit in main()
├── internal/
│   ├── config/config.go                  # env → typed Config + Enabled() feature toggles
│   ├── domain/
│   │   ├── errors.go                     # structured domain.Error type + KindUnknown/NotFound/AlreadyExists/Validation
│   │   ├── user/                         # entity (validate: struct tags), Repository PORT, XService interface
│   │   │   ├── user.go                   #   entity with validate: struct tags
│   │   │   ├── port.go                   #   Repository interface (the port)
│   │   │   └── errors.go                 #   domain.NotFound("user") / domain.ValidationError(...)
│   │   ├── todo/                         # (same pattern per resource)
│   │   └── account/                      # auth credentials (email + password hash), separate from user
│   ├── application/
│   │   ├── user/
│   │   │   └── service.go               # use cases + UserService interface + var _ UserService = (*Service)(nil)
│   │   ├── todo/
│   │   │   └── service.go               # use cases + TodoService interface + var _ TodoService = (*Service)(nil)
│   │   └── auth/                         # local-auth use cases: orchestrates account + user, issues JWTs (bcrypt)
│   │       └── service.go               #   var _ AuthService = (*Service)(nil)
│   ├── adapters/
│   │   ├── http/                         # flat: {resource}_handler.go, {resource}_dto.go, response.go
│   │   │   ├── user_handler.go           #   takes UserService interface (not *userapp.Service)
│   │   │   ├── user_dto.go               #   pure data shuttle (json tags only, no binding tags)
│   │   │   ├── todo_handler.go           #   (same pattern)
│   │   │   ├── todo_dto.go
│   │   │   ├── auth_handler.go           #   /auth/register, /auth/login (public), /auth/me (protected)
│   │   │   ├── auth_dto.go
│   │   │   ├── response.go               #   maps domain.Error.Kind → HTTP status; KindUnknown → 500
│   │   │   └── middleware/               #   logger.go (zerolog), cors.go, auth.go (Clerk), local_auth.go (local JWT)
│   │   └── persistence/
│   │       ├── postgres/                 # pgx pool + sqlc-generated queries
│   │       │   ├── user_repository.go    #   var _ userdomain.Repository = (*UserRepository)(nil)
│   │       │   ├── todo_repository.go    #   var _ tododomain.Repository = (*TodoRepository)(nil)
│   │       │   ├── account_repository.go #   var _ accountdomain.Repository = (*AccountRepository)(nil)
│   │       │   └── sqlc/                 #   GENERATED — do not edit by hand
│   │       └── memory/                   # in-memory repos (no-DB fallback): user, todo, account
│   └── platform/
│       ├── logger/                       # zerolog setup (logger.New)
│       ├── jwtutil/                      # JWTManager: Sign/Verify HMAC tokens for local auth (jwtutil.New)
│       └── validator/                    # Validator struct with New() constructor (no init(), no global var)
├── db/
│   ├── migrations/                       # Atlas versioned migrations
│   ├── queries/                          # sqlc input queries
│   └── schema.sql                        # desired schema state -> sqlc input
├── docs/                                 # GENERATED OpenAPI spec (swag)
├── atlas.hcl                             # Atlas env config (src = db/schema.sql, dir = db/migrations)
├── sqlc.yaml
└── go.mod
```

**Validation**: happens **only** in the application/service layer (single source of
truth). Domain entities carry `validate:` struct tags; the `platform/validator`
`Validator` struct (injected via `New()`) checks them. HTTP DTOs are pure data
shuttles with only `json:` tags — no `binding:` validation tags.

**Error handling**: the shared `domain.Error` struct in `internal/domain/errors.go`
carries a `Kind` (KindUnknown=0, KindNotFound=1, KindAlreadyExists=2,
KindValidation=3) and wraps the corresponding sentinel so both `errors.Is()` and
`errors.As()` work. Use the constructors: `domain.NotFound("user")`,
`domain.AlreadyExists("user")`, `domain.ValidationError("user","Email","reason")`.
`response.go` maps `domErr.Kind` to HTTP status; `KindUnknown` → 500.

**Interface compliance**: every adapter carries `var _ Port = (*Impl)(nil)` at the
top of the file. Handlers depend on the `XService` interface defined alongside the
service, not the concrete `*Service` type.

**No `init()` in application code**: `platform/validator` exposes `New()` and is
instantiated and injected in `cmd/api/main.go`.

**Exit Once**: `main()` calls `os.Exit(1)` at most once. All startup, wiring, and
server-lifecycle logic lives in `run() error`. Errors surface via `return`, not
`log.Fatal`.

**Goroutine lifecycle**: the HTTP server goroutine sends on a buffered error channel;
`run()` selects on that channel and the OS signal channel, ensuring the goroutine's
result is always observed before shutdown proceeds.

**File naming convention**: handlers use a flat `internal/adapters/http/` package
with `{resource}_handler.go` and `{resource}_dto.go` (e.g. `user_handler.go`,
`user_dto.go`, `todo_handler.go`, `todo_dto.go`).

**Authentication (dual mode)**: auth is always on, selected by config in
`NewRouter` (`internal/adapters/http/server.go`):

- **Clerk mode** — when `CLERK_SECRET_KEY` is set, all `/api/v1` routes are guarded
  by `middleware.ClerkAuth()`. No local-auth code is wired (`AuthHandler`/`JWTManager`
  stay nil).
- **Local mode** — the default fallback. `cmd/api/main.go` builds a `jwtutil.JWTManager`
  (24h HMAC tokens) and an `authapp.Service` over the `account` + `user` repos. Public
  `/api/v1/auth/register` and `/auth/login` are mounted first; then `middleware.LocalAuth`
  guards the rest (including `/auth/me`). Passwords are bcrypt-hashed; the `account`
  domain (credentials) is kept separate from the `user` domain (profile), sharing one ID.

If neither Clerk is configured nor an `AuthHandler` is wired, `/api/v1` is left
**unprotected** and the server logs a warning — fine for a fresh clone, not for prod.

**Module path**: `github.com/starterpack/api`. **Adding a domain**: see
`references/customization.md` and the checklist in `references/good-practices.md`.

## Package Naming

All packages use `@repo/<name>` and are imported by subpath:

```typescript
import { createApiClient } from '@repo/api-client';
import { AuthProvider, useAuth } from '@repo/auth';
import { Button } from '@repo/design-system/components/ui/button';
import { LoginForm } from '@repo/design-system/components/auth/login-form';
import '@repo/design-system/styles/globals.css';
```

> Consuming apps map `@repo/design-system/*` to the package `src` in their
> `tsconfig` `paths` so `tsc` resolves the source subpaths (Vite resolves them
> via the package `exports` map at bundle time).

## Turborepo pipeline (`turbo.json`)

| Task | Dependencies | Outputs | Cached | Persistent |
|------|--------------|---------|--------|------------|
| `build` | `^build` | `dist`, `storybook-static`, `.react-email` | Yes | No |
| `typecheck` | `^build` | — | Yes | No |
| `test` | `^test` | — | Yes | No |
| `dev` | — | — | No | Yes |
| `clean` | — | — | No | No |

Global dependencies include `**/.env.*local`. The Go API is **not** a turbo task;
the Makefile runs it alongside `turbo dev`.

## Makefile pipeline (root entrypoint)

`make help` lists everything. Key targets:

| Target | What it does |
|--------|--------------|
| `setup` | `install` + `go-deps` + `tools` + `generate` |
| `tools` | `go install` sqlc, swag |
| `generate` | `sqlc` + `openapi` + `gen-client` |
| `dev` | run Go API + JS apps concurrently (`make -j2 client server`) (bootstrapping) |
| `client` | run JS apps via turbo (TUI mode) |
| `server` | run Go API backend (clean stdout logs) |
| `deps-up` / `deps-up-all` / `deps-down` / `deps-reset` / `deps-logs` | Docker Compose local services |
| `db-diff` / `db-apply` / `db-status` / `db-reset` / `migrate` | Atlas |
| `build` | `build-js` (turbo) + `build-api` (go build) |
| `lint` / `test` / `clean` | JS + Go together |

## Build outputs

- `apps/<app>/dist/` — Vite static build (deployable to any static host)
- `apps/storybook/storybook-static/` — Storybook export
- `apps/api/bin/api` — compiled Go binary
- `apps/api/docs/` — generated OpenAPI (`swagger.json`/`swagger.yaml`/`docs.go`)
- `packages/api-client/src/schema.d.ts` — generated TS types

## Generated files (do not hand-edit)

- `apps/api/internal/adapters/persistence/postgres/sqlc/**` (sqlc)
- `apps/api/docs/**` (swag)
- `packages/api-client/src/schema.d.ts` (openapi-typescript)
- `apps/<app>/src/routeTree.gen.ts` (TanStack Router plugin, gitignored)

> `apps/api/db/schema.sql` is **hand-authored** — it is the desired-state input
> that Atlas diffs to generate migrations and that sqlc reads to generate Go. Edit
> it directly; do not treat it as generated. The generated migration files land in
> `apps/api/db/migrations/` (tracked by `atlas.sum`).
