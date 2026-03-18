# VeloTrace Project Blueprint

## 🛠 Local Development Setup

To run the project locally using Docker, you must first create a `.env` file in the root directory. This file is used as the **Single Source of Truth** for the project's configuration (ports, authentication, etc.).

### 1. Create the `.env` File

You can create it manually or run this simple **PowerShell script** in the project root:

```powershell
# Create .env file with default development values
@"
WEB_PORTAL_PORT=3000
GOOGLE_CLIENT_ID=your_google_client_id_here
IDENTITY_API_URL=http://localhost:8080
IDENTITY_API_PORT=8080
"@ | Out-File -Encoding utf8 .env
```

### 2. Run the Application

Once the `.env` file exists, you can build and start all services:

```powershell
docker-compose up --build
```

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
# Run deep analysis and catch bugs/mistakes
docker-compose exec identity golangci-lint run ./... -v
```

> **Note**: Our CI/CD pipeline enforces these rules. If your code is not formatted or contains linting errors, the GitHub Action will fail and block deployment.

---
