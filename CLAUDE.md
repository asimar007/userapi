# CLAUDE.md

Guidance for AI assistants (and humans) working in this repository.

## What this is

A small RESTful API for managing users (`name`, `dob`). The defining feature: a
user's **age is never stored** — it is calculated dynamically on read from `dob`
using Go's `time` package. Treat that invariant as load-bearing: do not add an
`age` column or persist age anywhere.

- Module path: `github.com/asimar007/userapi`
- Go version: 1.22
- Database: PostgreSQL (DATE column for `dob`, SERIAL id)

## Tech stack

- GoFiber v2 — HTTP framework
- pgx/v5 (`pgxpool`) — PostgreSQL driver / connection pool
- SQLC — generates the typed DB access layer from SQL
- Uber Zap — structured logging
- go-playground/validator/v10 — request validation
- Lefthook — Git hooks manager (`lefthook.yml`)
- Prettier — formatter for YAML, Markdown, JSON, SQL (`.prettierrc`)

## Architecture & layering

Requests flow strictly downward through these layers; never skip one:

```
handler  ->  service  ->  repository  ->  db/sqlc (generated)
(Fiber)      (logic)      (pgx wrapper)    (SQL)
```

- `internal/handler` — HTTP concerns only: parse/validate input, map errors to
  status codes, shape JSON responses. No business logic, no SQL.
- `internal/service` — business logic: DOB parsing, age calculation, DTO
  mapping. Holds an injectable clock (`s.now`) so tests can freeze time.
- `internal/repository` — thin wrapper over the generated `sqlc.Queries`;
  translates `pgx.ErrNoRows` into the domain `ErrNotFound`.
- `db/sqlc` — GENERATED CODE. Do not hand-edit. See "Database layer" below.
- `internal/models` — DTOs, validation tags, and `CalculateAge`.
- `internal/routes` — route registration + the centralized Fiber error handler.
- `internal/middleware` — request-id injection + request-duration logging.
- `internal/logger` — Zap logger factory.
- `config` — environment-based configuration with defaults.
- `cmd/server` — entrypoint: wiring, DB pool, graceful shutdown.

### Key conventions

- Handlers return `fiber.NewError(status, msg)`; the error handler in
  `internal/routes/routes.go` renders the consistent envelope
  `{ "error": ..., "status": ... }`. Add new error mapping there, not ad hoc.
- Create/Update responses OMIT `age` (use `models.UserResponse`).
  Get/List responses INCLUDE `age` (use `models.UserWithAgeResponse`).
- List pagination uses `?page=` and `?limit=` query params (defaults 1/20,
  max limit 100). The body stays a plain JSON array; pagination metadata goes
  in the `X-Total-Count`, `X-Page`, `X-Limit` response headers. If you change
  this to a wrapped object, update the README and SCREENSHOTS docs too.
- Dates use the layout `2006-01-02` (`models.DateLayout`). Validation tag is
  `datetime=2006-01-02`.

## Database layer (SQLC)

The files in `db/sqlc/*.go` are generated. The sources of truth are:

- `sqlc.yaml` — codegen config (engine: postgresql, sql_package: pgx/v5)
- `db/schema.sql` — the schema SQLC reads
- `db/queries/users.sql` — the named queries

To change DB access: edit `db/queries/users.sql` (and `db/schema.sql` if the
schema changes), then regenerate:

```bash
sqlc generate
```

Migrations live in `db/migrations/` (`*.up.sql` / `*.down.sql`). Keep
`db/schema.sql` in sync with the latest migration so SQLC codegen matches the
real schema.

## Common commands

```bash
# Run everything via Docker (app + postgres + auto-migration)
docker compose up --build
docker compose down            # add -v to wipe the DB volume

# Local run
createdb userdb
psql "$DATABASE_URL" -f db/migrations/000001_create_users_table.up.sql
go mod tidy                    # also generates go.sum
go run ./cmd/server

# Tests
go test ./... -v               # age logic is covered in internal/models/user_test.go

# Codegen
sqlc generate

# Git hooks (run once after cloning)
make hooks                     # installs lefthook into .git/hooks/
lefthook run pre-commit        # manually trigger pre-commit hooks

# Prettier (format non-Go files)
npx prettier --write "**/*.{yml,yaml,md,json,sql}"   # auto-fix
npx prettier --check "**/*.{yml,yaml,md,json,sql}"   # check only

# Make targets
make help                      # build / test / run / migrate-up / docker-up / hooks / etc.
```

## Configuration

Env vars (see `.env.example`). With Docker these are set in
`docker-compose.yml`. Either set `DATABASE_URL` directly or the `DB_*` parts.

| Var                       | Default                        |
| ------------------------- | ------------------------------ |
| `APP_PORT`                | `8080`                         |
| `LOG_LEVEL`               | `info` (debug/info/warn/error) |
| `DATABASE_URL`            | built from `DB_*`              |
| `DB_HOST` / `DB_PORT`     | `localhost` / `5432`           |
| `DB_USER` / `DB_PASSWORD` | `postgres` / `postgres`        |
| `DB_NAME` / `DB_SSLMODE`  | `userdb` / `disable`           |

## When making changes

- Adding an endpoint: define DTOs + validation in `internal/models`, add the
  query in `db/queries/users.sql` and run `sqlc generate`, expose a repo method,
  put logic in the service, add the handler, register the route in
  `internal/routes/routes.go`.
- Keep age calculation in `models.CalculateAge` and keep its unit tests green.
  It must: subtract a year if the birthday hasn't occurred yet this year, handle
  Feb-29 birthdays, and clamp future DOBs to 0.
- Run `go vet ./...` and `go test ./...` before considering a change done.
- Format non-Go files with Prettier before committing (`npx prettier --write`).
  Go files are formatted with `gofmt` only — Prettier ignores them (see
  `.prettierignore`).
- Do not commit secrets; `.env` is gitignored (`.env.example` is the template).

## Git hooks (Lefthook)

`lefthook.yml` defines three hooks that run automatically after `make hooks`:

|Hook|Trigger|Checks|
|---|---|---|
|pre-commit|`git commit`|`go vet`, `gofmt`, `go test`, Prettier|
|pre-push|`git push`|`go test ./... -v`|
|commit-msg|`git commit`|Conventional commit format|

Commit message format: `<type>(<scope>): <subject>`
Allowed types: `feat`, `fix`, `chore`, `docs`, `refactor`, `test`, `style`,
`ci`, `build`, `perf`.

## Known notes / gotchas

- The `db/sqlc/*.go` files were authored to match SQLC v1.27 + pgx/v5 output.
  Running `sqlc generate` should reproduce them; if output differs, trust the
  generator and commit its version.
- `go.sum` is generated on first build (`go mod tidy` or `go mod download`); the
  Dockerfile tolerates its absence and regenerates it during the image build.
