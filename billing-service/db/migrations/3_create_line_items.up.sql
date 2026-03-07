-- Create line items table
CREATE TABLE line_items (
    id              BIGSERIAL PRIMARY KEY,
    uuid            UUID NOT NULL DEFAULT gen_random_uuid() UNIQUE,
    bill_uuid       UUID NOT NULL,
    idempotency_key VARCHAR(255) NOT NULL,
    fee_type        VARCHAR(100) NOT NULL,
    description     TEXT,
    amount_cents    BIGINT NOT NULL,
    reference_uuid  UUID,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT unique_idempotency UNIQUE (bill_uuid, idempotency_key)
);

CREATE INDEX idx_line_items_uuid ON line_items(uuid);
CREATE INDEX idx_line_items_bill_uuid ON line_items(bill_uuid);
CREATE INDEX idx_line_items_fee_type ON line_items(fee_type);
CREATE INDEX idx_line_items_reference_uuid ON line_items(reference_uuid);
