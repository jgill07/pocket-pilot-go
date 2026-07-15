## 1. Docker Compose (Postgres)

- [x] 1.1 Create `docker-compose.yml` with a single `postgres:17` service: env for
  user/password and `POSTGRES_DB=pocket_pilot`, port map `5432:5432`, named volume
  `pgdata:/var/lib/postgresql/data`, and a `pg_isready` healthcheck.
- [x] 1.2 Verify `docker compose up -d` starts Postgres and it reports healthy;
  confirm data persists across restart and `down -v` resets the volume.
  (Verified live: migration ran, database exists.)

## 2. Environment config

- [x] 2.1 Create `.envrc.example` exporting
  `DATABASE_URL=postgres://<user>:<pass>@localhost:5432/pocket_pilot?sslmode=disable`.
- [x] 2.2 Add `.envrc` to `.gitignore` (keep `.envrc.example` tracked).

## 3. First migration

- [x] 3.1 Create `migrations/000001_create_transactions.up.sql`: `transactions` table
  with id (UUID PK), type (TEXT), description (TEXT), category (TEXT null), amount
  (NUMERIC(10,2)), currency (TEXT), occurred_at (TIMESTAMPTZ), created_at/updated_at
  (TIMESTAMPTZ DEFAULT now()), created_by (TEXT); plus an index on `type` and an
  index on `created_by` (all queries scope to a user). Keep it transaction-safe
  (no `CONCURRENTLY`).
- [x] 3.2 Create `migrations/000001_create_transactions.down.sql` dropping the table
  (and index).

## 4. Makefile

- [x] 4.1 Add `migrate-up`, `migrate-down` (one step), `migrate-create` (name arg),
  and `migrate-force` (version arg) targets using `migrate -path migrations -database
  $$DATABASE_URL ...`.
- [x] 4.2 Add `dev` target: `docker compose up -d && $(MAKE) migrate-up`; add `down`
  target wrapping `docker compose down`.

## 5. Verify workflow end-to-end

- [x] 5.1 From a clean state, run `make dev`; confirm the `transactions` table and the
  `type` index exist (psql `\d transactions`). (Verified live.)
- [x] 5.2 Run `make migrate-down`; confirm the table is dropped. Re-run `make
  migrate-up` to restore. (Verified live.)

## 6. Documentation

- [x] 6.1 Update `README.md` with setup: `brew install golang-migrate`, direnv setup
  (copy `.envrc.example` → `.envrc`, `direnv allow`), and `make dev`.
- [x] 6.2 Update `CLAUDE.md`: golang-migrate chosen — remove the "(or goose)" wording.
