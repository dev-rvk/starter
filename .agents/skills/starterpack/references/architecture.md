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
├── skills/                # Project skills (next-forge reference + this one)
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
├── cmd/api/main.go                       # composition root
├── internal/
│   ├── config/config.go                  # env → typed Config + feature toggles
│   ├── domain/
│   │   ├── errors.go                      # shared error sentinels (ErrNotFound, ErrAlreadyExists, ErrValidation)
│   │   ├── user/                          # entity (validate: struct tags), Repository PORT
│   │   │   ├── user.go                    #   entity with validate: struct tags
│   │   │   ├── port.go                    #   Repository interface (the port)
│   │   │   └── errors.go                  #   wraps shared sentinels: fmt.Errorf("user: %w", domain.ErrNotFound)
│   │   └── todo/                          # (same pattern per resource)
│   ├── application/user/service.go        # use cases — single source of truth for validation (platform validator)
│   ├── adapters/
│   │   ├── http/                          # flat: {resource}_handler.go, {resource}_dto.go, response.go
│   │   │   ├── user_handler.go            #   thin Gin handler + swag annotations
│   │   │   ├── user_dto.go                #   pure data shuttle (json tags only, no binding tags)
│   │   │   ├── todo_handler.go            #   (same pattern per resource)
│   │   │   ├── todo_dto.go
│   │   │   ├── response.go                #   maps shared domain sentinels → HTTP status
│   │   │   └── middleware/                #   logger (zerolog), cors, clerk auth
│   │   └── persistence/
│   │       ├── postgres/                  # pgx pool + sqlc-generated queries
│   │       │   └── sqlc/                   #   GENERATED — do not edit by hand
│   │       └── memory/                    # in-memory repo (no-DB fallback)
│   └── platform/
│       ├── logger/                         # zerolog setup
│       └── validator/                      # shared go-playground/validator instance
├── db/
│   ├── migrations/                        # Atlas versioned migrations
│   ├── queries/                           # sqlc input queries
│   └── schema.sql                         # desired schema state -> sqlc input
├── docs/                                  # GENERATED OpenAPI spec (swag)
├── sqlc.yaml
└── go.mod
```

**Validation**: happens **only** in the application/service layer (single source of
truth). Domain entities carry `validate:` struct tags; the platform validator
(`internal/platform/validator`) checks them. HTTP DTOs are pure data shuttles with
only `json:` tags — no `binding:` validation tags.

**Error handling**: shared error sentinels in `internal/domain/errors.go`
(`domain.ErrNotFound`, `domain.ErrAlreadyExists`, `domain.ErrValidation`). Each
domain wraps these: `fmt.Errorf("user: %w", domain.ErrNotFound)` so `errors.Is()`
matches both the specific wrapped error and the shared sentinel. `response.go`
imports only the shared `internal/domain` package (not each individual domain) and
maps errors by category: `domain.ErrNotFound` → 404, `domain.ErrAlreadyExists` →
409, `domain.ErrValidation` → 422.

**File naming convention**: handlers use a flat `internal/adapters/http/` package
with `{resource}_handler.go` and `{resource}_dto.go` (e.g. `user_handler.go`,
`user_dto.go`, `todo_handler.go`, `todo_dto.go`).

**Module path**: `github.com/starterpack/api`. **Adding a domain**: see
`references/customization.md`.

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
- `apps/api/db/schema.sql` (desired schema state)
- `packages/api-client/src/schema.d.ts` (openapi-typescript)
- `apps/<app>/src/routeTree.gen.ts` (TanStack Router plugin, gitignored)
