# VeloTrace Project Blueprint

Welcome to the VeloTrace project repository. This is a high-trust Bicycle Proof of Ownership Registry & Marketplace managed as an Nx Monorepo.

## 📖 Documentation

For detailed instructions on setting up your local environment, running the application, and following our development workflows, please refer to the:

👉 **[Onboarding Guide (ONBOARDING.md)](./ONBOARDING.md)**

## 🚀 Quick Start

If you already have your environment configured:

```powershell
docker-compose up --build
```

---

## 🌍 Production Environments

- **Web Portal**: [https://velotrace-seven.vercel.app/](https://velotrace-seven.vercel.app/)
- **Identity API**: [https://velotrace.onrender.com/](https://velotrace.onrender.com/)
- **Bikes API**: [https://velotrace-bikes-api.onrender.com/](https://velotrace-bikes-api.onrender.com/)

---

## 🏗 Microservices Architecture

- **Identity Service (`apps/identity-api`)**: User accounts and Legal Identity verification (Go).
- **Bikes Service (`apps/bikes-api`)**: Bike registration, ownership tracking, and marketplace (Go).
- **Web Portal (`apps/web-portal`)**: Frontend application (Nuxt 3).

## 🛡 Security & Privacy

This project follows the **"Silent Sentry"** principle (Security-by-design). Internal hashes, verification flags, and sensitive metadata are never leaked to the frontend or third parties.

---

_For more information, see the [Project Overview](./project-overview.md)._
