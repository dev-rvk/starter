# CI/CD & Cloud Deployment Reference

> Authoritative reference for deploying the starterpack monorepo.
> **Stack**: Docker Hub (images) · GCP Cloud Run (API) · Cloudflare Pages (frontends) · Neon Postgres · GitHub Actions CI/CD.

---

## 1. Branch Strategy

```
feature/* ──PR──▶ main (trunk) ──PR──▶ release/prod
                   │                        │
              staging deploy           prod deploy
              (on push)                (on push)
```

| Branch | Purpose | Deploy trigger |
|--------|---------|---------------|
| `main` | Trunk development branch | Push → staging |
| `release/prod` | Production mirror | Push → production |
| `feature/*` | Feature branches | PR into `main` (CI only) |

### Flow

1. **Feature work** — branch off `main`, open PR back to `main`. CI runs on every PR.
2. **Staging deploy** — merge PR into `main`. Staging pipeline runs automatically.
3. **Production deploy** — open PR from `main` → `release/prod`. Merge triggers prod pipeline with manual approval gate.
4. **Hotfixes** — branch off `release/prod`, PR back to `release/prod`. After merge, cherry-pick the commit into `main` to keep trunk in sync.

> The `release/prod` branch must be created manually before the first production deploy:
> ```bash
> git checkout main
> git checkout -b release/prod
> git push -u origin release/prod
> ```

---

## 2. What Deploys Where

| App | Build Artifact | Target | Docker | Notes |
|-----|---------------|--------|--------|-------|
| `apps/app` | `dist/` (Vite) | Cloudflare Pages | No | Main SaaS frontend |
| `apps/web` | `dist/` (Vite) | Cloudflare Pages | No | Marketing / landing site |
| `apps/api` | Docker image | Docker Hub → Google Cloud Run | Yes | Go API server |
| `apps/storybook` | `storybook-static/` | Cloudflare Pages (optional) | No | Design system preview |
| `apps/email` | dev preview only | **Not deployed** | — | React Email templates, preview-only |

---

## 3. Deploy Order (Critical)

```
Migrations → API → Frontends
```

**Never reverse this order.** The reason is dependency flow:

1. **Migrations first** — the API may reference new columns/tables added by the migration. Deploying the API before migrating would cause runtime errors.
2. **API second** — the API serves the OpenAPI spec from which `@repo/api-client` types are generated. If the API contract changes, frontends built against the old spec will still work (backward-compatible changes) or intentionally break (breaking changes caught in CI).
3. **Frontends last** — they are built with the latest `@repo/api-client` types. Deploying them before the API could cause calls to endpoints that don't exist yet.

In GitHub Actions this is enforced via `needs:` dependencies between jobs.

---

## 4. Makefile Targets for CI

These targets are defined in the **root `Makefile`** and are the canonical interface for CI jobs. CI workflows call `make <target>` — they never run raw commands directly.

```makefile
# ── Testing ──────────────────────────────────────────────
test-api:                              ## Run Go API tests with race detector
	cd apps/api && go test ./... -race -count=1

test-js:                               ## Run all JS/TS tests via Turborepo
	bun run turbo run test

test-js-affected:                      ## Run JS/TS tests only for packages affected since main
	bun run turbo run test --filter='[origin/main]'

# ── Linting ──────────────────────────────────────────────
lint-api:                              ## Run golangci-lint on the API
	cd apps/api && golangci-lint run ./...

lint-js:                               ## Run checks/linter across all JS/TS packages
	bun run check

# ── Type checking ────────────────────────────────────────
typecheck:                             ## Run TypeScript type-checking across all packages (excluding storybook)
	bun run turbo run typecheck --filter='!storybook'

# ── Code generation ──────────────────────────────────────
generate-check:                        ## Run all generators and fail if working tree is dirty
	make generate && git diff --exit-code -- apps/api/internal/sqlc apps/api/api/openapi.gen.go packages/api-client/src/generated

# ── Docker ───────────────────────────────────────────────
docker-build:                          ## Build the API Docker image
	docker build -t starterpack-api:$(TAG) -f apps/api/Dockerfile apps/api

docker-push:                           ## Tag and push the API image to the registry
	docker tag starterpack-api:$(TAG) $(REGISTRY):$(TAG)
	docker push $(REGISTRY):$(TAG)

# ── Database ─────────────────────────────────────────────
db-migrate-prod:                       ## Apply migrations to a remote database
	npx @ariga/atlas@0.37.0 migrate apply \
		--dir file://apps/api/db/migrations \
		--url $(DATABASE_URL)
```

