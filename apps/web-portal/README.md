# VeloTrace Web Portal

Nuxt 3 application for the VeloTrace monorepo.

## 🚀 Development Setup

This project uses **Docker** and **VS Code Dev Containers** to ensure a clean host machine and a consistent development environment.

### 1. Start the Services

From the **project root**, run:

```powershell
docker compose up --build web-portal
```

The app will be available at `http://localhost:3000`. Keep this terminal running.

### 2. Connect your IDE (Recommended)

While the container is already running:

1. Open the `apps/web-portal` folder in **VS Code**.
2. When prompted, click **"Reopen in Container"**.
3. Once the container is ready, open the VS Code terminal and run:
   ```bash
   npm run postinstall
   ```
   _This generates the hidden `.nuxt` types and fixes "missing import" errors._

## 🛠 Tech Stack

- **Framework**: Nuxt 3 (Vue 3)
- **Language**: TypeScript
- **Styling**: Tailwind CSS
- **Platform**: Vercel (Production)

## 🐳 Docker Details

- **User**: `node` (to avoid permission issues)
- **HMR**: Enabled via volume mapping and `WATCHPACK_POLLING`.
- **Isolation**: `node_modules` and `.nuxt` are stored in anonymous volumes (container-only).
