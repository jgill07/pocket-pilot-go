# pocket-pilot-go

Personal expense tracker. CLI-first, API-ready. Log expenses now; income, budgets, and an HTTP API planned later.

The core model is a generic **transaction** (`expense` | `income`) — not separate types — so income and reporting slot in without a rewrite.

## Status

Early. Domain layer only so far:

- `Transaction` model + validated construction (`transaction.New`)
- `Repository` persistence interface (no backend wired up yet)

CLI, Postgres storage, and migrations are next. See [Roadmap](#roadmap).

## Architecture

Domain is decoupled from storage and transport. CLI and the future API both call the same domain layer; storage sits behind an interface so the backend swaps without touching the domain.

```
cmd/            CLI entry (cobra)               — planned
internal/
  transaction/  domain: model + service logic  — no I/O deps
  storage/      Repository impl (pgx)           — planned
  api/          HTTP handlers                   — planned
migrations/     versioned SQL (golang-migrate)  — planned
```

**Rule:** `transaction/` never imports a storage implementation — it depends on the `Repository` interface only.

## Stack

- Go 1.26
- [`shopspring/decimal`](https://github.com/shopspring/decimal) — money (never float; maps to `NUMERIC`)
- [`google/uuid`](https://github.com/google/uuid) — IDs, generated in the domain layer
- Postgres via [`jackc/pgx`](https://github.com/jackc/pgx) — planned
- [`golang-migrate`](https://github.com/golang-migrate/migrate), [`spf13/cobra`](https://github.com/spf13/cobra) — planned

## Develop

```bash
go build ./...
go test ./...
```

## Roadmap

- [x] Transaction domain model + validation
- [x] Repository interface
- [ ] Postgres storage (pgx) + migrations
- [ ] CLI: log + query expenses
- [ ] Income transactions
- [ ] Budgets + reporting
- [ ] HTTP API + auth

## Contributing

Every change lands via a branch + PR — never commit to `main`. Conventional Commits (`type(scope): subject`). See [CLAUDE.md](CLAUDE.md) for full conventions.