> **Why `npx @ariga/atlas@0.37.0`?** The local Makefile already uses this exact invocation for `db-migrate` and `db-diff`. Pinning the same version in CI prevents version drift between local and deployed schemas. Do **not** install a standalone Atlas CLI or use a different version.

---

## 5. Dockerfile

Reference: `apps/api/Dockerfile`

The API uses a **multi-stage build**:

```
Stage 1: golang:1.26-alpine  (build)
Stage 2: gcr.io/distroless/static:nonroot  (runtime)
```

Key build flags:

```dockerfile
ENV CGO_ENABLED=0
RUN go build -ldflags="-s -w" -o /app ./cmd/api
```

- `CGO_ENABLED=0` — fully static binary, no libc dependency.
- `-ldflags="-s -w"` — strip debug info and DWARF symbols.
- Final image is **~10–15 MB**.
- Runs as non-root user via the `nonroot` distroless variant.
- Exposes port `8080` (Cloud Run default).

---

## 6. GitHub Actions Workflows

Three workflow files are needed. The YAML below is reference material — the actual files should be created at `.github/workflows/`.

### 6.1 `ci.yml` — Pull Request Checks

Runs on every PR targeting `main` or `release/prod`. Validates code generation, Go quality, and JS/TS quality in parallel.

```yaml
name: CI

on:
  pull_request:
    branches: [main, release/prod]

concurrency:
  group: ci-${{ github.head_ref }}
  cancel-in-progress: true

jobs:
  generate-check:
    name: Generate Check
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: "1.26"
      - uses: oven-sh/setup-bun@v2
      - run: bun install --frozen-lockfile
      - run: make tools
      - run: make generate-check

  api:
    name: API (Go)
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: "1.26"
      - uses: golangci/golangci-lint-action@v6
        with:
          version: latest
          working-directory: apps/api
      - run: make test-api

  js:
    name: JS/TS
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 2          # Turbo needs git history for --filter
      - uses: oven-sh/setup-bun@v2
      - uses: actions/cache@v4
        with:
          path: node_modules/.cache/turbo
          key: turbo-${{ runner.os }}-${{ hashFiles('bun.lock') }}
          restore-keys: turbo-${{ runner.os }}-
      - run: bun install --frozen-lockfile
      - run: make lint-js
      - run: make typecheck
      - run: make test-js-affected
```

### 6.2 `deploy-staging.yml` — Staging Deploy

Runs on push to `main` (i.e., after a PR merge). Deploys to staging environment.

