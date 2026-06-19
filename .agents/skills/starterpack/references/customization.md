# Customization

starterpack is modular: integrations are feature-toggled, and the hexagonal
backend makes business logic independent of frameworks. Common changes below.

## Add a domain to the backend (worked recipe)

Mirror the `user` example. Say you're adding `widget`:

1. **Domain** — `internal/domain/widget/`:
   - `widget.go` — entity with `validate:` struct tags for field rules (e.g.
     `validate:"required,min=2,max=100"`). Validation is not done here — the
     service layer handles it via the platform validator.
   - `port.go` — the `Repository` interface (the port the use cases depend on).
   - `errors.go` — transport-agnostic domain errors that wrap the shared sentinels:
     `var ErrWidgetNotFound = fmt.Errorf("widget: %w", domain.ErrNotFound)`. This
     lets `errors.Is(err, domain.ErrNotFound)` work at the HTTP layer.
2. **Application** — `internal/application/widget/service.go`: use cases that take
   the port and orchestrate the domain. This is the **single source of truth for
   validation** — call the platform validator (`internal/platform/validator`) to
   check `validate:` struct tags on the entity before persisting.
3. **Persistence** — add SQL:
   - `db/migrations/<ts>_create_widgets.sql` (`make migrate-new name=create_widgets`),
     then `make migrate`.
   - `db/queries/widgets.sql` with sqlc annotations (`-- name: CreateWidget :one`).
   - `make sqlc` to regenerate; implement the port in
     `internal/adapters/persistence/postgres/widget_repository.go` (wrap the
     generated queries) and optionally `memory/`.
4. **HTTP** — follow the flat naming convention:
   - `internal/adapters/http/widget_handler.go` — thin handlers with swag
     annotations; register routes.
   - `internal/adapters/http/widget_dto.go` — request/response DTOs with only
     `json:` tags (no `binding:` tags — DTOs are pure data shuttles).
   - `response.go` already maps the shared sentinels (`domain.ErrNotFound` → 404,
     `domain.ErrAlreadyExists` → 409, `domain.ErrValidation` → 422), so no
     per-domain error imports are needed.
5. **Wire** — construct the repo + service + handler in `cmd/api/main.go`.
6. **Contract** — `make openapi && make gen-client` to refresh the typed client.

## Add a route/feature to the frontend

- **Route**: drop a file in `apps/app/src/routes/` (or `web`). The TanStack Router
  plugin regenerates `routeTree.gen.ts` on `make dev`/`make build`.
- **Data**: use `useApiClient()` + TanStack Query; the client is typed from the
  OpenAPI spec, so new endpoints appear after `make gen-client`.
- **UI**: import from `@repo/design-system`; add shadcn components with
  `bunx shadcn@latest add <c> -c packages/design-system`.

## Turn a feature on/off

Set (or unset) the relevant env var — see the matrix in `references/setup.md`.
Backend toggles resolve in `apps/api/internal/config/config.go`; frontend toggles
in `apps/<app>/src/features.ts`. To make a NEW integration toggleable:

- **Backend**: add a typed sub-config with an `Enabled()` method; wire the adapter
  only when enabled (provide a no-op/fallback otherwise, like the in-memory repo).
- **Frontend**: add a `VITE_*` flag to `features.ts` and mount its provider
  conditionally.

## Swapping providers

| Concern | Default | To change |
|---------|---------|-----------|
| DB driver/queries | pgx + sqlc | Edit `sqlc.yaml` + `db/queries`; or replace the persistence adapter (GORM/ent) keeping the domain port |
| Migrations | Atlas | Swap the `db-*` / `migrate` Makefile targets for goose/dbmate/golang-migrate; keep `schema.sql` as sqlc's input |
| HTTP framework | Gin | Reimplement `internal/adapters/http` against chi/echo; domain + application are untouched |
| Auth | Clerk | Replace `@repo/auth` + the Go `auth` middleware; keep the `ApiProvider` token interface |
| Payments | Stripe (planned) | Add a `payments` package + a Go adapter + webhook handler under `/api/v1` |
| Analytics | GA + PostHog | Edit `features.ts` and the provider mounts; PostHog `host` points at your self-hosted instance |
| Error tracking | Sentry (toggle) | Add the SDK behind the `Sentry.Enabled()` toggle |

Because the domain depends only on ports, swapping an adapter (DB, HTTP, auth)
never touches business logic — that's the point of the hexagonal layout.

## Adding a new package

Create `packages/<name>/` with a `package.json` named `@repo/<name>`, `type:
module`, an `exports` map, and a `tsconfig.json` extending
`@repo/typescript-config`. Add it as a `workspace:*` dependency where consumed. If
it ships React components imported by subpath, mirror the design-system pattern
(exports + a `paths` mapping in consuming apps so `tsc` resolves source).

## Deployment

- **Frontends**: `make build` → deploy `apps/<app>/dist` to any static host/CDN.
  Set `VITE_*` vars at build time.
- **API**: `make build-api` → ship `apps/api/bin/api` in a container/VM. Provide
  runtime env (`DATABASE_URL`, `CLERK_SECRET_KEY`, …).
- **Migrations**: run `atlas migrate apply` (or `make db-apply`) as a discrete release step (not on app boot).
- **Observability** (Prometheus/Grafana), SSR/SEO for marketing, and hosting
  manifests are intentionally left open for the team to choose.
