set dotenv-load

migrations_dir := "./migrations"

export GOOSE_DRIVER := "postgres"
db_connection := "postgres://" \
    + env("DB_USER", "test1") \
    + ":" + env("DB_PASSWORD", "test") \
    + "@" + env("DB_HOST", "localhost") \
    + ":" + env("DB_PORT", "5433") \
    + "/" + env("DB_NAME", "test1") \
    + "?sslmode=disable"
export GOOSE_MIGRATION_DIR := migrations_dir

export GOOSE_DBSTRING := db_connection

goose := "goose"

dev:
    @go run ./cmd/*.go

build:
    @go build -o bin/app ./cmd

live:
    @watchexec --restart -- go run ./cmd/*.go

generate:
    @sqlc generate

db-up:
    @{{ goose }} up

[confirm("Rollback migrations down by one version? (y/N): ")]
db-down:
	@{{ goose }} down

[confirm("Rollback migrations down to the specified version? (y/N): ")]
db-down-to version:
	@{{ goose }} down-to {{version}}

db-status:
	@{{ goose }} status

db-create name:
	@{{ goose }} -s create {{name}} sql

test:
	@go test ./... -v