```yaml
name: Deploy Staging

on:
  push:
    branches: [main]

concurrency:
  group: deploy-staging
  cancel-in-progress: false        # Never cancel a mid-flight deploy

env:
  GCP_REGION: us-central1
  IMAGE: ${{ secrets.DOCKERHUB_USERNAME }}/starterpack-api

jobs:
  # ── Test ─────────────────────────────────────────────
  test-api:
    name: Test API
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: "1.26"
      - run: make test-api

  lint-api:
    name: Lint API
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: "1.26"
      - uses: golangci/golangci-lint-action@v6
        with:
          version: latest
          working-directory: apps/api

  test-js:
    name: Test JS/TS
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 2
      - uses: oven-sh/setup-bun@v2
      - uses: actions/cache@v4
        with:
          path: node_modules/.cache/turbo
          key: turbo-${{ runner.os }}-${{ hashFiles('bun.lock') }}
          restore-keys: turbo-${{ runner.os }}-
      - run: bun install --frozen-lockfile
      - run: make lint-js
      - run: make typecheck
      - run: make test-js

  # ── Build + Push Docker ─────────────────────────────
  build-push:
    name: Build & Push Docker Image
    needs: [test-api, lint-api, test-js]
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: docker/login-action@v3
        with:
          username: ${{ secrets.DOCKERHUB_USERNAME }}
          password: ${{ secrets.DOCKERHUB_TOKEN }}
      - run: make docker-build TAG=${{ github.sha }}
      - run: make docker-push TAG=${{ github.sha }} REGISTRY=${{ env.IMAGE }}

  # ── Migrate ─────────────────────────────────────────
  migrate:
    name: DB Migrations (Staging)
    needs: [build-push]
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - run: make db-migrate-prod
        env:
          DATABASE_URL: ${{ secrets.STAGING_DATABASE_URL }}

  # ── Deploy API ──────────────────────────────────────
  deploy-api:
    name: Deploy API to Cloud Run (Staging)
    needs: [migrate]
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: google-github-actions/auth@v2
        with:
          credentials_json: ${{ secrets.GCP_SA_KEY }}
      - uses: google-github-actions/deploy-cloudrun@v2
        with:
          service: starterpack-api-staging
          region: ${{ env.GCP_REGION }}
          image: docker.io/${{ env.IMAGE }}:${{ github.sha }}
          env_vars: |
            GIN_MODE=release
            APP_ENV=staging
          secrets: |
            CLERK_SECRET_KEY=STAGING_CLERK_SECRET_KEY:latest
            JWT_SECRET=STAGING_JWT_SECRET:latest
            DATABASE_URL=STAGING_DATABASE_URL:latest

  # ── Deploy Frontends ────────────────────────────────
  deploy-app:
    name: Deploy app (Staging)
    needs: [deploy-api]
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: oven-sh/setup-bun@v2
      - run: bun install --frozen-lockfile
      - run: bun run turbo run build --filter=app
        env:
          VITE_CLERK_PUBLISHABLE_KEY: ${{ secrets.STAGING_VITE_CLERK_PUBLISHABLE_KEY }}
          VITE_API_URL: ${{ secrets.STAGING_API_URL }}
      - uses: cloudflare/wrangler-action@v3
        with:
          apiToken: ${{ secrets.CF_API_TOKEN }}
          accountId: ${{ secrets.CF_ACCOUNT_ID }}
          command: pages deploy apps/app/dist --project-name=starterpack-app --branch=staging

  deploy-web:
    name: Deploy web (Staging)
    needs: [deploy-api]
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: oven-sh/setup-bun@v2
      - run: bun install --frozen-lockfile
      - run: bun run turbo run build --filter=web
        env:
          VITE_APP_URL: ${{ secrets.STAGING_APP_URL }}
      - uses: cloudflare/wrangler-action@v3
        with:
          apiToken: ${{ secrets.CF_API_TOKEN }}
          accountId: ${{ secrets.CF_ACCOUNT_ID }}
          command: pages deploy apps/web/dist --project-name=starterpack-web --branch=staging
```

### 6.3 `deploy-prod.yml` — Production Deploy

Runs on push to `release/prod`. Includes a manual approval gate via GitHub Environments.

