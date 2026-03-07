-- Create bills table
CREATE TABLE bills (
    id              BIGSERIAL PRIMARY KEY,
    uuid            UUID NOT NULL DEFAULT gen_random_uuid() UNIQUE,
    customer_uuid   VARCHAR(36) NOT NULL DEFAULT '',
    currency        VARCHAR(3) NOT NULL,
    status          VARCHAR(20) NOT NULL DEFAULT 'OPEN',
    period_start    TIMESTAMPTZ NOT NULL,
    period_end      TIMESTAMPTZ NOT NULL,
    closed_at       TIMESTAMPTZ,
    total_cents     BIGINT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_bills_uuid ON bills(uuid);
CREATE INDEX idx_bills_customer_id ON bills(customer_uuid);
CREATE INDEX idx_bills_status ON bills(status);
CREATE INDEX idx_bills_period_end ON bills(period_end);
