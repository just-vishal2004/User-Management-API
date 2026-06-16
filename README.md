# Ainyx Backend — User Management API

A production-quality RESTful API built with **Go**, **Fiber**, **PostgreSQL**, and **SQLC**. Manages users with their name and date of birth, calculating age dynamically at query time.

Built as part of the Ainyx Solutions Software Engineering Intern (Backend) assessment.

---

## Tech Stack

| Layer | Technology |
|---|---|
| Language | Go 1.26 |
| HTTP Framework | Fiber v2 |
| Database | PostgreSQL 16 |
| DB Access Layer | SQLC (type-safe SQL) |
| DB Driver | pgx/v5 |
| Migrations | golang-migrate |
| Logging | Uber Zap (structured JSON) |
| Validation | go-playground/validator |
| Containerization | Docker + docker-compose |

---

## Project Structure

ainyx-backend/

├── cmd/server/          # Application entry point (main.go)

├── config/              # Environment-based configuration

├── db/

│   ├── migrations/      # SQL migration files (up/down)

│   ├── sqlc/            # SQLC config and raw SQL queries

│   └── sqlc/generated/  # Auto-generated type-safe Go DB code

├── internal/

│   ├── handler/         # HTTP handlers (request parsing, response)

│   ├── logger/          # Uber Zap structured logger setup

│   ├── middleware/       # Request ID + duration logging middleware

│   ├── models/          # Request/response/pagination structs

│   ├── repository/      # Database access layer

│   ├── routes/          # Route registration

│   └── service/         # Business logic + age calculation

├── Dockerfile           # Multi-stage production Docker build

├── docker-compose.yml   # App + PostgreSQL orchestration

├── Makefile             # Common development commands

└── .env.example         # Environment variable template

---

## Prerequisites

- Go 1.21+
- PostgreSQL 16
- [golang-migrate](https://github.com/golang-migrate/migrate)
- [sqlc](https://sqlc.dev)
- Docker (optional, for containerized setup)

---

## Local Setup

### 1. Clone the repository

```bash
git clone https://github.com/vishalkumar/ainyx-backend.git
cd ainyx-backend
```

### 2. Configure environment

```bash
cp .env.example .env
```

Edit `.env` with your local PostgreSQL credentials:

```env
APP_PORT=3000
APP_ENV=development
DB_HOST=localhost
DB_PORT=5432
DB_USER=your_db_user
DB_PASSWORD=your_db_password
DB_NAME=ainyx_users
DB_SSLMODE=disable
```

### 3. Create the database

```bash
psql -U your_db_user -d postgres -c "CREATE DATABASE ainyx_users;"
```

### 4. Run migrations

```bash
make migrate-up
```

### 5. Start the server

```bash
make run
```

The server starts at `http://localhost:3000`.

---

## Docker Setup

Run the entire stack (app + PostgreSQL) with one command:

```bash
make docker-up
```

This builds the image, starts PostgreSQL with a healthcheck, and launches the app once the database is ready.

To stop:

```bash
make docker-down
```

---

## API Reference

Base URL: `http://localhost:3000/api/v1`

### Health Check
GET /health

Response:
```json
{ "status": "ok" }
```

---

### Create User
POST /api/v1/users

Request body:
```json
{
  "name": "Alice Smith",
  "dob": "1998-05-15"
}
```

Response `201 Created`:
```json
{
  "id": 1,
  "name": "Alice Smith",
  "dob": "1998-05-15",
  "age": 28
}
```

---

### Get User by ID
GET /api/v1/users/:id

Response `200 OK`:
```json
{
  "id": 1,
  "name": "Alice Smith",
  "dob": "1998-05-15",
  "age": 28
}
```

Response `404 Not Found`:
```json
{ "error": "user not found" }
```

---

### Update User
PUT /api/v1/users/:id

Request body:
```json
{
  "name": "Alice Updated",
  "dob": "1998-05-15"
}
```

Response `200 OK`:
```json
{
  "id": 1,
  "name": "Alice Updated",
  "dob": "1998-05-15",
  "age": 28
}
```

---

### Delete User
DELETE /api/v1/users/:id

Response: `204 No Content`

---

### List Users (with Pagination)
GET /api/v1/users?page=1&page_size=10

| Parameter | Type | Default | Description |
|---|---|---|---|
| `page` | integer | 1 | Page number |
| `page_size` | integer | 10 | Results per page (max 100) |

Response `200 OK`:
```json
{
  "data": [
    {
      "id": 2,
      "name": "Bob Johnson",
      "dob": "1990-11-30",
      "age": 35
    }
  ],
  "total": 3,
  "page": 1,
  "page_size": 10,
  "total_pages": 1
}
```

---

## Validation Rules

| Field | Rules |
|---|---|
| `name` | Required, 2–100 characters |
| `dob` | Required, format `YYYY-MM-DD`, cannot be in the future |

Validation error response `400 Bad Request`:
```json
{ "error": "Name is required" }
```

---

## Available Make Commands

```bash
make build          # Compile binary to bin/
make run            # Run the server
make dev            # Run with live reload (air)
make test           # Run all tests
make test-cover     # Run tests with coverage
make migrate-up     # Apply all migrations
make migrate-down   # Roll back last migration
make sqlc           # Regenerate SQLC code
make lint           # Run go vet
make docker-up      # Start with docker-compose
make docker-down    # Stop docker-compose
make clean          # Remove build artifacts
```

---

## Design Decisions

**Why SQLC?** Type-safe SQL without an ORM. Queries are plain SQL — readable, optimizable, and verified at generation time rather than runtime.

**Why no `age` column?** Age changes every year. Storing it creates stale data. We store `dob` (the immutable fact) and calculate age dynamically using Go's `time` package.

**Why layered architecture?** Handler → Service → Repository separation means each layer has one responsibility. The database can be swapped without touching handlers. Business logic can be tested without HTTP.

**Why `pgxpool`?** A connection pool handles concurrent requests efficiently. Each request gets its own connection from the pool without blocking others.

---

## Running Tests

```bash
make test
```

Unit tests cover the age calculation function with 8 cases including edge cases: birthdays today, leap years, newborns, and future birthdays.