# Testing

## Unit tests

```bash
go test ./...
```

or `make ci`, which also runs `gofmt`, `go vet`, and `go build`. These don't need a database. Integration tests are skipped unless `GODO_TEST_DB_DSN` is set, so this always works on a fresh checkout.

## Integration tests

Some tests run against a real Postgres database, because the behavior they check lives in SQL: row locks, transactions, and soft deletes. They are in:

- `internal/server/data`: the per-user seq lock (`seq_test.go`), and the todo write paths (`todos_test.go`), including soft deletes and concurrent writes for one user.
- `cmd/api`: the `/v1/todos` handlers (`todos_handlers_test.go`), including `If-Match` and `409 Conflict`.

They run when `GODO_TEST_DB_DSN` holds the DSN of a database with all migrations applied.

> [!WARNING]
> Use a dedicated test database, not your development database. Each test creates its own users with random emails and deletes them when it finishes, so tests don't interfere with each other or with existing data. But a test that crashes partway can leave rows behind.

### With a local Postgres

```bash
# Create the database. Adjust the user and host for your setup.
createdb -U postgres godo_test

export GODO_TEST_DB_DSN='postgres://postgres@localhost:5432/godo_test?sslmode=disable'

# Apply the migrations. Rerun this after pulling new migrations.
migrate -path ./db/migrations -database "$GODO_TEST_DB_DSN" up

go test -race ./...
```

### With Docker

If you don't have Postgres installed, run a throwaway container on a spare port:

```bash
docker run --rm -d --name godo-test-db -p 5433:5432 \
  -e POSTGRES_USER=godo -e POSTGRES_PASSWORD=godo -e POSTGRES_DB=godo_test \
  postgres:16

export GODO_TEST_DB_DSN='postgres://godo:godo@localhost:5433/godo_test?sslmode=disable'
migrate -path ./db/migrations -database "$GODO_TEST_DB_DSN" up
go test -race ./...

# When you're done:
docker stop godo-test-db
```

### Running a subset

```bash
go test -race -run 'Concurrent' -v ./internal/server/data
go test -race -run 'TestTodoIfMatch' -v ./cmd/api
```

Add `-count=1` to bypass Go's test cache. The cache doesn't know about the database, so a cached pass may be from before you changed a migration.

## CI

`.github/workflows/ci.yml` starts a `postgres:16` service container, applies all migrations, rolls them all back, and applies them again (so a broken down migration fails CI too). It then sets `GODO_TEST_DB_DSN` and runs `scripts/ci.sh`, so the integration tests run on every push and pull request to `main`.
