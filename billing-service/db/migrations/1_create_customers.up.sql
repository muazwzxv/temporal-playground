-- Create customers table
CREATE TABLE customers (
    id BIGSERIAL PRIMARY KEY,
    uuid VARCHAR(36) NOT NULL UNIQUE,
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Index for looking up customers by UUID
CREATE INDEX idx_customers_uuid ON customers(uuid);

-- Index for looking up customers by email
CREATE INDEX idx_customers_email ON customers(email);
