## 1. Dependencies

- [x] 1.1 `go get github.com/google/uuid`
- [x] 1.2 `go get github.com/shopspring/decimal`
- [x] 1.3 `go mod tidy` and confirm `go.mod`/`go.sum` are clean

## 2. Type enum

- [x] 2.1 Create `internal/transaction/transaction.go` with `type Type string` and `Expense`/`Income` consts
- [x] 2.2 Add `func (t Type) Valid() bool` returning true only for `Expense`/`Income`

## 3. Transaction model + construction

- [x] 3.1 Define the `Transaction` struct (ID, Type, Description, Category, Amount, Currency, OccurredAt, CreatedAt, UpdatedAt, CreatedBy) using `uuid.UUID`, `decimal.Decimal`, `time.Time`
- [x] 3.2 Define `CreateParams` (Type, Description, Category, Amount, Currency, OccurredAt, CreatedBy) — no json tags
- [x] 3.3 Implement `New(p CreateParams, now time.Time) (Transaction, error)`: validate, generate `uuid.New()` for ID, set `OccurredAt` from params, stamp `CreatedAt`/`UpdatedAt` from `now`
- [x] 3.4 Implement validation: valid `Type`, non-empty `Description`, `Amount` > 0, non-empty `Currency`; return zero `Transaction` + non-nil error on failure

## 4. Repository interface

- [x] 4.1 Create `internal/transaction/repository.go` with the domain-owned `Repository` interface (minimal method set for capture; document that list results use `make`, never nil)
- [x] 4.2 Confirm no implementation is added and the domain imports no storage/transport/framework package

## 5. Tests

- [x] 5.1 Test `Type.Valid()` for known and unknown values
- [x] 5.2 Test `New` happy path: non-nil ID, `CreatedAt`/`UpdatedAt` == fixed `now`, `OccurredAt` preserved
- [x] 5.3 Test `New` rejects invalid type, empty description, non-positive amount, empty currency (returns error + zero value)
- [x] 5.4 Test injected clock: two calls with same fixed `now` produce identical timestamps

## 6. Verify

- [x] 6.1 `gofmt`/`go vet` clean; `go build ./...` and `go test ./...` pass
- [x] 6.2 Confirm no discarded errors and no nil slice returns (per conventions)