```yaml
name: Deploy Production

on:
  push:
    branches: [release/prod]

concurrency:
  group: deploy-prod
  cancel-in-progress: false        # NEVER cancel a production deploy mid-flight

env:
  GCP_REGION: us-central1
  IMAGE: ${{ secrets.DOCKERHUB_USERNAME }}/starterpack-api

jobs:
  # ── Test ─────────────────────────────────────────────
  test-api:
    name: Test API
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: "1.26"
      - run: make test-api

  lint-api:
    name: Lint API
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: "1.26"
      - uses: golangci/golangci-lint-action@v6
        with:
          version: latest
          working-directory: apps/api

  test-js:
    name: Test JS/TS
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 2
      - uses: oven-sh/setup-bun@v2
      - uses: actions/cache@v4
        with:
          path: node_modules/.cache/turbo
          key: turbo-${{ runner.os }}-${{ hashFiles('bun.lock') }}
          restore-keys: turbo-${{ runner.os }}-
      - run: bun install --frozen-lockfile
      - run: make lint-js
      - run: make typecheck
      - run: make test-js

  # ── Build + Push Docker ─────────────────────────────
  build-push:
    name: Build & Push Docker Image
    needs: [test-api, lint-api, test-js]
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: docker/login-action@v3
        with:
          username: ${{ secrets.DOCKERHUB_USERNAME }}
          password: ${{ secrets.DOCKERHUB_TOKEN }}
      - run: make docker-build TAG=${{ github.sha }}
      - run: make docker-push TAG=${{ github.sha }} REGISTRY=${{ env.IMAGE }}

  # ── Approval Gate ───────────────────────────────────
  approve:
    name: Production Approval
    needs: [build-push]
    runs-on: ubuntu-latest
    environment: production           # Requires manual approval in GitHub
    steps:
      - run: echo "Production deploy approved"

  # ── Migrate ─────────────────────────────────────────
  migrate:
    name: DB Migrations (Production)
    needs: [approve]
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - run: make db-migrate-prod
        env:
          DATABASE_URL: ${{ secrets.PROD_DATABASE_URL }}

  # ── Deploy API ──────────────────────────────────────
  deploy-api:
    name: Deploy API to Cloud Run (Production)
    needs: [migrate]
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: google-github-actions/auth@v2
        with:
          credentials_json: ${{ secrets.GCP_SA_KEY }}
      - uses: google-github-actions/deploy-cloudrun@v2
        with:
          service: starterpack-api
          region: ${{ env.GCP_REGION }}
          image: docker.io/${{ env.IMAGE }}:${{ github.sha }}
          env_vars: |
            GIN_MODE=release
            APP_ENV=production
          secrets: |
            CLERK_SECRET_KEY=PROD_CLERK_SECRET_KEY:latest
            JWT_SECRET=PROD_JWT_SECRET:latest
            DATABASE_URL=PROD_DATABASE_URL:latest

  # ── Deploy Frontends ────────────────────────────────
  deploy-app:
    name: Deploy app (Production)
    needs: [deploy-api]
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: oven-sh/setup-bun@v2
      - run: bun install --frozen-lockfile
      - run: bun run turbo run build --filter=app
        env:
          VITE_CLERK_PUBLISHABLE_KEY: ${{ secrets.PROD_VITE_CLERK_PUBLISHABLE_KEY }}
          VITE_API_URL: ${{ secrets.PROD_API_URL }}
      - uses: cloudflare/wrangler-action@v3
        with:
          apiToken: ${{ secrets.CF_API_TOKEN }}
          accountId: ${{ secrets.CF_ACCOUNT_ID }}
          command: pages deploy apps/app/dist --project-name=starterpack-app --branch=production

  deploy-web:
    name: Deploy web (Production)
    needs: [deploy-api]
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: oven-sh/setup-bun@v2
      - run: bun install --frozen-lockfile
      - run: bun run turbo run build --filter=web
        env:
          VITE_APP_URL: ${{ secrets.PROD_APP_URL }}
      - uses: cloudflare/wrangler-action@v3
        with:
          apiToken: ${{ secrets.CF_API_TOKEN }}
          accountId: ${{ secrets.CF_ACCOUNT_ID }}
          command: pages deploy apps/web/dist --project-name=starterpack-web --branch=production
```

---

## 7. Turborepo + Go Split

Turborepo manages **only** JS/TS tasks (`build`, `lint`, `test`, `typecheck`, `dev`). It does not orchestrate Go.

Go tasks are always driven by the **root Makefile** (`test-api`, `lint-api`, `docker-build`, etc.).

In CI, these are **independent job trees**:

```
ci.yml
├── generate-check     (Go + JS — runs make generate-check)
├── api                (Go only — golangci-lint + make test-api)
└── js                 (JS only — turbo lint, typecheck, test)
```

The deploy workflows combine both trees but the split remains: Go jobs use `setup-go` + Makefile, JS jobs use `setup-bun` + turbo.

### Why not turbo for Go?

- Turbo's cache hashing is built around `package.json` dependency graphs. Go modules have their own dependency graph (`go.sum`).
- Go's build cache (`GOCACHE`) is already effective and is cached natively by `actions/setup-go`.
- Mixing the two creates confusing cache invalidation rules.

