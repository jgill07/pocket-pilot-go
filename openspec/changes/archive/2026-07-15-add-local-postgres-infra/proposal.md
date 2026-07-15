## Why

The domain layer (`Transaction` model + `Repository` interface) exists, but there is
no database to run against. Before writing the pgx `Repository` implementation we
need a real Postgres to develop and test against, plus a repeatable, versioned way to
create and evolve the schema. This change stands up local Postgres infrastructure and
the migration workflow so the storage layer has a target.

## What Changes

- Add `docker-compose.yml` with a **postgres:17** service only (healthcheck, named
  volume, port 5432). Infrastructure only — no schema, no app.
- Adopt **golang-migrate** as the migration tool. Migrations are versioned SQL files
  under `migrations/`, run as an **external command** — never embedded in the app
  binary and never added to `go.mod`.
- Add a `Makefile` with `dev`, `migrate-up`, `migrate-down`, `migrate-create`,
  `migrate-force`, and `down` targets. `make dev` boots the DB and applies migrations
  in one command (`docker compose up -d && make migrate-up`).
- Add the first migration `000001_create_transactions.{up,down}.sql` creating the
  `transactions` table per the CLAUDE.md data model, with `created_at`/`updated_at`
  as `TIMESTAMPTZ DEFAULT now()` and an index on `type`.
- Add `.envrc` (direnv, gitignored) + committed `.envrc.example` exporting a single
  `DATABASE_URL` pointing at `localhost:5432`.
- Update README with setup steps (`brew install golang-migrate`, direnv, `make dev`).
- Update CLAUDE.md: golang-migrate is chosen — drop the "(or goose)" wording.

Deferred (not in this change): pgx `Repository` implementation, CLI wiring, GitHub
Actions workflow to run migrations in CI/cloud.

## Capabilities

### New Capabilities
- `local-postgres-infra`: Local Postgres via Docker Compose and a versioned,
  command-driven schema migration workflow (golang-migrate) that mirrors the intended
  cloud model — infrastructure separate from schema, migrations always an external
  command.

### Modified Capabilities
<!-- None. The transaction-model spec is unchanged; the first migration realizes that
     model in SQL but does not alter its domain requirements. -->

## Impact

- **New files:** `docker-compose.yml`, `Makefile`, `migrations/000001_create_transactions.up.sql`,
  `migrations/000001_create_transactions.down.sql`, `.envrc.example`.
- **Gitignored:** `.envrc` (local secrets/URL).
- **Docs:** `README.md`, `CLAUDE.md` updated.
- **Tooling dependency:** developers install the `migrate` CLI binary and `direnv`
  locally. No new Go module dependency.
- **No code changes** to `internal/transaction/`; unblocks the future storage impl.
