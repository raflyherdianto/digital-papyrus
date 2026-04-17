# Digital Papyrus API

Production-grade RESTful backend for the **Digital Papyrus** book publishing platform, built with **Go 1.26**, **Gin**, and **SQLite**.

## Architecture

```
backend/
├── cmd/server/          # Application entry point
├── internal/
│   ├── config/          # Centralized configuration (env vars)
│   ├── database/        # SQLite connection, migrations, seeding
│   ├── middleware/       # Auth (JWT), CORS, Rate Limiting, Security Headers
│   ├── model/           # Domain entities (User, Book, Service)
│   ├── handler/         # HTTP request handlers
│   ├── repository/      # Database access layer (parameterized queries)
│   ├── service/         # Business logic layer
│   └── router/          # Route definitions & middleware wiring
├── pkg/
│   ├── response/        # Standardized API response envelope
│   └── validator/       # Input validation utilities
├── tests/               # Integration tests
│   └── testutil/        # Shared test setup
├── Dockerfile           # Multi-stage production build
├── Makefile             # Build, test, lint automation
└── .env.example         # Configuration template
```

## Quick Start

### Prerequisites
- Go 1.26+

### Development

```bash
# Copy and configure environment
cp .env.example .env

# Install dependencies
go mod download

# Run in development mode
go run ./cmd/server/main.go

# Run tests
go test -v ./...
```

The API will be available at `http://localhost:8080`.

### Default Credentials

| Role        | Email                              | Password         |
|-------------|-------------------------------------|-------------------|
| Super Admin | superadmin@digitalpapyrus.web.id   | SuperAdmin@2026!  |
| Author      | author@digitalpapyrus.web.id       | Demo@2026!        |
| Customer    | customer@digitalpapyrus.web.id     | Demo@2026!        |

> ⚠️ **Change all default passwords before deploying to production!**

## API Endpoints

### Public

| Method | Path                | Description        |
|--------|---------------------|---------------------|
| GET    | `/api/v1/health`    | Health check        |
| POST   | `/api/v1/auth/login`| User login          |
| GET    | `/api/v1/books`     | List books (paginated) |
| GET    | `/api/v1/books/:id` | Book detail         |
| GET    | `/api/v1/services`  | List service packages |
| GET    | `/api/v1/services/:id` | Service detail   |

### Protected (JWT Required)

| Method | Path               | Role Required       | Description    |
|--------|--------------------|----------------------|----------------|
| GET    | `/api/v1/auth/me`  | Any authenticated    | Current user   |
| POST   | `/api/v1/auth/logout` | Any authenticated | Logout         |
| POST   | `/api/v1/books`    | superadmin, author   | Create book    |
| PUT    | `/api/v1/books/:id`| superadmin, author   | Update book    |
| DELETE | `/api/v1/books/:id`| superadmin           | Delete book    |
| POST   | `/api/v1/services` | superadmin           | Create service |
| PUT    | `/api/v1/services/:id`| superadmin        | Update service |
| DELETE | `/api/v1/services/:id`| superadmin        | Delete service |

### Query Parameters (Books)

- `page` — Page number (default: 1)
- `per_page` — Items per page (default: 12, max: 100)
- `status` — Filter by status: `draft`, `published`, `archived`
- `category` — Filter by category name
- `search` — Search by title, author, or ISBN

## Security

- **JWT Authentication** with HS256 signing and configurable expiry
- **bcrypt** password hashing (cost=12)
- **Rate Limiting** — 100 req/min general, 5 req/min for login
- **CORS** — Restricted to `https://digitalpapyrus.web.id` in production
- **Security Headers** — HSTS, CSP, X-Frame-Options, X-Content-Type-Options
- **Parameterized SQL queries** — SQL injection prevention
- **Input validation & sanitization** on all endpoints
- **Graceful shutdown** with 30-second drain period
- **Request ID tracing** via `X-Request-ID` header

## Deployment (Docker + Cloudflare Tunnel)

From the project root (`digital-papyrus/`):

```bash
# Set required secrets
export JWT_SECRET="your-strong-secret-here"
export SEED_SUPERADMIN_PASSWORD="YourSecurePassword!"
export CLOUDFLARE_TUNNEL_TOKEN="your-tunnel-token"

# Build & deploy
docker compose up -d --build
```

### Domain Mapping (Cloudflare Tunnel)

| Domain                         | Target Container | Port |
|--------------------------------|------------------|------|
| `api.digitalpapyrus.web.id`    | dp-backend       | 8080 |
| `digitalpapyrus.web.id`        | dp-frontend      | 80   |

## Testing

```bash
# Run all tests
go test -v ./...

# With coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
```
