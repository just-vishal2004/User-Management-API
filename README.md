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