# Setup

## Prerequisites

1. Install PostgreSQL and the `psql` CLI tool
2. Install the `migrate` CLI tool:
   ```bash
   go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
   ```
3. Install `air`:
   This enables the `make run/air` target to automatically restart the API
   server when code changes are detected.
   ```bash
   go install github.com/air-verse/air@latest
   ```

## Local Development Setup

1. Configure PostgreSQL authentication in `/etc/postgresql/*/main/pg_hba.conf`:

   ```
   # PostgreSQL superuser access for local development
   local   all             postgres                                trust
   host    all             postgres        127.0.0.1/32            trust
   host    all             postgres        ::1/128                 trust
   ```

2. Create `.env.local` file:

   ```env.local
   DB_NAME=godo_dev_db
   DB_USER=godo_dev_user
   DB_PASSWORD=your_dev_password
   DB_HOST=localhost
   DB_PORT=5432
   DB_DSN=postgresql://${DB_USER}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${DB_NAME}?sslmode=disable
   EDITOR=vim
   ```

3. Setup database and run migrations:
   ```bash
   make db/setup
   make db/migrations/up
   ```
