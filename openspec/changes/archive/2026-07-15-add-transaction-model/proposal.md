## Why

pocket-pilot-go currently has only scaffolding — no domain code. Before any CLI
command or storage backend can exist, the app needs its core domain type: the
generic **Transaction**. Everything else (CLI capture, Postgres persistence,
future income/budgets/API) builds on this one model, so it must be defined
cleanly with its invariants enforced from the start.

## What Changes

- Introduce the `transaction` domain package (`internal/transaction/`) with zero
  I/O or framework imports.
- Add the `Transaction` struct with the fields defined in the data model
  (id, type, description, category, amount, currency, occurred_at, created_at,
  updated_at, created_by).
- Add a typed `Type` enum (`type Type string`; `Expense`/`Income` consts;
  `Valid()` method). The model carries both types; v0 only ever constructs
  `expense` — income is not exposed yet but the model accommodates it.
- Add a single construction door: `New(CreateParams, now time.Time)
  (Transaction, error)` that validates input, generates the UUID, and stamps
  timestamps. Both CLI and (future) API build a `CreateParams` and call `New` —
  no other path produces a valid `Transaction`.
- Add the domain-owned `Repository` interface (contract only, no implementation).
  The postgres/pgx impl arrives in a later change.
- Add dependencies: `google/uuid` (ids) and `shopspring/decimal` (money).

Non-goals (explicitly deferred): service layer, storage implementation,
migrations, CLI wiring, HTTP API, income capture.

## Capabilities

### New Capabilities
- `transaction-model`: the core Transaction domain type — its fields, the typed
  `Type` enum, construction/validation invariants via `New`, and the
  domain-owned `Repository` interface contract.

### Modified Capabilities
<!-- none — this is the first capability -->

## Impact

- **New code:** `internal/transaction/` (model, constructor, `Repository`
  interface). No storage, CLI, or API code.
- **Dependencies added:** `github.com/google/uuid`, `github.com/shopspring/decimal`.
- **Layering:** establishes the domain layer that all later layers depend on;
  domain imports neither storage nor transport.
- **No schema/migration impact yet** — the `transactions` table lands with the
  storage change.
