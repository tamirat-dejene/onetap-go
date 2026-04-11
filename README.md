# Salary Advance Loan Service

A Go backend for a salary advance loan system, implementing Clean Architecture with dependency injection, JWT authentication, and an in-memory database.

## Table of Contents
- [Quick Start](#quick-start)
- [Architecture](#architecture)
- [Configuration](#configuration)
- [API Endpoints](#api-endpoints)
- [Validation Logic](#validation-logic)
- [Rating Logic](#rating-logic)
- [Security Measures](#security-measures)
- [Testing](#testing)

## Quick Start

```bash
# 1. Clone / navigate
cd /path/to/onetap-go

# 2. Copy example env file and configure
cp .env.example .env     # already done; edit JWT_SECRET for production

# 3. Install dependencies
go mod tidy

# 4. Run the server
make run
# OR
go run ./cmd/server/main.go

# 5. Server is available at http://localhost:8080
```

## Architecture

The project follows **Clean Architecture** with strict layer separation:

```
cmd/server/          ← Entrypoint (wires all layers via DI)
internal/
  domain/            ← Entities + Repository/Usecase interfaces (no deps)
  repository/memory/ ← In-memory implementations of repository interfaces
  usecase/           ← Business logic (depends only on domain interfaces)
  delivery/http/     ← HTTP handlers, middleware, router (depends on usecase interfaces)
pkg/
  config/            ← .env loader
  jwt/               ← JWT sign/verify
  loader/            ← Seed data from JSON files on startup
```

**Dependency Injection** is performed manually in `cmd/server/main.go` — no framework required.

## Configuration

All variables are loaded from `.env` (never hardcoded). See `.env.example`:

| Variable | Default | Description |
|---|---|---|
| `SERVER_PORT` | `8080` | HTTP listen port |
| `JWT_SECRET` | — | **Required.** HS256 signing key |
| `JWT_EXPIRY_HOURS` | `24` | Token validity period |
| `ADMIN_USERNAME` | — | Admin login username |
| `ADMIN_PASSWORD` | — | Admin login password |
| `UPLOADER_USERNAME` | — | Uploader login username |
| `UPLOADER_PASSWORD` | — | Uploader login password |
| `BCRYPT_COST` | `12` | bcrypt work factor (10–14 recommended) |
| `RATE_LIMIT_REQUESTS` | `5` | Max login attempts per window |
| `RATE_LIMIT_WINDOW_SECONDS` | `60` | Rate limit window duration |
| `CUSTOMERS_FILE` | `inputs/customers.json` | Canonical customer list |
| `TRANSACTIONS_FILE` | `inputs/transactions.json` | Transaction history |
| `SAMPLE_CUSTOMERS_FILE` | `inputs/sample_customers.csv` | Batch to validate |

## API Endpoints

### Authentication
| Method | Path | Auth | Description |
|---|---|---|---|
| `POST` | `/api/v1/auth/login` | Public | Get JWT token |

### Customers
| Method | Path | Auth | Description |
|---|---|---|---|
| `GET` | `/api/v1/customers/validate` | Bearer | Validate sample CSV & persist verified records |
| `GET` | `/api/v1/customers/verified` | Bearer | List all verified customers |

### Ratings
| Method | Path | Auth | Description |
|---|---|---|---|
| `GET` | `/api/v1/ratings` | Bearer | Ratings for all verified customers |
| `GET` | `/api/v1/ratings/{accountNo}` | Bearer | Rating for one account |
| `GET` | `/health` | Public | Health check |

**Example workflow:**
```bash
# 1. Login
TOKEN=$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"Admin@123"}' | jq -r .token)

# 2. Validate customers
curl -s http://localhost:8080/api/v1/customers/validate \
  -H "Authorization: Bearer $TOKEN" | jq .

# 3. Get all ratings (run AFTER validate)
curl -s http://localhost:8080/api/v1/ratings \
  -H "Authorization: Bearer $TOKEN" | jq .

# 4. Get rating for a specific account
curl -s http://localhost:8080/api/v1/ratings/1050001047301 \
  -H "Authorization: Bearer $TOKEN" | jq .
```

## Validation Logic

When `GET /api/v1/customers/validate` is called:

1. The service reads `inputs/sample_customers.csv` (50 records with intentional errors).
2. Each record is checked against the canonical `inputs/customers.json`:
   - **Account number must exist** in the canonical list (exact match).
   - **Customer name must match** the canonical entry (case-insensitive, whitespace-normalized).
3. Records passing all checks are flagged as `verified: true` and persisted to the in-memory store.
4. The response includes a full report: `total`, `verified`, `failed`, and per-record details with error descriptions and the normalized record for valid entries.

**Known intentional errors in sample data:**
- Mismatched names (e.g. different person's name for the same account)
- Wrong account number format/digits

## Rating Logic

Ratings (1–10) are computed from four weighted factors:

| Factor | Weight | Computation |
|---|---|---|
| **Transaction Count** | 25% | `min(count / 50, 1.0)` |
| **Total Volume** | 30% | `log10(volume + 1) / log10(10,000,000)` |
| **Time Span** | 25% | `min(span_days / 365, 1.0)` |
| **Balance Stability** | 20% | `min(avg_cleared_balance / declared_balance, 1.0)` |

Final score: `round(composite × 9) + 1`, clamped to [1, 10].

| Score | Label |
|---|---|
| 1–2 | Very High Risk |
| 3–4 | High Risk |
| 5–6 | Moderate Risk |
| 7–8 | Low Risk |
| 9–10 | Very Low Risk / Excellent |

**Synthetic transactions:** Customers with no transaction history receive 1–3 deterministically generated small transactions (seeded by account number) to enable rating.

## Security Measures

| Measure | Implementation |
|---|---|
| **Password hashing** | bcrypt with configurable cost factor (default 12) |
| **JWT authentication** | HS256-signed tokens, verified on every protected request |
| **Rate limiting** | Per-IP sliding window; 5 login attempts / 60 seconds (configurable) |
| **No hardcoded secrets** | All credentials and secrets loaded from `.env` |
| **Panic recovery** | Chi `Recoverer` middleware prevents crashes from leaking stack traces |
| **Request ID tracking** | Chi `RequestID` middleware for traceability |

## Testing

```bash
# Run all tests
make test

# Run with coverage (≥70% required)
make coverage

# View HTML coverage report
open coverage.html
```

Test coverage targets:
- **Auth usecase** — login success/failure, rate limiting, JWT validity
- **Customer usecase** — validation logic, name matching, persistence
- **Transaction usecase** — rating calculation, synthetic generation, edge cases
- **Repositories** — all CRUD operations on in-memory stores

## API Documentation

Full OpenAPI 3.0 specification: [`docs/swagger.yaml`](docs/swagger.yaml)

You can load this into [Swagger UI](https://editor.swagger.io) or any OpenAPI-compatible tool.
