# VeloTrace

VeloTrace is a Proof of Ownership Registry and Marketplace designed to secure bicycle ownership and facilitate safe transfers.

## Project Overview

VeloTrace aims to combat bike theft and fraud by providing a blockchain-inspired registry where owners can verify their identity and register their bicycles. The platform supports a progressive profiling model, allowing users to start as guests and transition to verified owners using modern identity standards.

## Architecture

The project follows a microservices architecture:

- **Identity Service**: Handles user registration, social authentication (Google OAuth 2.0), and legal identity verification (EU Digital Identity Wallet / OpenID4VP).
- **Core Services** (Upcoming): Ownership registry, bike registration, and marketplace functionality.

## Tech Stack

- **Backend**: Go with the [Echo](https://echo.labstack.com/) framework.
- **Database**: PostgreSQL 16.
- **Infrastructure**: Docker & Docker Compose for local development and orchestration.
- **Identity**: 
  - Google OAuth 2.0 (OIDC) for guest access.
  - OpenID4VP (EU Digital Identity Wallet) for verified ownership.

## Getting Started

### Prerequisites

- [Docker](https://www.docker.com/) and [Docker Compose](https://docs.docker.com/compose/) installed.
- [Go](https://golang.org/) 1.22+ (for local development outside Docker).

### Running the Project

1. Clone the repository.
2. Start the services using Docker Compose:
   ```bash
   docker-compose up --build
   ```
3. The Identity Service will be available at `http://localhost:8080`.

## User Lifecycle & Progressive Profiling

1. **Stage 1: Guest (Social Auth)**
   - Authenticate via Google.
   - Access: Read-only browsing.
2. **Stage 2: Verified Owner (Legal Identity)**
   - Authenticate via EU Digital Identity Wallet.
   - Access: Create/Transfer bike ownership.

## Project Structure

```text
.
├── services/
│   └── identity/       # Go-based identity and verification service
├── docker-compose.yml  # Local development orchestration
└── README.md           # Project documentation
```
