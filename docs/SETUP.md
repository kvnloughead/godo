# Setup

## Database

Prerequisites:

1. Install the `psql` CLI tool.
2. Install the `migrate` CLI tool.

   ```bash
   go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
   ```

Setup steps:

1. Set the environment variables in the `.env` file.

   ```bash
   DB_USER=fill_me_in
   DB_PASSWORD=fill_me_in
   DB_NAME=fill_me_in
   DB_HOST=localhost
   DB_PORT=5432
   DB_DSN=postgresql://${DB_USER}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${DB_NAME}?sslmode=disable
   EDITOR=vim
   ```

2. Run `make db/setup` to create the database and user.
3. Run `make db/migrations/up` to apply all migrations.