---

## 8. Secrets

All secrets are stored in **GitHub Repository Secrets** or scoped to **GitHub Environments** (`staging`, `production`).

### Repository-level secrets (shared)

| Secret | Used by | Description |
|--------|---------|-------------|
| `DOCKERHUB_USERNAME` | Docker Hub push | Docker Hub username |
| `DOCKERHUB_TOKEN` | Docker Hub push | Docker Hub access token (Read & Write) |
| `GCP_PROJECT_ID` | All GCP steps | Google Cloud project ID |
| `GCP_SA_KEY` | Cloud Run deploy | Service account JSON key |
| `CF_API_TOKEN` | Cloudflare Pages deploy | Cloudflare API token with Pages permissions |
| `CF_ACCOUNT_ID` | Cloudflare Pages deploy | Cloudflare account identifier |

### Environment-scoped secrets

| Secret | Environment | Used by |
|--------|-------------|---------|
| `STAGING_DATABASE_URL` | staging | Atlas migrate step |
| `PROD_DATABASE_URL` | production | Atlas migrate step |
| `STAGING_VITE_CLERK_PUBLISHABLE_KEY` | staging | Vite build for `app` |
| `PROD_VITE_CLERK_PUBLISHABLE_KEY` | production | Vite build for `app` |
| `STAGING_API_URL` | staging | Vite build for `app` (`VITE_API_URL`) |
| `PROD_API_URL` | production | Vite build for `app` (`VITE_API_URL`) |
| `STAGING_APP_URL` | staging | Vite build for `web` (`VITE_APP_URL`) |
| `PROD_APP_URL` | production | Vite build for `web` (`VITE_APP_URL`) |

### Secrets managed via Google Secret Manager

These are **not** GitHub secrets. They are referenced by Cloud Run at runtime using the `secrets:` field in the deploy action:

| Secret Manager Name | Used by |
|---------------------|---------|
| `STAGING_CLERK_SECRET_KEY` / `PROD_CLERK_SECRET_KEY` | Cloud Run API (Clerk auth) |
| `STAGING_JWT_SECRET` / `PROD_JWT_SECRET` | Cloud Run API (JWT signing) |
| `STAGING_DATABASE_URL` / `PROD_DATABASE_URL` | Cloud Run API (database connection) |

> **Best practice**: Prefer Google Secret Manager refs from Cloud Run over passing secrets as plain `env_vars`. The `secrets:` field in `deploy-cloudrun` mounts secrets from Secret Manager at runtime, so they never appear in the Cloud Run revision config.

---

## 9. GCP Service Account Permissions

Create a dedicated CI/CD service account:

```bash
gcloud iam service-accounts create starterpack-ci \
  --display-name="Starterpack CI/CD" \
  --project=$GCP_PROJECT_ID
```

### Required IAM roles

| Role | Purpose |
|------|---------|
| `roles/run.developer` | Deploy new Cloud Run revisions |
| `roles/iam.serviceAccountUser` | Act as the Cloud Run runtime service account |
| `roles/secretmanager.secretAccessor` | Read secrets from Secret Manager at deploy time |

Grant the roles:

```bash
SA_EMAIL="starterpack-ci@${GCP_PROJECT_ID}.iam.gserviceaccount.com"

for ROLE in \
  roles/run.developer \
  roles/iam.serviceAccountUser \
  roles/secretmanager.secretAccessor; do
  gcloud projects add-iam-policy-binding $GCP_PROJECT_ID \
    --member="serviceAccount:${SA_EMAIL}" \
    --role="${ROLE}"
done
```


### Generate the service account key

```bash
gcloud iam service-accounts keys create key.json \
  --iam-account="${SA_EMAIL}"
```

Store the contents of `key.json` as the `GCP_SA_KEY` GitHub secret, then **delete the local file**.

---

## 10. Cloudflare Pages

### Two deployment options

| Option | How | Pros | Cons |
|--------|-----|------|------|
| **Git integration** | Connect repo in CF dashboard | Zero-config, automatic | Cannot enforce deploy order; builds independently of API |
| **`wrangler-action`** | Deploy via GitHub Actions | Full control over ordering via `needs:` | Requires `CF_API_TOKEN` and `CF_ACCOUNT_ID` secrets |

