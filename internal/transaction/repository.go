package transaction

import (
	"context"

	"github.com/google/uuid"
)

// Repository is the persistence contract for transactions, owned by the domain.
// The domain depends on this interface only; the concrete backend (postgres/pgx)
// is a separate implementation in the storage package and arrives in a later
// change. Adding or swapping a backend is a new impl of this interface —
// nothing in the domain changes.
//
// Implementations must never return a nil slice from List; return an empty,
// make-initialized slice when there are no results (per project conventions).
type Repository interface {
	// Save persists a transaction (insert or update by ID).
	Save(ctx context.Context, t Transaction) error
	// Get returns the transaction with the given ID.
	Get(ctx context.Context, id uuid.UUID) (Transaction, error)
	// List returns stored transactions. Never returns a nil slice.
	List(ctx context.Context) ([]Transaction, error)
}
