# User API — Go + Fiber + SQLC + PostgreSQL

A small, production-style RESTful API for managing users with a `name` and a date of birth (`dob`). The user's **age is calculated dynamically** on read using Go's `time` package — it is never stored in the database.

> 📖 Looking for the demo screenshots? See **[docs/SCREENSHOTS.md](docs/SCREENSHOTS.md)** (also embedded in the [Demo](#demo) section below).

## Features

- Full CRUD for users (`POST` / `GET` / `PUT` / `DELETE`)
- **Age computed on the fly** from `dob` — handles leap days, birthdays not yet reached, and clamps future dates to `0`
- SQLC-generated, type-safe database layer
- Structured logging with Uber Zap
- Request validation with go-playground/validator
- Clean, consistent HTTP status codes and JSON error envelope
- **Bonus:** Docker support, pagination, request-id + duration middleware, unit-tested age logic

## Tech stack

- **[GoFiber](https://gofiber.io/)** — HTTP framework
- **PostgreSQL** + **[SQLC](https://sqlc.dev/)** — typed DB access layer generated from SQL
- **[pgx/v5](https://github.com/jackc/pgx)** — PostgreSQL driver / connection pool
- **[Uber Zap](https://github.com/uber-go/zap)** — structured logging
- **[go-playground/validator](https://github.com/go-playground/validator)** — request validation

## Project layout

```
.
├── cmd/server/main.go          # entrypoint: wiring, DB pool, graceful shutdown
├── config/                     # env-based configuration
├── db/
│   ├── migrations/             # up/down SQL migrations
│   ├── queries/users.sql       # SQLC query definitions
│   ├── schema.sql              # schema used by SQLC for codegen
│   └── sqlc/                   # generated DB access layer
├── internal/
│   ├── handler/                # HTTP handlers (Fiber)
│   ├── repository/             # persistence wrapper over sqlc
│   ├── service/                # business logic + age calculation + mapping
│   ├── routes/                 # route registration + error handler
│   ├── middleware/             # request-id + duration logging
│   ├── models/                 # DTOs, validation tags, CalculateAge
│   └── logger/                 # Zap logger factory
├── docs/                       # screenshots & extra docs
├── Dockerfile
├── docker-compose.yml
├── Makefile
├── sqlc.yaml
└── README.md
```

---

## Quick start with Docker (recommended)

The only prerequisite is Docker. One command builds the app, starts PostgreSQL, runs the migration automatically, and serves the API on `http://localhost:8080`:

```bash
docker compose up --build
```

When you see `connected to database` and `starting server` in the logs, it's ready. Stop with `Ctrl+C`, then `docker compose down` (add `-v` to also wipe the database).

## Quick start locally (without Docker)

Prerequisites: Go 1.22+, PostgreSQL 13+.

```bash
# 1. Create the database
createdb userdb

# 2. Apply the migration
psql "postgres://postgres:postgres@localhost:5432/userdb?sslmode=disable" \
  -f db/migrations/000001_create_users_table.up.sql

# 3. Download dependencies (also generates go.sum)
go mod tidy

# 4. Run the server
export DATABASE_URL="postgres://postgres:postgres@localhost:5432/userdb?sslmode=disable"
go run ./cmd/server
```

---

## Configuration

All configuration comes from environment variables (see `.env.example`). With Docker, these are already set in `docker-compose.yml`. You can either set `DATABASE_URL` directly or provide the individual `DB_*` parts.

| Variable | Default | Description |
|----------|---------|-------------|
| `APP_PORT` | `8080` | HTTP port |
| `LOG_LEVEL` | `info` | `debug` / `info` / `warn` / `error` |
| `DATABASE_URL` | built from `DB_*` | full Postgres DSN |
| `DB_HOST` | `localhost` | |
| `DB_PORT` | `5432` | |
| `DB_USER` | `postgres` | |
| `DB_PASSWORD` | `postgres` | |
| `DB_NAME` | `userdb` | |
| `DB_SSLMODE` | `disable` | |

---

## API

Base URL: `http://localhost:8080`

Every response carries an `X-Request-Id` header (echoed from the request if present, otherwise generated). Request method, path, status, and duration are logged via Zap.

### Create user — `POST /users`

```bash
curl -X POST http://localhost:8080/users \
  -H "Content-Type: application/json" \
  -d '{"name":"Alice","dob":"1990-05-10"}'
```

`201 Created`
```json
{ "id": 1, "name": "Alice", "dob": "1990-05-10" }
```

### Get user — `GET /users/:id`

```bash
curl http://localhost:8080/users/1
```

`200 OK` — note the dynamically computed `age`:
```json
{ "id": 1, "name": "Alice", "dob": "1990-05-10", "age": 35 }
```

### List users — `GET /users`

Supports optional pagination via `?page=` and `?limit=` (defaults: page 1, limit 20, max limit 100). Pagination metadata is returned in the `X-Total-Count`, `X-Page`, and `X-Limit` response headers, while the body stays a plain JSON array.

```bash
curl "http://localhost:8080/users?page=1&limit=10"
```

`200 OK`
```json
[
  { "id": 1, "name": "Alice", "dob": "1990-05-10", "age": 35 }
]
```

### Update user — `PUT /users/:id`

```bash
curl -X PUT http://localhost:8080/users/1 \
  -H "Content-Type: application/json" \
  -d '{"name":"Alice Updated","dob":"1991-03-15"}'
```

`200 OK`
```json
{ "id": 1, "name": "Alice Updated", "dob": "1991-03-15" }
```

### Delete user — `DELETE /users/:id`

```bash
curl -X DELETE http://localhost:8080/users/1
```

`204 No Content`

### Health — `GET /health`

`200 OK` → `{ "status": "ok" }`

### Status codes & errors

| Situation | Status |
|-----------|--------|
| Created | `201` |
| OK | `200` |
| Deleted | `204` |
| Invalid body / validation failure / bad id | `400` |
| User not found | `404` |
| Unexpected error | `500` |

Errors use a consistent envelope:
```json
{ "error": "user not found", "status": 404 }
```

---

## Demo

> Replace the placeholder images below with your own screenshots (drop them in `docs/img/`). Full gallery in **[docs/SCREENSHOTS.md](docs/SCREENSHOTS.md)**.

**Containers running**

![Docker compose up](docs/img/01-docker-up.png)

**Create a user — returns `201`**

![Create user](docs/img/02-create.png)

**Get user — note the dynamically computed `age`**

![Get user with age](docs/img/03-get-age.png)

**List with pagination headers**

![List users](docs/img/04-list.png)

---

## Running tests

```bash
go test ./... -v
```

The age-calculation logic is unit-tested in `internal/models/user_test.go`, covering birthday-passed/not-passed, same-day, leap-day, newborn, and future-DOB edge cases.

## Regenerating the DB layer

The files in `db/sqlc/` are generated by SQLC from `sqlc.yaml`, `db/schema.sql`, and `db/queries/users.sql`:

```bash
sqlc generate
```

## Notes on age calculation

`models.CalculateAge(dob, now)` returns full years elapsed, subtracting one if the birthday has not yet occurred in the current year, handling leap-day birthdays, and clamping negative (future-dated) results to `0`. The reference clock is injectable in the service layer for deterministic testing.
