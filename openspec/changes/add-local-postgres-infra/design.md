## Context

The domain layer (`internal/transaction`) is complete: a `Transaction` model and a
`Repository` interface. There is no database and no schema tooling. The next planned
work — a pgx `Repository` implementation — needs a real Postgres to develop and test
against, plus a versioned way to create the schema.

Project constraints (from CLAUDE.md):
- Postgres from day one, behind a `Repository` interface.
- Migrations are versioned SQL files under `migrations/`, never hand-edited live.
- Config via `DATABASE_URL`; never hardcode creds; keep them out of git.
- Designed to become a cloud app later (managed Postgres + CI-run migrations).

## Goals / Non-Goals

**Goals:**
- A one-command local database: `make dev` → Postgres up + schema applied.
- A migration workflow that is an external command, identical in spirit to how CI
  will run it in the cloud later.
- First migration realizing the `transactions` table from the CLAUDE.md data model.
- Keep migrations fully decoupled from the app binary and `go.mod`.

**Non-Goals:**
- pgx `Repository` implementation (next change).
- CLI wiring / app entrypoint.
- GitHub Actions workflow to run migrations in CI (deferred).
- Auth, income transactions, budgets, reporting.
- `updated_at` auto-touch trigger (the app sets `updated_at` on writes).

## Decisions

### Decision: golang-migrate over goose

Chosen because the project's constraints map to golang-migrate's native shape: pure
SQL files, a CLI-first external command, and eventual CI execution as a shell command.

**Alternatives considered:**
- **goose** — its main advantages are a clean Go library API, `embed.FS`, and
  startup/programmatic migration. The project explicitly does NOT want any of those:
  migrations run as an external command, never embedded. Goose's Go-function
  migrations would also tempt violating the "pure SQL" rule. Net: its edge is wasted
  here.

### Decision: `docker-compose.yml` runs Postgres only; no one-shot migrate service

An earlier design considered a one-shot `migrate` service inside compose that waits on
Postgres health and applies migrations. It was dropped: since the developer also wants
`make migrate-up` (which requires the `migrate` binary on the host), the binary is a
dependency regardless — so the one-shot service's only benefit (avoiding the local
binary) evaporates, and it would only add duplication.

Consequences of dropping it:
- **Single migration path** (`make migrate-up`) — no drift between two runners.
- **Single `DATABASE_URL`** (`@localhost`) — eliminates the two-URL problem where an
  in-network service would need `@postgres:5432` while host tools use `@localhost`.
- **compose = pure infrastructure** — mirrors production exactly (managed Postgres +
  CI runs the migrate command; no migrate container in prod either).

`make dev` composes the two steps: `docker compose up -d && make migrate-up`.

**Alternative considered:** compose one-shot migrate service — rejected as above.

### Decision: Timestamps default to `now()` in SQL; `updated_at` touched by the app

`created_at` and `updated_at` are `TIMESTAMPTZ DEFAULT now()` so a row always has valid
audit timestamps even for raw inserts. `updated_at` is set by the application on
updates (no DB trigger) to keep the first migration simple and the write path explicit.
IDs remain generated in the domain layer (UUID), per CLAUDE.md — not by the DB.

### Decision: `DATABASE_URL` via direnv (`.envrc`)

A single `DATABASE_URL` is enough for now. `.envrc` (direnv) exports it for host tools;
`.envrc.example` is the committed template; real `.envrc` is gitignored. Local URL uses
`sslmode=disable` against `localhost:5432`.

### Decision: Index `type` and `created_by` in the first migration

`type` is filtered on constantly (expense vs income). `created_by` is indexed now even
though auth doesn't exist yet: the app is user-scoped by design, so every query filters
by the owning user (`created_by` = `"self"` today, a real user id once auth lands).
Indexing it up front avoids a later migration on a table that only grows. Other
reporting indexes (`occurred_at`, `category`) are deferred until reporting lands, per
CLAUDE.md.

### Decision: Migrations kept transaction-safe

Postgres runs DDL inside transactions, so a failed migration rolls back cleanly and
"dirty" state is a flag-only concern, not partial schema. Migrations MUST avoid
non-transactional statements (e.g. `CREATE INDEX CONCURRENTLY`). The first migration
uses a plain index on `type`, which is transaction-safe.

## Risks / Trade-offs

- **golang-migrate "dirty" state on failed migration** → Postgres transactional DDL
  means the schema itself is unchanged; recovery is `migrate force <last-good>` then
  `migrate up`. Locally, `docker compose down -v` is a faster reset. Mitigated further
  by small, tested, single-purpose migrations.
- **Host tooling dependency** (`migrate` binary + `direnv`) → documented in README
  (`brew install golang-migrate`, direnv setup). Accepted: `make migrate-up` requires
  it regardless.
- **`postgres:17` major pin** → pinned to a major to avoid surprise upgrades; bump
  deliberately. Local volume tied to the PG data format for that major.
- **Two-URL confusion avoided** by dropping the in-network migrate service — only
  `@localhost` exists now.

## Migration Plan

1. Add `docker-compose.yml` (postgres:17, healthcheck, named volume, port map).
2. Add `.envrc.example` (+ gitignore `.envrc`).
3. Add `migrations/000001_create_transactions.{up,down}.sql`.
4. Add `Makefile` targets.
5. Update README (setup) and CLAUDE.md (drop "(or goose)").
6. Verify: `make dev` from a clean state → table + index present; `migrate-down`
   drops cleanly; `down -v` resets.

Rollback: this change is additive tooling/infra with no app code changes; reverting
the files removes it with no data-model impact.

## Open Questions

None outstanding — all forks resolved during exploration (tool choice, single migration
path, timestamp defaults, DB name `pocket_pilot`, GHA deferred).
