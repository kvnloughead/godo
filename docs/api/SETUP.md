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

## Pre-commit Hook

Run this once after cloning:

```bash
make setup
```

This points git at the hooks in `.githooks/`. The pre-commit hook runs `scripts/ci.sh`, which is the same set of checks CI runs: `gofmt`, `go vet`, `go build`, and `go test -race`. If a check fails, the commit is aborted. To skip the hook for a single commit, use `git commit --no-verify`.

You can run the checks yourself at any time with `make ci`.

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

   # Mailtrap SMTP settings (for development email testing)
   SMTP_HOST=sandbox.smtp.mailtrap.io
   SMTP_PORT=2525
   SMTP_USERNAME=your_mailtrap_username    # From Mailtrap SMTP Settings
   SMTP_PASSWORD=your_mailtrap_password    # From Mailtrap SMTP Settings
   SMTP_SENDER="Godo <no-reply@godo.example.com>"
   ```

   To get your Mailtrap credentials:

   1. Create a free account at [Mailtrap.io](https://mailtrap.io)
   2. Go to Email Testing → Inboxes
   3. Click on your inbox
   4. Find the SMTP credentials in the "SMTP Settings" section

3. Setup database and run migrations:
   ```bash
   make db/setup
   make db/migrations/up
   ```
