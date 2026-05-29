# Apps & Packages

Apps are in `/apps/`; shared packages in `/packages/` (imported as `@repo/<name>`).

## Backend — `apps/api` (Go)

**Stack**: Go, Gin, zerolog, pgx + sqlc, dbmate, Clerk SDK, go-playground/validator,
swag (OpenAPI). **Architecture**: hexagonal (ports & adapters).

**Layout** (see `references/architecture.md` for the full tree):
- `internal/domain/<x>/` — entities, value objects (validate in constructors),
  and the `Repository` **port** (interface).
- `internal/application/<x>/` — use cases; depend only on ports.
- `internal/adapters/http/` — Gin handlers (thin), DTOs, response mapping,
  middleware (`logger` zerolog, `cors`, `auth` Clerk JWT). Handlers carry swag
  annotations that generate the OpenAPI spec.
- `internal/adapters/persistence/{postgres,memory}/` — two implementations of the
  repository port. `postgres` wraps sqlc-generated queries over a pgx pool;
  `memory` is the no-DB fallback.
- `internal/config/config.go` — env → typed `Config`; `Enabled()` per service.

**Endpoints**: `GET /health`, `GET /swagger/*`, and `/api/v1/users`
(POST/GET/GET:id) as the worked example. Routes under `/api/v1` are protected by
Clerk middleware when `CLERK_SECRET_KEY` is set.

**DB workflow**: edit a dbmate migration in `db/migrations/`, `make migrate`
(applies + dumps `db/schema.sql`), `make sqlc` (regenerates typed Go from
`db/queries/` against `schema.sql`).

## Frontend apps

### `apps/app` (dashboard)
Vite + TanStack Router + TanStack Query. Routes in `src/routes/`: `__root`,
`index` (guarded dashboard + users query), `sign-in`, `sign-up`,
`forgot-password`. `src/lib/api.tsx` provides `ApiProvider`/`useApiClient`
(token-aware when Clerk is on). `src/features.ts` is the toggle map.

### `apps/web` (marketing)
Vite + TanStack Router. Routes: `index` (landing) and `pricing`. Uses the design
system; links to the app via `VITE_APP_URL`.

## `@repo/design-system`

**Provider**: shadcn/ui (Radix + Tailwind v4), `new-york` style, `neutral` base.

**Contents**:
- `src/components/ui/*` — ~56 shadcn components.
- `src/styles/globals.css` — Tailwind v4 `@theme` with **hex design tokens**:
  shadcn semantic tokens + sidebar/chart, a **brand** scale (`--color-brand-*`),
  neutrals, semantic colors (success/warning/danger/info), and **fonts**
  (`--font-sans` Inter, `--font-display` Cal Sans, `--font-mono` JetBrains Mono).
  Dark mode overrides the same `--color-*` names under `.dark` in `@layer base`.
  `@source` globs make the one stylesheet scan the package and `apps/**`
  (Tailwind v4 ignores `node_modules`). Fonts load via Google Fonts `<link>`s in
  each app's `index.html` / Storybook `preview-head.html` (add Cal Sans manually).
- `src/lib/utils.ts` — `cn` helper.
- `src/components/theme-provider.tsx` — next-themes (works in Vite SPAs);
  `mode-toggle.tsx` — light/dark/system switch.
- `src/components/auth/*` — **bespoke** `LoginForm`, `SignUpForm`,
  `ForgotPasswordForm` (presentational, props-driven, zod + react-hook-form).

**Imports** use package subpaths (e.g.
`@repo/design-system/components/ui/button`), resolved via the package `exports`
map. Add components with `bunx shadcn@latest add <c> -c packages/design-system`
(`components.json` is configured so generated imports use `@repo/design-system/*`).

## `@repo/auth`

**Provider**: Clerk. `AuthProvider` mounts `ClerkProvider` **only** when
`VITE_CLERK_PUBLISHABLE_KEY` is set (graceful when the toggle is off);
`isAuthEnabled()` / `useIsAuthEnabled()` report status. Re-exports Clerk hooks
and components (`useAuth`, `useSignIn`, `useSignUp`, `SignedIn`, `UserButton`, …).
The design-system auth forms are wired to these headless hooks in the app's auth
routes.

## `@repo/api-client`

**Generated** typed client. `bun run generate` (or `make gen-client`) runs
`swag → swagger2openapi → openapi-typescript`, producing `src/schema.d.ts`.
`createApiClient({ baseUrl, getToken? })` wraps `openapi-fetch` and injects a
Clerk bearer token via middleware. Pairs with TanStack Query.

```typescript
import { createApiClient } from '@repo/api-client';
const api = createApiClient({ baseUrl: 'http://localhost:3002/api/v1' });
const { data, error } = await api.GET('/users');   // fully typed
```

## `@repo/email`

**Provider**: Resend + React Email. Templates in `templates/` (rendered by the
`email` app's preview). `keys.ts` gates the Resend token; sending is inert until
`RESEND_TOKEN` is set.

## `@repo/typescript-config`

Shared tsconfig presets: `base.json`, `react-library.json` (and the legacy
`nextjs.json`). Packages/apps extend these.
