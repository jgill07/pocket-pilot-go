package transaction

import (
	"errors"
	"testing"
	"time"

	"github.com/shopspring/decimal"
)

func TestTypeValid(t *testing.T) {
	cases := []struct {
		in   Type
		want bool
	}{
		{Expense, true},
		{Income, true},
		{Type("transfer"), false},
		{Type(""), false},
	}
	for _, c := range cases {
		if got := c.in.Valid(); got != c.want {
			t.Errorf("Type(%q).Valid() = %v, want %v", c.in, got, c.want)
		}
	}
}

// validParams returns a params value that New should accept.
func validParams() CreateParams {
	return CreateParams{
		Type:        Expense,
		Description: "coffee",
		Category:    "food",
		Amount:      decimal.RequireFromString("4.50"),
		Currency:    "USD",
		OccurredAt:  time.Date(2026, 7, 1, 9, 0, 0, 0, time.UTC),
		CreatedBy:   "self",
	}
}

func TestNewHappyPath(t *testing.T) {
	now := time.Date(2026, 7, 15, 12, 0, 0, 0, time.UTC)
	p := validParams()

	tx, err := New(p, now)
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}
	if tx.ID.String() == "00000000-0000-0000-0000-000000000000" {
		t.Error("expected a generated non-nil UUID")
	}
	if !tx.CreatedAt.Equal(now) || !tx.UpdatedAt.Equal(now) {
		t.Errorf("timestamps = (%v, %v), want both %v", tx.CreatedAt, tx.UpdatedAt, now)
	}
	if !tx.OccurredAt.Equal(p.OccurredAt) {
		t.Errorf("OccurredAt = %v, want %v", tx.OccurredAt, p.OccurredAt)
	}
	if tx.Type != p.Type || tx.Description != p.Description ||
		tx.Category != p.Category || tx.Currency != p.Currency ||
		tx.CreatedBy != p.CreatedBy || !tx.Amount.Equal(p.Amount) {
		t.Errorf("field mismatch: got %+v from params %+v", tx, p)
	}
}

func TestNewInjectedClockDeterministic(t *testing.T) {
	now := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)

	a, err := New(validParams(), now)
	if err != nil {
		t.Fatalf("New a: %v", err)
	}
	b, err := New(validParams(), now)
	if err != nil {
		t.Fatalf("New b: %v", err)
	}
	if !a.CreatedAt.Equal(now) || !b.CreatedAt.Equal(now) ||
		!a.UpdatedAt.Equal(now) || !b.UpdatedAt.Equal(now) {
		t.Error("timestamps must come from injected now, identical across calls")
	}
	if a.ID == b.ID {
		t.Error("each call should generate a distinct ID")
	}
}

func TestNewValidation(t *testing.T) {
	now := time.Date(2026, 7, 15, 12, 0, 0, 0, time.UTC)
	cases := []struct {
		name    string
		mutate  func(p *CreateParams)
		wantErr error
	}{
		{"invalid type", func(p *CreateParams) { p.Type = Type("transfer") }, ErrInvalidType},
		{"empty description", func(p *CreateParams) { p.Description = "" }, ErrEmptyDescription},
		{"zero amount", func(p *CreateParams) { p.Amount = decimal.Zero }, ErrNonPositiveAmount},
		{"negative amount", func(p *CreateParams) { p.Amount = decimal.RequireFromString("-1") }, ErrNonPositiveAmount},
		{"empty currency", func(p *CreateParams) { p.Currency = "" }, ErrEmptyCurrency},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			p := validParams()
			c.mutate(&p)

			tx, err := New(p, now)
			if !errors.Is(err, c.wantErr) {
				t.Fatalf("New error = %v, want %v", err, c.wantErr)
			}
			if tx != (Transaction{}) {
				t.Errorf("expected zero Transaction on error, got %+v", tx)
			}
		})
	}
}
