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
BIKES_API_PORT=8081
STORAGE_ACCESS_KEY=admin
STORAGE_SECRET_KEY=password123
STORAGE_ENDPOINT=http://minio:9000
STORAGE_PRESIGN_ENDPOINT=http://localhost:9000
STORAGE_PUBLIC_BASE_URL=http://localhost:9000
STORAGE_REGION=us-east-1
STORAGE_BUCKET=velotrace-assets
#
"@ | Out-File -Encoding utf8 .env
```

### 2. Run the Application

Once the `.env` file exists, you can build and start all services:

```powershell
docker-compose up --build
```

### 3. Admin Access Setup (Optional)

To elevate your account to Administrator for testing privileged routes:

1.  **Run the containers** using the command above.
2.  **Log in** to the Web Portal (http://localhost:3000) using your Google account to create your user record in the database.
3.  **Elevate your role** by running this command in your terminal (replace `your-email@example.com` with your actual email):

```bash
docker exec velotrace-db-1 psql -U postgres -d identity -c "UPDATE users SET role = 'admin' WHERE email = 'your-email@example.com';"
```

4.  **Re-login**: Log out and log back in on the Web Portal to receive a fresh token with your new admin role.

### 4. Direct API Authentication (Optional)

If you want to authenticate with the backend directly without using the frontend:

1.  **Obtain a Google ID Token**: You can get one from the [Google OAuth2 Playground](https://developers.google.com/oauthplayground) or by using the browser's DevTools network tab after a successful login.
2.  **Exchange for App JWT**: Use `curl` to call the `identity-api` directly:

```bash
curl -X POST http://localhost:8080/auth/google \
     -H "Content-Type: application/json" \
     -d '{"credential": "YOUR_GOOGLE_ID_TOKEN"}'
```

The API will return a JSON containing the `token`. You can use this token in the `Authorization: Bearer <token>` header for subsequent calls to other services.

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

> **Note**: Our CI/CD pipeline enforces these rules. If your code is not formatted or contains linting errors, the GitHub Action will fail and block deployment.

---
