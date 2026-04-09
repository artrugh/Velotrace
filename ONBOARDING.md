# Velotrace Onboarding Guide

Welcome to the **Velotrace** project! This guide will help you set up your development environment and understand the architecture, infrastructure, and workflows used in this monorepo.

---

## 🚀 Getting Started

### 1. Prerequisites

Ensure you have the following installed on your machine:

- **Node.js** (v20+ recommended) & **pnpm** (v10+)
- **Docker** & **Docker Compose**

### 2. Initial Setup in Containers

1.  **Clone the repository:**
    ```bash
    git clone https://github.com/your-repo/velotrace.git
    cd velotrace
    ```
2.  **Install dependencies:**
    ```bash
    pnpm install
    ```
3.  **Environment Variables:**
    The project uses a root `.env` file for Docker Compose and local development. Create one by copying the following template:

    ```bash
    # Infrastructure
    WEB_PORTAL_PORT=3000
    GOOGLE_CLIENT_ID=your_google_client_id
    IDENTITY_API_URL=http://identity:8080
    BIKES_API_URL=http://bikes:8081
    IDENTITY_API_PORT=8080
    BIKES_API_PORT=8081
    STORAGE_ACCESS_KEY=admin
    STORAGE_SECRET_KEY=password123
    STORAGE_ENDPOINT=http://minio:9000
    STORAGE_PRESIGN_ENDPOINT=http://localhost:9000
    STORAGE_PUBLIC_BASE_URL=http://localhost:9000
    STORAGE_REGION=us-east-1
    STORAGE_BUCKET=velotrace-assets

    # Security (RS256 Keys - generate for local use)
    JWT_PRIVATE_KEY="your-private-key-content"
    JWT_PUBLIC_KEY="your-public-key-content"
    ```

### 3. Running the Project

The entire stack can be started using Docker Compose:

```bash
docker compose up --build
```

This will start:

- **PostgreSQL**: Database for all services.
- **Minio**: S3-compatible storage for bike images.
- **Identity API**: Go service for user management (port 8080).
- **Bikes API**: Go service for bike registration (port 8081).
- **Web Portal**: Nuxt 3 frontend (port 3000).

### 3. Running WebPortal outside docker

```bash
# Keep the backend containers running first, then replace only the web-portal:
node ./tools/setup-web-portal.mjs "your_google_client_id"
```

---

## 🏗 Architecture Overview

Velotrace is managed as an **Nx Monorepo**, ensuring consistency across the backend and frontend.

### 1. Monorepo Structure

- `apps/identity-api`: Go backend for authentication and user profiles.
- `apps/bikes-api`: Go backend for bike registry and marketplace.
- `apps/web-portal`: Nuxt 3 (Vue.js) frontend.
- `libs/go-auth`: Shared Go library for JWT verification (RS256).
- `libs/go-utils`: Shared Go utilities (validators, logging).
- `libs/api-contract`: Shared OpenAPI/Swagger definitions and generated TypeScript types.

### 2. Backend Design

- **Framework**: [Echo](https://echo.labstack.com/) for Go.
- **Security**: Asymmetric RS256 JWT. `identity-api` issues tokens; other services verify them using the shared public key.
- **Documentation**: Swagger specs are generated automatically from Go code using `swaggo/swag`.

### 3. Frontend Design

- **Framework**: Nuxt 3 with TypeScript.
- **API Integration**: Uses `openapi-fetch` for type-safe requests based on the shared API contracts.

---

## 🛠 Development Workflow

### 1. Admin Access Setup (Optional)

To elevate your account to Administrator for testing privileged routes:

1.  **Run the containers** using the command above.
2.  **Log in** to the Web Portal (http://localhost:3000) using your Google account to create your user record in the database.
3.  **Elevate your role** by running this command in your terminal (replace `your-email@example.com` with your actual email):
    ```bash
    docker exec velotrace-db-1 psql -U postgres -d identity -c "UPDATE users SET role = 'admin' WHERE email = 'your-email@example.com';"
    ```
4.  **Re-login**: Log out and log back in on the Web Portal to receive a fresh token with your new admin role.

### 2. Direct API Authentication (Optional)

If you want to authenticate with the backend directly without using the frontend:

1.  **Obtain a Google ID Token**: You can get one from the [Google OAuth2 Playground](https://developers.google.com/oauthplayground).
2.  **Exchange for App JWT**: Use `curl` to call the `identity-api` directly:
    ```bash
    curl -X POST http://localhost:8080/auth/google \
         -H "Content-Type: application/json" \
         -d '{"credential": "YOUR_GOOGLE_ID_TOKEN"}'
    ```

### 3. Storage Public Access Setup

By default, MinIO buckets are private. To allow the web portal to display uploaded bike images, you must set the `bikes/` folder to public read access:

```bash
docker-compose exec minio sh -c "mc alias set local http://localhost:9000 admin password123 && mc anonymous set download local/velotrace-assets/bikes"
```

### 4. Database Migrations (Goose)

Migrations are located in `apps/[service]/internal/db/sql/` and are triggered automatically when the containers start.

- **Apply migrations**:
  ```bash
  goose -dir apps/identity-api/internal/db/sql postgres "postgres://postgres:postgres@localhost:5432/identity?sslmode=disable" up
  ```
- **Create new migration**:
  ```bash
  goose -dir apps/identity-api/internal/db/sql create your_migration_name sql
  ```

### 5. API Contract Synchronization

When backend routes change, the frontend types must be updated:

1.  **Generate Swagger**: `nx generate-swang identity-api`
2.  **Generate TS Types**: Run the sync script (integrated in the build pipeline or via `node tools/generate-api-contracts.js`).

---

## 🧹 Code Quality & Linting

To maintain code consistency and prevent syntax errors, please run the following commands before pushing your changes:

### 1. Format & Lint Frontend (Nuxt, YAML, JSON)

Run this from your host machine:

```powershell
npx prettier --write .
```

### 2. Format & Lint Backend (Go)

Run these inside the running Docker containers:

```powershell
# Auto-format Go code
docker-compose exec identity go fmt ./...
docker-compose exec bikes go fmt ./...
# Run deep analysis and catch bugs/mistakes
docker-compose exec identity golangci-lint run ./... -v
docker-compose exec bikes golangci-lint run ./... -v
```

---

## ☁ CI/CD & Infrastructure

### 1. GitHub Actions

Workflows are defined in `.github/workflows/`:

- `ci-quality.yml`: Runs linting and tests on every PR.
- `deploy-*.yml`: Handles deployment to Render (APIs) and Vercel (Web Portal) using `nx affected`.

### 2. Infrastructure

- **Development**: Docker Compose manages all dependencies.
- **Storage**: Minio is used locally to simulate S3. Access the console at `http://localhost:9001`.

---

## 🛡 Security Standards

- **Secrets**: Never commit `.env` files or raw JWT keys to Git.
- **Privacy**: sensitive Go struct fields must use `json:"-"`.
- **Verification**: Use the "Silent Sentry" principle—avoid leaking internal validation flags to the frontend.

---

Happy coding! If you have any questions, reach out to the Platform Team.
