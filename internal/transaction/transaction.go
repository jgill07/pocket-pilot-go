// Package transaction holds the core domain model for pocket-pilot-go.
//
// A Transaction is the single, generic money movement (expense or income).
// This package is the domain layer: it depends only on the standard library
// and pure value libraries (uuid, decimal). It must never import a storage
// implementation, transport, or CLI/HTTP framework — that keeps it testable in
// isolation and reusable behind both the CLI and a future API.
package transaction

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// Type discriminates a transaction. The model carries both kinds; v0 only ever
// constructs Expense — Income exists so the model accommodates it, but nothing
// builds one yet.
type Type string

const (
	Expense Type = "expense"
	Income  Type = "income"
)

// Valid reports whether t is a known transaction type.
func (t Type) Valid() bool {
	switch t {
	case Expense, Income:
		return true
	default:
		return false
	}
}

// Transaction is a single money movement. Amount is always stored positive;
// Type conveys the sign. OccurredAt is user-meaningful (when it happened);
// CreatedAt/UpdatedAt are audit timestamps owned by the code, stamped at
// construction from an injected clock.
type Transaction struct {
	ID          uuid.UUID
	Type        Type
	Description string
	Category    string
	Amount      decimal.Decimal
	Currency    string
	OccurredAt  time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
	CreatedBy   string
}

// CreateParams is the domain's construction contract — the flat set of inputs
// New needs. It deliberately carries no transport tags (no json): the API keeps
// its own request DTO and maps it here, so transport shape never leaks into the
// domain and clients can never set ID/CreatedAt/UpdatedAt.
type CreateParams struct {
	Type        Type
	Description string
	Category    string
	Amount      decimal.Decimal
	Currency    string
	OccurredAt  time.Time
	CreatedBy   string
}

// Validation errors returned by New. Exported so callers (CLI/API) can map them
// to exit codes / HTTP statuses without string matching.
var (
	ErrInvalidType       = errors.New("transaction: invalid type")
	ErrEmptyDescription  = errors.New("transaction: description is required")
	ErrNonPositiveAmount = errors.New("transaction: amount must be greater than zero")
	ErrEmptyCurrency     = errors.New("transaction: currency is required")
)

// New is the only supported way to produce a valid Transaction. It validates
// the params, generates a fresh UUID in the domain layer, sets OccurredAt from
// the params, and stamps CreatedAt/UpdatedAt from the injected now. The clock is
// passed in (never read here) so the domain stays pure and tests are
// deterministic: callers pass time.Now(); tests pass a fixed instant.
//
// On any invalid input it fails closed — returning the zero Transaction and a
// non-nil error — so an invalid Transaction cannot come into existence through
// the supported path.
func New(p CreateParams, now time.Time) (Transaction, error) {
	if err := p.validate(); err != nil {
		return Transaction{}, err
	}
	return Transaction{
		ID:          uuid.New(),
		Type:        p.Type,
		Description: p.Description,
		Category:    p.Category,
		Amount:      p.Amount,
		Currency:    p.Currency,
		OccurredAt:  p.OccurredAt,
		CreatedAt:   now,
		UpdatedAt:   now,
		CreatedBy:   p.CreatedBy,
	}, nil
}

// validate enforces the construction invariants: valid Type, non-empty
// Description, strictly positive Amount, non-empty Currency.
func (p CreateParams) validate() error {
	if !p.Type.Valid() {
		return fmt.Errorf("%w: %q", ErrInvalidType, p.Type)
	}
	if p.Description == "" {
		return ErrEmptyDescription
	}
	if p.Amount.LessThanOrEqual(decimal.Zero) {
		return fmt.Errorf("%w: %s", ErrNonPositiveAmount, p.Amount.String())
	}
	if p.Currency == "" {
		return ErrEmptyCurrency
	}
	return nil
}
