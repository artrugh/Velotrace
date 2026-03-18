## 1. Project Overview

- **Goal**: A high-trust Bicycle Proof of Ownership Registry & Marketplace.
- **Philosophy**: "Registry-First" (Identity must be verified before property can be registered).
- **Core Principle**: The "Silent Sentry" (Security-by-design; never leak internal hashes or verification flags).

## 2. Technical Stack

- **Monorepo Management**: [Nx.dev](https://nx.dev) (using `@nx-go/nx-go` and `@nx/nuxt`).
- **Backend**: Go 1.26+ with the **Echo** framework.
- **Database**: PostgreSQL (Managed via Supabase for production).
- **WebPortal**: Nuxt 3/4 (Vue.js)
- **Infrastructure**: Docker-based local development with a shared `velotrace` network.
- **CI/CD**: GitHub Actions using `nx affected` to stay within free-tier limits.

## 3. Microservices Architecture

### Identity Service (`apps/identity-api`)

- **Responsibility**: User accounts and Legal Identity verification.
- **Auth Strategy**:
  - **Step 1 (Guest)**: Social Sign-up via Google OAuth. Store `google_id`, `email`, and `display_name`.
  - **Step 2 (Verified)**: Legal Identity via EU Digital Identity Wallet (OpenID4VP).
- **Database Fields**: `id`, `email`, `google_id`, `display_name`, `first_name` (Legal), `last_name` (Legal), `is_verified`, `id_card_hash`.
- **Privacy**: Sensitive fields (`google_id`, `id_card_hash`) must use `json:"-"` in Go structs.

### 4. Development & Deployment Logic

- **Local Dev**: Use `docker-compose.yml`.
  - Services: `db` (Postgres 16-alpine), `identity-api` (using `Dockerfile.dev` + `Air`).
- **Production Deployment**:
  - **Target**: Koyeb or Render ($0 Free Tiers).
  - **Build**: Multi-stage `Dockerfile` to produce a minimal static Go binary (under 20MB).
  - **Dynamic Configuration**: The app must read `PORT` and `DATABASE_URL` from env variables.

### Vault Service (Planned)

- **Responsibility**: Bicycle "Digital Twin" registration and ownership history.

## Frontend: WebPortal (Nuxt)

- **Framework**: Nuxt 3/4 (Vue.js)
- **Deployment**: Vercel (Native Nitro Preset)
- **Development**: Dockerized with Hot Module Replacement (HMR)
- **Integration**: Consumes Identity API (Go) via `useFetch`
- **Build Tool**: Nx (Affected builds)