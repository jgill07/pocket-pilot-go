# transaction-model

## Purpose

Define the core `Transaction` domain type and its construction rules for the
pocket-pilot expense tracker. The model is a single generic transaction
(`expense` | `income`) that is decoupled from storage and transport: the domain
package owns the type, its validation invariants, the single construction door,
and the `Repository` interface it depends on. v0 only constructs `expense`
transactions, but the model accommodates `income` from the start.

## Requirements

### Requirement: Transaction domain type

The system SHALL provide a `Transaction` domain type in the `internal/transaction/`
package carrying the fields: `ID` (UUID), `Type`, `Description`, `Category`,
`Amount`, `Currency`, `OccurredAt`, `CreatedAt`, `UpdatedAt`, and `CreatedBy`.
The domain package SHALL NOT import any storage implementation, transport, or
framework packages.

#### Scenario: Domain package has no I/O imports

- **WHEN** the `internal/transaction/` package is compiled
- **THEN** it imports only the standard library and pure domain dependencies
  (`google/uuid`, `shopspring/decimal`)
- **AND** it imports no storage, HTTP, or CLI framework package

### Requirement: Typed transaction Type enum

The system SHALL represent transaction type as a typed enum `type Type string`
with constants `Expense` (`"expense"`) and `Income` (`"income"`), exposing a
`Valid()` method that reports whether the value is one of the known types. The
model SHALL accommodate both types even though v0 only constructs `expense`.

#### Scenario: Known types are valid

- **WHEN** `Valid()` is called on `Expense` or `Income`
- **THEN** it returns `true`

#### Scenario: Unknown type is invalid

- **WHEN** `Valid()` is called on any value other than `Expense` or `Income`
- **THEN** it returns `false`

### Requirement: Single construction door

The system SHALL provide `New(p CreateParams, now time.Time) (Transaction, error)`
as the only supported way to produce a valid `Transaction`. `New` SHALL validate
the input, generate the `ID` as a new UUID in the domain layer, and stamp
`CreatedAt` and `UpdatedAt` from the supplied `now` (the clock is injected, never
read inside the domain). `OccurredAt` SHALL be taken from `CreateParams`.
Both the CLI and the future API SHALL construct a `CreateParams` and call `New`;
no other construction path is supported.

#### Scenario: Valid params produce a Transaction

- **WHEN** `New` is called with valid `CreateParams` and a `now` timestamp
- **THEN** it returns a `Transaction` with a freshly generated non-nil `ID`
- **AND** `CreatedAt` and `UpdatedAt` equal `now`
- **AND** `OccurredAt` equals the value from `CreateParams`
- **AND** the returned error is nil

#### Scenario: Timestamps come from the injected clock

- **WHEN** `New` is called twice with the same fixed `now`
- **THEN** both returned transactions have `CreatedAt` and `UpdatedAt` equal to
  that fixed `now`
- **AND** the domain does not call the system clock itself

### Requirement: Construction validation invariants

`New` SHALL reject invalid input by returning a non-nil error and the zero
`Transaction`. It SHALL enforce: `Type` must be valid; `Description` must be
non-empty; `Amount` must be strictly greater than zero; `Currency` must be
non-empty. Amount is always stored positive — the `Type` conveys sign.

#### Scenario: Invalid type rejected

- **WHEN** `New` is called with a `Type` that is not `Expense` or `Income`
- **THEN** it returns a non-nil error and no `Transaction` is produced

#### Scenario: Empty description rejected

- **WHEN** `New` is called with an empty `Description`
- **THEN** it returns a non-nil error

#### Scenario: Non-positive amount rejected

- **WHEN** `New` is called with an `Amount` of zero or a negative value
- **THEN** it returns a non-nil error

#### Scenario: Empty currency rejected

- **WHEN** `New` is called with an empty `Currency`
- **THEN** it returns a non-nil error

### Requirement: Domain-owned Repository interface

The system SHALL define a `Repository` interface in the domain package
describing persistence operations for transactions. The domain SHALL depend on
this interface only; no implementation is provided in this change. Any returned
collection of transactions SHALL be a non-nil slice (initialized with `make`).

#### Scenario: Repository is an interface with no domain-side implementation

- **WHEN** the domain package is inspected
- **THEN** `Repository` is declared as an interface
- **AND** the domain package contains no concrete storage implementation of it
