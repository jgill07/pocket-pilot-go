## ADDED Requirements

### Requirement: Local Postgres via Docker Compose

The system SHALL provide a `docker-compose.yml` that runs a single `postgres:17`
service for local development. The service MUST expose port 5432 on the host, persist
data in a named volume, and declare a healthcheck so dependents can wait for
readiness. The compose file MUST NOT run migrations or the application — it provides
infrastructure only.

#### Scenario: Postgres boots and accepts connections

- **WHEN** a developer runs `docker compose up -d`
- **THEN** a `postgres:17` container starts, listens on `localhost:5432`, and reports
  healthy via its healthcheck once ready to accept connections

#### Scenario: Data persists across restarts

- **WHEN** the developer stops and restarts the postgres service without removing the
  named volume
- **THEN** previously written data is still present

#### Scenario: Volume removal resets the database

- **WHEN** the developer runs `docker compose down -v`
- **THEN** the named volume is removed and the next `up` starts with an empty database

### Requirement: Command-driven schema migrations

The system SHALL manage schema changes with golang-migrate using versioned SQL files
under `migrations/`. Migrations MUST be run as an external command and MUST NOT be
embedded in the application binary or added as a Go module dependency. Every migration
MUST provide both an `up` and a `down` file using sequential numeric prefixes.

#### Scenario: Applying migrations creates the schema

- **WHEN** the developer runs the migrate-up command against an empty database
- **THEN** all pending migrations apply in order and golang-migrate records the
  applied version in its tracking table

#### Scenario: Rolling back a migration

- **WHEN** the developer runs the migrate-down command for one step
- **THEN** the most recent migration's `down` file runs and the recorded version is
  decremented

#### Scenario: Recovering from a failed (dirty) migration

- **WHEN** a migration fails and golang-migrate marks the database dirty
- **THEN** the developer can fix the SQL, force the version to the last good version,
  and re-run migrate-up to apply the corrected migration
- **AND** locally the developer may instead run `docker compose down -v` to reset

### Requirement: Make targets orchestrate the local workflow

The system SHALL provide a `Makefile` exposing `dev`, `migrate-up`, `migrate-down`,
`migrate-create`, `migrate-force`, and `down` targets. `make dev` MUST boot Postgres
and apply migrations in a single command. Migration targets MUST read the database URL
from the `DATABASE_URL` environment variable.

#### Scenario: One command boots DB and schema

- **WHEN** a developer runs `make dev`
- **THEN** Postgres is started in the background and all migrations are applied,
  leaving a ready-to-use database

#### Scenario: Creating a new migration

- **WHEN** a developer runs `make migrate-create name=<desc>`
- **THEN** a new sequentially-numbered `up`/`down` SQL file pair is scaffolded under
  `migrations/`

### Requirement: Database connection configured via DATABASE_URL

The system SHALL configure the database connection through a single `DATABASE_URL`
environment variable, managed locally with direnv. The repository MUST commit a
`.envrc.example` template and MUST NOT commit the real `.envrc`.

#### Scenario: Local environment supplies DATABASE_URL

- **WHEN** a developer copies `.envrc.example` to `.envrc` and allows direnv
- **THEN** `DATABASE_URL` is exported pointing at `localhost:5432` and the migrate
  commands connect using it

#### Scenario: Real env file stays out of version control

- **WHEN** the repository is inspected
- **THEN** `.envrc` is gitignored and only `.envrc.example` is tracked

### Requirement: First migration creates the transactions table

The system SHALL include a first migration that creates the `transactions` table
matching the domain data model. The table MUST include `created_at` and `updated_at`
columns typed `TIMESTAMPTZ` defaulting to `now()` (UTC), an index on `type`, and an
index on `created_by` (every query is scoped to a user). The migration MUST be
transaction-safe (no non-transactional statements). The `down` migration MUST drop the
table.

#### Scenario: Up migration creates the table and indexes

- **WHEN** the first migration is applied
- **THEN** a `transactions` table exists with columns for id, type, description,
  category, amount, currency, occurred_at, created_at, updated_at, created_by
- **AND** an index on `type` exists
- **AND** an index on `created_by` exists
- **AND** `created_at` and `updated_at` default to `now()`

#### Scenario: Down migration drops the table

- **WHEN** the first migration is rolled back
- **THEN** the `transactions` table no longer exists
