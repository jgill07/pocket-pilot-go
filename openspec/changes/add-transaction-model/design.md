## Context

pocket-pilot-go is greenfield — only scaffolding exists (`go.mod`, docs, worktree
script). This change lands the first domain code: the generic `Transaction`. Per
CLAUDE.md the architecture decouples domain from storage and transport; the
domain layer must be pure (no I/O/framework imports) so both the CLI and a later
HTTP API can reuse it. The data model, money/id/timestamp rules, and the
transaction-vs-DB-transaction naming rule are already fixed in CLAUDE.md and are
treated as constraints here.

## Goals / Non-Goals

**Goals:**
- Define the `Transaction` struct and typed `Type` enum.
- Establish one construction/validation door (`New`) usable identically by CLI
  and API.
- Declare the domain-owned `Repository` interface (contract only).
- Keep the domain package free of I/O and framework imports.

**Non-Goals:**
- Service layer, storage (pgx) implementation, migrations, CLI/API wiring.
- Income capture (model carries `Income`; nothing constructs it in v0).
- Multi-currency logic, reporting/aggregation, auth (`CreatedBy` defaults
  `"self"`).

## Decisions

### Typed `Type` enum over bare string
`type Type string` with `Expense`/`Income` consts and a `Valid()` method.
Centralizes validation, lets the compiler help, and maps cleanly to the TEXT
column. Alternative — raw `string` everywhere — scatters validation and invites
typos. Rejected.

### Single construction door: `New(CreateParams, now time.Time)`
Both edges converge on one function.

```
   CLI (flags)          API (json body → CreateRequest DTO in api/)
        \                      /   maps to CreateParams
         \                    /
          New(CreateParams, now) (Transaction, error)
                     │ validate + uuid + stamp timestamps
                     ▼
                Transaction  (pure domain, no json tags)
```

- `CreateParams` is the domain's construction contract (plain struct, no json
  tags): `Type, Description, Category, Amount, Currency, OccurredAt, CreatedBy`.
- The API defines its **own** json-tagged `CreateRequest` in `api/` and maps it
  to `CreateParams`. It never unmarshals into `Transaction`. Rationale: unmarshal
  bypasses the constructor (invalid model could exist), leaks transport tags into
  the domain, and would let a client set `ID`/`CreatedAt`/`CreatedBy` (audit &
  integrity hole). The apparent "two structs for one thing" is the boundary/DTO
  pattern and is intentional.
- Rejected alternative: functional options (`New(desc, amt, opts...)`). Fields
  are mostly required and few are optional; a flat params struct is clearer with
  no variadic ceremony.

### Injected clock (approach A) over inline `time.Now()`
`now time.Time` is passed explicitly to `New`, which stamps `CreatedAt` and
`UpdatedAt`. CLI/API pass `time.Now()`; tests pass a fixed instant → deterministic
and keeps the domain pure (it never reads the wall clock). Inline `time.Now()`
was rejected: flaky-ish tests and hidden dependency. `OccurredAt` is
user-meaningful and comes from `CreateParams`; `CreatedAt`/`UpdatedAt` are audit
timestamps owned by code here (not DB defaults) so the value is consistent across
backends.

### Fail-closed construction
`New` returns `(Transaction, error)`; on any invalid input it returns the zero
`Transaction` and a non-nil error. An invalid `Transaction` cannot come into
existence through the supported path. Invariants: valid `Type`, non-empty
`Description`, `Amount > 0`, non-empty `Currency`. Amount always positive; `Type`
gives sign.

### Money & IDs
`shopspring/decimal` for `Amount` (maps to `NUMERIC(10,2)`, never float);
`google/uuid` generated in the domain (not the DB). Both are pure libraries and
allowed in the domain package.

### Domain-owned `Repository` interface, contract only
The interface lives in the domain package so the domain depends on abstraction,
not a backend. The pgx implementation is a later change and a new file — nothing
in the domain changes when a backend is added/swapped. Returned slices are
initialized with `make` (never nil).

### Naming
Per CLAUDE.md, a DB `BEGIN/COMMIT` transaction is always aliased `dbTx` to avoid
shadowing the domain `Transaction`. No DB code in this change, but the rule is
recorded for the storage change that follows.

## Risks / Trade-offs

- **Repository interface shape guessed before real query needs** → keep it
  minimal in this change (the operations the CLI capture path needs); widen when
  the storage/reporting changes land. Interface, so additive.
- **Two structs (`CreateParams` + API `CreateRequest`) read as duplication** →
  documented as the intentional boundary pattern; the API DTO does not exist yet
  in this change, so no cost is paid until the API lands.
- **Injected `now` slightly more verbose at call sites** → accepted; determinism
  and purity outweigh one extra argument.

## Migration Plan

No data or schema migration — no persistence yet. Additive new package; nothing
to roll back beyond removing the package. The `transactions` table and its
migration arrive with the storage change.

## Open Questions

- Exact method set on `Repository` (e.g. `Save`, `Get`, `List`) — settle to the
  minimum the first CLI capture command needs, in the storage change.
- Whether a `Signed()` helper (negative amount for expense) belongs in the domain
  or the reporting layer — deferred until reporting exists.
