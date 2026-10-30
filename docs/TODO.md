# TODO

The design lives in [SPEC.md](SPEC.md). This file tracks the work: the stages that build toward the spec, in order, and a backlog of smaller items. Each stage's number in the [SPEC build order](SPEC.md#build-order) is noted in brackets.

## Roadmap

### Done

- [x] **Stage 0: code moves and contracts** (PR #1). Moved the CLI to `cmd/godo` and added the shared contracts as stubs: `store.Store`, `syncproto`, `merge`, and `sync.Engine`. Also added CI and the pre-commit hook.
- [x] **B: server sync groundwork** (PR #2) [3]. Migration 000006 (`uid`, `seq`, `updated_at`, soft deletes), the per-user `sync_seq` lock, `If-Match` on `PATCH` and `DELETE`, and integration tests against Postgres.
- [x] **C: Docker dev setup** (PR #3) [7, dev half]. Compose stack with Postgres, migrations, Mailpit, and the API, plus `make docker/*`.
- [x] **Parser** (PR #4) [1]. `todo.Parse`, `Line`, and `String`, including `id:` and `mod:`.

### In progress

- [ ] **Web app scaffold** (branch `feat/web-app`) [8]. Vue + TypeScript + Vite in `web/`, the `Backend` interface with an `HttpBackend` stub, the API serving the embedded build, and strict `tsc` and ESLint in CI. See [SPEC: Web app](SPEC.md#web-app).

### Next

- [ ] **Parser follow-up** (owner: Kevin)
  - [ ] Store created and completed dates as tags (for example `created:` and `done:`) instead of todo.txt prefix dates. Keep reading prefix dates for compatibility, and reserve the new keys.
  - [ ] Accept reserved tags anywhere in the line, rejecting only duplicates and bad values. The formatter moves them to the end.
  - [ ] Require tag keys to start with a letter, so `10:30` isn't a tag.
  - [ ] Lowercase error strings and add a sentinel `ErrEmptyDescription`.
  - [ ] Round-trip tests: `Parse(line).String() == line`.
  - [ ] Update [TODOTXT.md](TODOTXT.md) and the SPEC format section.
- [ ] **D: `apistore` + CLI onto `Store`** [2]
  - [ ] Implement `apistore.Store` and move the CLI's API calls behind it (`internal/apiclient`).
  - [ ] Move `ValidateTokenPlaintext` out of `internal/server/data`, so the CLI doesn't link the server's Postgres driver. Add a CI check that nothing under `cmd/godo` imports `internal/server/...`.
  - [ ] Handle activation tokens in the token package.
  - [ ] Return authentication errors from the CLI clearly.
- [ ] **E: `POST /sync`** [4]. Needs the parser and B.
  - [ ] Server transaction in `internal/server/syncsvc`, with integration tests, including concurrent syncs for one user.
  - [ ] The server parses todo.txt lines to fill priority, projects, and contexts. Remove the README limitation once this works.
  - [ ] Filter `GET /v1/todos` by `@context` and `+project`. `GetAll` supports this, but the handler passes `nil`.
- [ ] **F: sync engine + `filestore`** [5]. Needs the parser and `merge.Apply`.
  - [ ] `merge.Resolve` and `merge.Apply` (owner: Kevin).
  - [ ] Diffing, state files, `flock`, and the base snapshot.
  - [ ] Unparseable lines: keep the line, skip it, and warn in `godo status`. Add this rule to the SPEC first.
- [ ] **G: status UX** [6]. `godo status`, offline messages, and the CLI's 2s sync wait.
- [ ] **Web app** [8]
  - [ ] Decide on auth in the browser: bearer token, or an HttpOnly session cookie.
  - [ ] Implement `HttpBackend` and the UI.
  - [ ] CORS, needed only if the frontend is served from another origin: allow `If-Match` in preflights, and use `Vary` with `Add` instead of `Set`.
  - [ ] Give `todo.Todo` JSON tags matching `/v1/todos`, for the Wails adapter.
  - [ ] Wails desktop app wrapping the same frontend.
- [ ] **Deploy to AWS** [7, deploy half]
  - [ ] Root MFA, a budget alarm, and an IAM Identity Center user with `aws configure sso`.
  - [ ] Choose hosting: Lightsail or EC2 running Compose, or EC2 plus RDS.
  - [ ] Hardening, much of which goes away with a fresh setup:
    - Don't expose Postgres or port 4000 publicly.
    - Use TLS for any remote DB connection.
    - Keep secrets in SSM Parameter Store, not in systemd args.
    - Use SSM Session Manager instead of SSH.
    - The CLI should install with an `https://` API URL.
  - [ ] Real SMTP instead of the Mailtrap sandbox.
  - [ ] Remove the old deploy path: the `run/*`, `db/setup`, `build/linux`, and `deploy/gcp` targets, the old sections of `SETUP.md` and `DEPLOYMENT.md`, `godotenv` with `.env.local` and `.env.production`, and `db/scripts/drop.sql`.
  - [ ] Bump the `go` directive in `go.mod` and re-add `go mod tidy -diff` to `scripts/ci.sh`.
  - [ ] Migrate any data from the GCP database with `pg_dump`.
- [ ] **Later** [9]
  - [ ] SSE `GET /events` push.
  - [ ] Field-level merge.
  - [ ] Conflict log (`godo conflicts`). This is an open question in the SPEC.

## Backlog

### CLI

- [ ] Implement `pri`, `depri`, and `update`. They are still Cobra stubs.
- [ ] Show creation and modification dates in `list`.
- [ ] Show due dates (`due:` tags) and sort by them.
- [ ] Variadic arguments for commands, not just in interactive mode.
- [ ] Re-prompt for todos or a command when none is given.
- [ ] Add archived status to `list --plain` output.
- [ ] Document `list` more fully, and DRY up `list.go`.

### Interactive mode

- [ ] Pagination, search, sorting, and filtering.
- [ ] Command history and completion.
- [ ] Generate command help from the command struct tags.
- [ ] Log commands immediately, not after interactive mode exits.
- [ ] Make "unknown command" errors more visible.

### API

- [ ] Verify that every sort, filter, and pagination option of `GET /v1/todos` works, and document them in the README.
- [ ] Log suspicious request patterns separately, and track repeated failed authentication attempts.
