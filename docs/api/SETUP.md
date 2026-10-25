# Setup

There are two ways to run the API locally:

- **With Docker (recommended).** One command starts Postgres, runs the migrations, starts a local mail catcher, and starts the API. You don't need to install Postgres or edit `pg_hba.conf`.
- **Without Docker.** You install and configure Postgres yourself. See [Setup without Docker](#setup-without-docker).

## Setup with Docker

### Prerequisites

- [Docker](https://docs.docker.com/get-docker/) with Compose v2 (`docker compose version` should work). Docker Desktop includes both.
- Go, only if you want to run `make ci`, build the CLI, or run the API outside Docker.

### First run on a new device

```bash
git clone git@github.com:kvnloughead/godo.git
cd godo
make setup                 # install the pre-commit hook (see below)
cp .env.example .env       # dev defaults; edit if you like
docker compose up --build
```

`make docker/up` does the same in the background (and creates `.env` from the example if it doesn't exist).

When it's up:

| What               | Where                                      |
| ------------------ | ------------------------------------------ |
| API                | http://localhost:4000/v1/healthcheck       |
| Mailpit (email UI) | http://localhost:8025                      |
| Postgres           | `localhost:5432`, or just run `make docker/psql` |

All ports are bound to `127.0.0.1`, so nothing is reachable from the rest of your network. If a port is already taken (for example by a Postgres installed on the host), change `POSTGRES_HOST_PORT`, `API_PORT`, or `MAILPIT_UI_PORT` in `.env`.

### What runs

`compose.yaml` defines four services, started in this order:

1. `db`: Postgres 16. Data lives in the `pgdata` named volume, so it survives restarts. The database and user from `.env` are created the first time the volume is empty.
2. `migrate`: a one-shot [golang-migrate](https://github.com/golang-migrate/migrate) container that runs every up migration in `db/migrations` and exits. It waits for `db` to be healthy.
3. `mailpit`: [Mailpit](https://mailpit.axllent.org/), an SMTP server that catches every email the API sends and shows it in a web UI. Registration and activation emails work offline, with no Mailtrap account.
4. `api`: the API, built from the `Dockerfile`. It starts only after `migrate` finishes successfully.

### Trying it out

```bash
curl localhost:4000/v1/healthcheck

curl -X POST localhost:4000/v1/users \
  -d '{"name": "Alice", "email": "alice@example.com", "password": "pa55word1234"}'
```

Open http://localhost:8025 to read the welcome email and get the activation token.

To point the CLI at the local API, run `make cli/setup-local`. It writes `~/.config/godo/settings.local.json` with `http://localhost:4000/v1`.

### Everyday commands

| Command                       | What it does                                                  |
| ----------------------------- | ------------------------------------------------------------- |
| `make docker/up`              | Build and start everything in the background                  |
| `make docker/dev`             | Same, but runs the API under `air` and rebuilds on code changes |
| `make docker/down`            | Stop and remove the containers (the database is kept)         |
| `make docker/logs`            | Follow all logs, or one service: `make docker/logs service=api` |
| `make docker/ps`              | Show containers and their health                              |
| `make docker/psql`            | Open `psql` inside the Postgres container                     |
| `make docker/migrate`         | Apply new up migrations (after `make db/migrations/new`)      |
| `make docker/reset`           | Delete the database volume and start over (asks to confirm)   |

After changing Go code, `make docker/up` rebuilds the image. With `make docker/dev`, the checkout is mounted into a Go container and `air` rebuilds the API when files change. The dev image is defined by the `dev` stage of the `Dockerfile` and `compose.dev.yaml`.

### Configuration

Everything comes from `.env`, which Compose reads automatically. See `.env.example` for each setting. `.env` is gitignored. The API gets its settings as environment variables and flags in the `api` service of `compose.yaml`. Run `make help/web` to see all flags.

### Known issue: email in development mode

When `API_ENV=development` (the default), the API builds the email sender address by running `go list -m`. The runtime image has no Go toolchain, so the sender ends up empty and sending fails with `gomail: invalid address " <no-reply@>"`. Registration still succeeds, but no email reaches Mailpit. Until this is fixed in `internal/injector/config.go`, do one of the following:

- Use `make docker/dev`. The dev image includes Go, so the sender is built correctly.
- Or set `API_ENV=production` in `.env`. The API then uses `SMTP_SENDER`. This also shortens authentication tokens from 28 to 14 days and makes the healthcheck report `production`.

## Pre-commit Hook

Run this once after cloning:

```bash
make setup
```

This points git at the hooks in `.githooks/`. The pre-commit hook runs `scripts/ci.sh`, which is the same set of checks CI runs: `gofmt`, `go vet`, `go build`, and `go test -race`. If a check fails, the commit is aborted. To skip the hook for a single commit, use `git commit --no-verify`.

You can run the checks yourself at any time with `make ci`.

## Setup without Docker

Use this if you'd rather run Postgres and the API directly on your machine.

### Prerequisites

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

### Local Development Setup

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

   Note: the API reads the SMTP host and port only from the `-smtp-host` and `-smtp-port` flags, not from `SMTP_HOST` and `SMTP_PORT`. Pass them with `make FLAGS="-smtp-host=sandbox.smtp.mailtrap.io -smtp-port=2525" run/api`.

3. Setup database and run migrations:
   ```bash
   make db/setup
   make db/migrations/up
   ```

4. Run the API with `make run/api`, or with live reload using `make run/air`.
