package repository

import "errors"

// Custom repository errors are optional. The generated repository methods
// return raw SQL errors by default for transparency and flexibility.
// You can define domain-specific errors here if you prefer to wrap SQL errors
// with custom errors that are more meaningful to your business logic.
//
// Example usage:
//   - Wrap sql.ErrNoRows with a custom ErrNotFound
//   - Create domain-specific errors like ErrDuplicateEntry, ErrInvalidData, etc.
var (
	ErrNotFound      = errors.New("record not found")
	ErrDatabaseError = errors.New("database error")
)
