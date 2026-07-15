# pocket-pilot-go

Personal expense tracker. Log expenses now; income + budgets planned later.
Core model is a generic **transaction** (expense | income), not separate types.

Module: `github.com/jgill07/pocket-pilot-go` · Go 1.26.

## What it is

CLI-first, API-ready. Start as a CLI for logging/querying expenses; promote the
same domain layer behind an HTTP API later. Single-user today, designed so it
could serve multiple users (auth) down the road.

## Scope

- **Now (v0):** capture + query transactions of type `expense`.
- **Later:** income transactions, budgets, HTTP API, auth, cloud storage.
- Build v0 clean but leave seams for later — do not build the later features now.
  The transaction model already accommodates income; v0 just doesn't expose it.

## Architecture

Domain is decoupled from storage and transport. Both CLI and (future) API call
the same service; storage sits behind an interface so the backend is swappable.

```
cmd/            CLI entry (cobra)
internal/
  transaction/  domain: Transaction model + service logic (no I/O deps)
  storage/      Repository interface + postgres (pgx) implementation
  api/          HTTP handlers (later — reuse the same service)
migrations/     versioned SQL migrations (golang-migrate)
```

Rule: `transaction/` domain never imports a storage implementation — it depends
on the `Repository` interface only. Swapping/adding a backend = a new impl file,
nothing else changes.

## Storage

- **Postgres from day one**, behind a `Repository` interface.
- Local dev via Docker (`docker-compose` with a `postgres` service). Cloud later
  = point at RDS/Neon/Supabase; no engine switch, no dialect surprises.
- Relational on purpose — income + budgets + reporting need aggregation
  (sum by category, by month, net income−expense).
- Driver: **`jackc/pgx`** (v5). Migrations: **`golang-migrate`** (or goose) —
  versioned SQL files under `migrations/`, never hand-edit the live schema.

### Postgres notes

- Native types — use them: money → `NUMERIC`, ids → `UUID`, timestamps →
  `TIMESTAMPTZ` (store UTC), bool → `BOOLEAN`.
- Config via env / `DATABASE_URL`. Never hardcode creds; keep them out of git.

## Data model — `transactions`

Single table, single-table inheritance. `type` discriminates; type-specific
fields are **sparse nullable columns** added as needed (no JSON blob).

| field         | type         | expense | income | notes                                     |
|---------------|--------------|:-------:|:------:|-------------------------------------------|
| id            | UUID         |    ✓    |   ✓    | `google/uuid`, generated in domain layer  |
| type          | TEXT         |    ✓    |   ✓    | `expense`\|`income`, **indexed**; validated in domain |
| description   | TEXT         |    ✓    |   ✓    | required                                   |
| category      | TEXT         |    ✓    |   ✓    | optional but desired                       |
| amount        | NUMERIC(10,2)|    ✓    |   ✓    | always positive; `type` gives sign         |
| currency      | TEXT         |    ✓    |   ✓    | single currency for now, room to grow      |
| occurred_at   | TIMESTAMPTZ  |    ✓    |   ✓    | when the transaction should/did take place |
| created_at    | TIMESTAMPTZ  |    ✓    |   ✓    | when the row was created in the DB          |
| updated_at    | TIMESTAMPTZ  |    ✓    |   ✓    | last DB update                             |
| created_by    | TEXT         |    ✓    |   ✓    | default "self" now; real user id once auth |

- **`occurred_at` vs `created_at`:** `occurred_at` = when the transaction
  should/did happen (user-meaningful, can be past or future). `created_at` = when
  the row landed in the DB (system audit). They are different on purpose.
- **Money:** `shopspring/decimal` in Go ↔ `NUMERIC(10,2)` column. Never float.
  Scale 2 = USD cents; precision 10 = ~100M ceiling (ample for personal, no
  waste — NUMERIC is variable-length so headroom is free). Bump precision only if
  this ever tracks org/business money. Amount always positive; `type` gives sign.
  Widening later is a safe one-line migration (`ALTER COLUMN amount TYPE
  NUMERIC(14,2)`) — no data loss; rewrites the table but trivial at personal scale.
- **IDs:** `UUID` column, value generated in the domain layer, not the DB.
- **Sparse growth:** type-only fields (e.g. expense `merchant`, income `source`)
  → new nullable columns, not a metadata blob. Keeps them queryable/indexable.
- Index on `type`; add `occurred_at` / `category` indexes when reporting lands.

### Naming — transaction collision

Domain type is **`Transaction`** (table `transactions`). "Transaction" also means
a DB `BEGIN/COMMIT`. **Always alias the DB transaction as `dbTx`** (never a bare
`tx` that shades the domain type) to keep the two unambiguous.

## Conventions

- **Errors:** never discard. Handle or at least log — including deferred
  `Close()`. No bare `_ = err`.
- **Slices:** initialize returned slices with `make()`; never return a nil slice.
- **Layering:** domain has no I/O/framework imports. Keep it testable in isolation.
- Match surrounding style; standard Go idioms; `gofmt`.

## Git workflow

Every fix/feature follows this flow:

1. **Branch + worktree — always via `scripts/worktree.sh`.** Never run raw
   `git worktree add` / `git checkout -b` for new work; the script is the single
   entry point. It validates `<type>`, creates branch `type/desc` + worktree
   `.worktrees/type-desc`, and copies `.env` in.

   ```bash
   scripts/worktree.sh <type> <desc>     # e.g. scripts/worktree.sh feat add-transaction
   git worktree remove .worktrees/<type>-<desc>   # clean up after merge
   ```

   - `<type>`: one of `feat fix docs refactor test chore build ci perf style`
     (invalid type is rejected). `<desc>`: short kebab-case.
   - Worktrees live **inside the repo** under `.worktrees/` (gitignored) — sibling
     dirs (`../name`) are blocked by the sandbox. They share `.git` but start
     clean; all share the same local Postgres.
2. **Commit** with Conventional Commits: `type(scope): subject` (imperative,
   ≤50 chars; body = why, not what, wrapped ~72). No `Co-Authored-By: Claude`.
   - types: `feat` `fix` `docs` `refactor` `test` `chore` `build` `ci` `perf` `style`
   - scopes: `transaction` `storage` `api` `cli` `migrations` `db`
   - **Show the message and wait for explicit approval before committing.**
3. **PR, never push to `main`.** On a `gh` repo-not-found error, flag it

## Suggested deps

- `spf13/cobra` — CLI
- `jackc/pgx` (v5) — Postgres driver
- `golang-migrate/migrate` — schema migrations (SQL files in `migrations/`)
- `google/uuid` — IDs
- `shopspring/decimal` — money (maps to `NUMERIC`)
