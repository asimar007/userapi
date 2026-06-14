.PHONY: help run build test tidy sqlc migrate-up migrate-down docker-up docker-down hooks

DB_URL ?= postgres://postgres:postgres@localhost:5432/userdb?sslmode=disable

help:
	@echo "Targets:"
	@echo "  run           Run the API server locally"
	@echo "  build         Build the server binary into ./server"
	@echo "  test          Run unit tests"
	@echo "  tidy          go mod tidy"
	@echo "  sqlc          Regenerate the DB layer from db/queries (requires sqlc)"
	@echo "  migrate-up    Apply the schema migration (requires psql)"
	@echo "  migrate-down  Drop the schema (requires psql)"
	@echo "  docker-up     Start app + postgres via docker compose"
	@echo "  docker-down   Stop docker compose stack"
	@echo "  hooks         Install lefthook git hooks"

run:
	go run ./cmd/server

build:
	CGO_ENABLED=0 go build -o server ./cmd/server

test:
	go test ./... -v

tidy:
	go mod tidy

sqlc:
	sqlc generate

migrate-up:
	psql "$(DB_URL)" -f db/migrations/000001_create_users_table.up.sql

migrate-down:
	psql "$(DB_URL)" -f db/migrations/000001_create_users_table.down.sql

docker-up:
	docker compose up --build

docker-down:
	docker compose down

hooks:
	go install github.com/evilmartians/lefthook@latest
	lefthook install
