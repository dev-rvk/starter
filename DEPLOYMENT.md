# Deployment Guide

This guide walks you through everything you need to set up to make the CD
pipeline work. The staging pipeline (`deploy-staging.yml`) deploys automatically
on every push to `main`.

> **Architecture**: Go API → Docker Hub → GCP Cloud Run · React Frontends → Cloudflare Pages · Database → Neon Postgres

---

## Pipeline Overview

```
Push to main → deploy-staging.yml
  ├── test-api  ───┐
  ├── lint-api  ───┼── build-push (Docker Hub) → migrate → deploy-api (Cloud Run) ──┬── deploy-app (CF Pages)
  └── test-js   ───┘                                                                └── deploy-web (CF Pages)
```

**Deploy order is strictly enforced**: Migrations → API → Frontends.

---

## Prerequisites

| Provider | Purpose | Free Tier? |
|----------|---------|------------|
| [Docker Hub](https://hub.docker.com) | Container image registry | Yes (1 private repo, unlimited public) |
| [Google Cloud Platform](https://console.cloud.google.com) | Cloud Run (API hosting), Secret Manager | Yes (generous) |
| [Cloudflare](https://dash.cloudflare.com) | Pages (frontend hosting) | Yes |
| [Neon](https://console.neon.tech) | Serverless Postgres | Yes |
| [GitHub](https://github.com) | Actions (CI/CD) | Yes |

---

## Step 1 — Docker Hub Setup

### 1.1 Create an Account

If you don't already have one, sign up at [hub.docker.com](https://hub.docker.com).

### 1.2 Create the Repository

#### Via UI:
1. Go to [hub.docker.com](https://hub.docker.com) → **Repositories** → **Create Repository**
2. Name: `starterpack-api`
3. Visibility: **Public** (Cloud Run can pull public images without extra auth config)
4. Click **Create**

#### Via CLI:
Docker Hub doesn't have a CLI to create repos — pushing an image auto-creates it:
```bash
docker login
docker build -t YOUR_USERNAME/starterpack-api:test -f apps/api/Dockerfile apps/api
docker push YOUR_USERNAME/starterpack-api:test
```

### 1.3 Create an Access Token

#### Via UI:
1. Go to [hub.docker.com](https://hub.docker.com) → click your avatar → **Account Settings**
2. Go to **Personal access tokens** (left sidebar under "Security")
3. Click **Generate new token**
4. Description: `github-actions-ci`
5. Access permissions: **Read & Write**
6. Click **Generate** and **copy the token** — you won't see it again

You'll store this as `DOCKERHUB_TOKEN` in GitHub secrets (Step 5).

---

## Step 2 — Google Cloud Platform Setup

> **Note**: Since we're using Docker Hub for images, GCP setup is simpler —
> no Artifact Registry needed.

### 2.1 Create a GCP Project

#### Via UI:
1. Go to [console.cloud.google.com](https://console.cloud.google.com)
2. Click the project dropdown at the top → **New Project**
3. Name it (e.g., `starterpack`) and note the **Project ID**

#### Via CLI:
```bash
gcloud projects create starterpack --name="Starterpack"
export GCP_PROJECT_ID="starterpack"
gcloud config set project $GCP_PROJECT_ID
```

### 2.2 Enable Required APIs

#### Via UI:
1. Go to **APIs & Services** → **Library**
2. Search for and enable each of these:
   - Cloud Run Admin API
   - Secret Manager API
   - Identity and Access Management (IAM) API

#### Via CLI:
```bash
gcloud services enable \
  run.googleapis.com \
  secretmanager.googleapis.com \
  iam.googleapis.com
```

### 2.3 Create a CI/CD Service Account

#### Via UI:
1. Go to **IAM & Admin** → **Service Accounts**
2. Click **Create Service Account**
3. Name: `starterpack-ci`, ID: `starterpack-ci`
4. Click **Create and Continue**
5. Add these roles (one by one):
   - `Cloud Run Developer`
   - `Service Account User`
   - `Secret Manager Secret Accessor`
6. Click **Done**

#### Via CLI:
```bash
gcloud iam service-accounts create starterpack-ci \
  --display-name="Starterpack CI/CD" \
  --project=$GCP_PROJECT_ID

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

| Role | Why It's Needed |
|------|-----------------|
| `run.developer` | Deploy new Cloud Run revisions |
| `iam.serviceAccountUser` | Act as the Cloud Run runtime service account |
| `secretmanager.secretAccessor` | Read secrets from GCP Secret Manager at runtime |

### 2.4 Generate & Download the Service Account Key

#### Via UI:
1. Go to **IAM & Admin** → **Service Accounts**
2. Click on `starterpack-ci` → **Keys** tab
3. **Add Key** → **Create new key** → JSON → **Create**
4. A `key.json` file downloads automatically

#### Via CLI:
```bash
gcloud iam service-accounts keys create key.json \
  --iam-account="${SA_EMAIL}"
```

> ⚠️ **Delete `key.json` from your machine after storing it as a GitHub secret.**

### 2.5 Create Secrets in GCP Secret Manager

These secrets are mounted into the Cloud Run container at **runtime**. They are
**not** GitHub secrets — they live inside GCP.

#### Via UI:
1. Go to **Security** → **Secret Manager**
2. Click **Create Secret** for each of the following:

| Secret Name | Value |
|-------------|-------|
| `STAGING_JWT_SECRET` | A random string (e.g., run `openssl rand -base64 32`) |
| `STAGING_DATABASE_URL` | Neon staging connection string (from Step 3) |
| `STAGING_CLERK_SECRET_KEY` | From your [Clerk Dashboard](https://dashboard.clerk.com) → API Keys (skip if not using Clerk) |

#### Via CLI:
```bash
# JWT secret
echo -n "$(openssl rand -base64 32)" | \
  gcloud secrets create STAGING_JWT_SECRET --data-file=-

# Database URL (you'll get this from Neon in Step 3)
echo -n "postgres://user:pass@host/dbname?sslmode=require" | \
  gcloud secrets create STAGING_DATABASE_URL --data-file=-

# Clerk secret key (skip if not using Clerk)
echo -n "sk_test_xxxxx" | \
  gcloud secrets create STAGING_CLERK_SECRET_KEY --data-file=-
```

> Update a secret's value later:
> ```bash
> echo -n "new-value" | gcloud secrets versions add SECRET_NAME --data-file=-
> ```

---

## Step 3 — Cloudflare Pages Setup

### 3.1 Get Your Account ID

#### Via UI:
1. Go to [dash.cloudflare.com](https://dash.cloudflare.com)
2. Click any domain or the main overview page
3. On the right sidebar, find **Account ID** — copy it

### 3.2 Create an API Token

#### Via UI:
1. Go to [dash.cloudflare.com/profile/api-tokens](https://dash.cloudflare.com/profile/api-tokens)
2. Click **Create Token**
3. Click **Get started** next to **Create Custom Token**
4. Configure:
   - Token name: `github-actions-deploy`
   - Permissions: **Account** → **Cloudflare Pages** → **Edit**
   - Account Resources: **Include** → your account
5. Click **Continue to summary** → **Create Token**
6. Copy the token

### 3.3 Create the Pages Projects

> **Important**: Do NOT connect a Git repository. The GitHub Actions workflow
> handles the build and deploy. Connecting Git would cause double deployments.

#### Via UI:
1. Go to **Workers & Pages** → **Create**
2. Select **Pages** → **Upload assets** (Direct Upload)
3. Name the project `starterpack-app` → click **Create Project**
4. Repeat for `starterpack-web`

#### Via CLI:
```bash
npx wrangler pages project create starterpack-app
npx wrangler pages project create starterpack-web
```

---

## Step 4 — Neon Postgres Setup

### 4.1 Create a Project

#### Via UI:
1. Go to [console.neon.tech](https://console.neon.tech)
2. Click **New Project**
3. Name: `starterpack`, Region: pick one close to your Cloud Run region (e.g., `us-east-1`)

### 4.2 Create a Staging Branch

#### Via UI:
1. Go to your project → **Branches** (left sidebar)
2. Click **Create Branch**
3. Name: `staging`, Parent: `main`
4. Click **Create Branch**

#### Via CLI:
```bash
neonctl branches create --name staging --project-id $NEON_PROJECT_ID
```

### 4.3 Get the Connection String

#### Via UI:
1. Go to **Dashboard** → select the `staging` branch
2. Click **Connection Details** → copy the connection string

It looks like:
```
postgres://user:pass@ep-xxxx.us-east-2.aws.neon.tech/neondb?sslmode=require
```

You'll need this for:
- `STAGING_DATABASE_URL` GitHub secret (for Atlas migrations in CI)
- `STAGING_DATABASE_URL` GCP Secret Manager secret (for the Cloud Run container)

---

## Step 5 — GitHub Secrets Configuration

Go to your GitHub repo → **Settings** → **Secrets and variables** → **Actions** → **New repository secret**.

### All Required Secrets

| Secret Name | Value | Where to Get It |
|-------------|-------|-----------------|
| `DOCKERHUB_USERNAME` | Your Docker Hub username | [hub.docker.com](https://hub.docker.com) |
| `DOCKERHUB_TOKEN` | Docker Hub access token | Step 1.3 |
| `GCP_PROJECT_ID` | GCP project ID (e.g., `starterpack`) | GCP Console → project dropdown |
| `GCP_SA_KEY` | Entire contents of `key.json` | Step 2.4 |
| `CF_API_TOKEN` | Cloudflare API token | Step 3.2 |
| `CF_ACCOUNT_ID` | Cloudflare Account ID | Step 3.1 |
| `STAGING_DATABASE_URL` | Neon staging connection string | Step 4.3 |
| `STAGING_VITE_CLERK_PUBLISHABLE_KEY` | Clerk publishable key (`pk_test_xxx`) | [Clerk Dashboard](https://dashboard.clerk.com) → API Keys |
| `STAGING_API_URL` | Cloud Run URL (available after first deploy) | GCP Console → Cloud Run |
| `STAGING_APP_URL` | Cloudflare Pages URL (available after first deploy) | CF Dashboard → Pages |

> **Note on `STAGING_API_URL` and `STAGING_APP_URL`**: These URLs won't exist
> until after your first deploy. For the first deploy, set them to placeholder
> values (e.g., `https://placeholder.example.com`), deploy, then update with the
> real URLs and re-run the workflow.

---

## Step 6 — First Deployment

### 6.1 Verify Locally

```bash
make lint-js lint-api typecheck test
make docker-build TAG=test
```

### 6.2 Push to Main

```bash
git add .
git commit -m "cd: add staging deployment pipeline"
git push origin main
```

Watch the pipeline in your repo's **Actions** tab.

### 6.3 After First Deploy — Update URLs

No custom domain is needed. Every service gets a free URL out of the box:

| Service | Default URL Pattern |
|---------|-------------------|
| API (Cloud Run) | `https://starterpack-api-staging-xxxxx-uc.a.run.app` |
| App (Cloudflare Pages) | `https://staging.starterpack-app.pages.dev` |
| Web (Cloudflare Pages) | `https://staging.starterpack-web.pages.dev` |

After the first deploy completes, grab the actual URLs and update GitHub secrets:

1. **Get Cloud Run URL**:
   - Go to [GCP Console → Cloud Run](https://console.cloud.google.com/run)
   - Click on `starterpack-api-staging` → copy the URL at the top
   - Update `STAGING_API_URL` in GitHub Secrets

2. **Get Cloudflare Pages URLs**:
   - Go to [Cloudflare Dashboard → Workers & Pages](https://dash.cloudflare.com)
   - Click `starterpack-app` → Deployments → the staging deploy shows a URL
     like `staging.starterpack-app.pages.dev`
   - Update `STAGING_APP_URL` in GitHub Secrets

3. **Re-run the workflow** from the Actions tab to rebuild frontends with the
   correct API URL baked in.

> **Custom domains**: If you later buy a domain, you can map it in the
> Cloudflare Pages dashboard (**Custom domains** tab on each project) or via
> Cloud Run (**Integrations** → **Custom domains**). But the default
> `*.pages.dev` and `*.run.app` URLs work perfectly for staging.

---

## Secrets Reference

| Where | Secret Name | Purpose |
|-------|-------------|---------|
| **GitHub** | `DOCKERHUB_USERNAME` | Docker Hub login for pushing images |
| **GitHub** | `DOCKERHUB_TOKEN` | Docker Hub access token |
| **GitHub** | `GCP_PROJECT_ID` | Identifies your GCP project |
| **GitHub** | `GCP_SA_KEY` | Authenticates GitHub Actions with GCP |
| **GitHub** | `CF_API_TOKEN` | Authenticates with Cloudflare |
| **GitHub** | `CF_ACCOUNT_ID` | Identifies your Cloudflare account |
| **GitHub** | `STAGING_DATABASE_URL` | DB connection string for Atlas migrations |
| **GitHub** | `STAGING_VITE_CLERK_PUBLISHABLE_KEY` | Baked into `apps/app` at build time |
| **GitHub** | `STAGING_API_URL` | Baked into `apps/app` at build time |
| **GitHub** | `STAGING_APP_URL` | Baked into `apps/web` at build time |
| **GCP Secret Manager** | `STAGING_CLERK_SECRET_KEY` | Mounted into Cloud Run at runtime |
| **GCP Secret Manager** | `STAGING_JWT_SECRET` | Mounted into Cloud Run at runtime |
| **GCP Secret Manager** | `STAGING_DATABASE_URL` | Mounted into Cloud Run at runtime |

> **Why two places?** GitHub secrets are used during the _build_ phase (CI
> steps that need values at build time, like Vite env vars and migration URLs).
> GCP Secret Manager secrets are used at _runtime_ (injected into the running
> Cloud Run container, so they never appear in image metadata or CI logs).

---

## Troubleshooting

### Pipeline fails at "Build & Push to Docker Hub"
- **Auth error**: Verify `DOCKERHUB_USERNAME` and `DOCKERHUB_TOKEN` are correct.
  The token must have **Read & Write** access.
- **Repo doesn't exist**: If you didn't pre-create the repo, the first push
  auto-creates it as a **public** repo. Verify at
  [hub.docker.com/repositories](https://hub.docker.com/repositories).

### Pipeline fails at "DB Migrations"
- **Connection refused**: Ensure `STAGING_DATABASE_URL` is correct and the Neon
  branch is active (not suspended due to inactivity).
- **Migration checksum mismatch**: Run `make db-hash` locally and commit the
  updated hash file.

### Pipeline fails at "Deploy API to Cloud Run"
- **Permission denied**: Ensure the service account has `run.developer` and
  `iam.serviceAccountUser` roles.
- **Image pull error**: Ensure the Docker Hub repo is **public**. If private,
  you'll need to configure Cloud Run to authenticate with Docker Hub (not
  covered in this guide).
- **Secret not found**: Verify secrets exist in GCP Secret Manager:
  ```bash
  gcloud secrets list
  ```

### Pipeline fails at "Deploy App/Web Frontend"
- **Project not found**: Ensure the Cloudflare Pages projects exist and are
  **not** connected to a Git repo.
- **Auth error**: Verify `CF_API_TOKEN` has `Cloudflare Pages: Edit` permission.
