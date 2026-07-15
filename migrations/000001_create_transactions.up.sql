-- transactions: single-table model for expense | income (see CLAUDE.md).
-- id is generated in the domain layer (UUID), not by the DB — no DEFAULT here.
-- amount is always positive; `type` conveys the sign.
CREATE TABLE transactions (
    id          UUID           PRIMARY KEY,
    type        TEXT           NOT NULL,
    description TEXT           NOT NULL,
    category    TEXT,
    amount      NUMERIC(10, 2) NOT NULL,
    currency    TEXT           NOT NULL,
    occurred_at TIMESTAMPTZ    NOT NULL,
    created_at  TIMESTAMPTZ    NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ    NOT NULL DEFAULT now(),
    created_by  TEXT           NOT NULL
);

-- `type` is queried/filtered constantly (expense vs income); index it.
CREATE INDEX idx_transactions_type ON transactions (type);

-- Every query is scoped to a user (created_by is the user id once auth lands, so filtering by created_by is a constant access pattern.
CREATE INDEX idx_transactions_created_by ON transactions (created_by);