### Recommendation: `wrangler-action`

Use `cloudflare/wrangler-action@v3` in the deploy workflows. This is strongly recommended for production because:

1. **Deploy order is enforced** — frontend deploy jobs use `needs: [deploy-api]`, ensuring the API is live before frontends go out.
2. **Branch-based environments** — Cloudflare Pages treats `--branch=production` as the production deployment and any other branch name (e.g., `--branch=staging`) as a preview deployment.
3. **Consistent build** — the same `bun install` + `turbo build` runs locally and in CI. No separate Cloudflare build step.

### Initial setup

Create the Pages projects in the Cloudflare dashboard (or via Wrangler CLI) **without** connecting a git repo:

```bash
npx wrangler pages project create starterpack-app
npx wrangler pages project create starterpack-web
```

---

## 11. Neon DB: Branch per Environment

Neon supports **database branching**. Use this to isolate staging and production:

```
main (Neon branch)  ← production database
  └── staging       ← branched from main
```

### Create the staging branch

```bash
# Via Neon CLI
neonctl branches create --name staging --project-id $NEON_PROJECT_ID

# Or via the Neon dashboard: Project → Branches → Create Branch
```

Each branch gets its own connection string. Store these as:
- `STAGING_DATABASE_URL` — points to the `staging` branch
- `PROD_DATABASE_URL` — points to the `main` branch

### Migration strategy

Migrations run against **both** branches independently via `make db-migrate-prod`. The `DATABASE_URL` environment variable determines which branch receives the migration:

```yaml
# Staging
- run: make db-migrate-prod
  env:
    DATABASE_URL: ${{ secrets.STAGING_DATABASE_URL }}

# Production
- run: make db-migrate-prod
  env:
    DATABASE_URL: ${{ secrets.PROD_DATABASE_URL }}
```

> **Tip**: If staging schema drifts too far from production, you can reset the staging branch from the Neon dashboard (or CLI) to re-branch from `main`.

---

## 12. golangci-lint

The linter configuration lives at `apps/api/.golangci.yml`.

### CI integration

CI uses `golangci/golangci-lint-action@v6` which:
- Automatically downloads and caches the golangci-lint binary.
- Uses the config file found in `working-directory` (i.e., `apps/api/.golangci.yml`).
- Annotates PRs with inline lint comments.

```yaml
- uses: golangci/golangci-lint-action@v6
  with:
    version: latest
    working-directory: apps/api
```

The Makefile target `lint-api` runs the same linter locally:

```bash
make lint-api
# equivalent to: cd apps/api && golangci-lint run ./...
```

### Keeping configs in sync

The action and the Makefile both read `apps/api/.golangci.yml`. There is no separate CI config. If you modify linter rules, the change applies everywhere.

---

## 13. TanStack Router Generated Files

The TanStack Router Vite plugin generates `routeTree.gen.ts` in each frontend
app's `src/` directory. These files **must be committed** to version control so
that `tsc --noEmit` (typecheck) works in CI without running a Vite build first.

The files are excluded from Biome linting in `biome.jsonc`:

```json
"!**/routeTree.gen.ts"
```

They contain `as any` casts and unsorted imports by design — TanStack Router
generates the code this way intentionally.

> **Do not re-add `**/src/routeTree.gen.ts` to `.gitignore`.** Doing so will
> break `make typecheck` in CI.

---

## Quick Reference: Full Pipeline Diagram

```
PR opened → ci.yml
  ├── generate-check
  ├── api (lint + test)
  └── js  (lint + typecheck + test-affected)

Merge to main → deploy-staging.yml
  ├── test-api ──┐
  ├── lint-api ──┼── build-push → migrate → deploy-api → deploy-app
  └── test-js  ──┘                                     → deploy-web

Merge to release/prod → deploy-prod.yml
  ├── test-api ──┐
  ├── lint-api ──┼── build-push → approve → migrate → deploy-api → deploy-app
  └── test-js  ──┘                                                → deploy-web
```
