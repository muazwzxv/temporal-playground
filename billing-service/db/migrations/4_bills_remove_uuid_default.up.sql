-- Remove auto-generated UUID default from bills table.
-- The uuid is now client-provided (idempotency key from the API).
ALTER TABLE bills ALTER COLUMN uuid DROP DEFAULT;
