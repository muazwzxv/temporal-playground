package user

import "fmt"

// PermanentError represents a non-retryable error.
// Use this for validation failures, business rule violations,
// duplicate entries, or any error that won't succeed on retry.
type PermanentError struct {
	Code    string
	Message string
}

func (e *PermanentError) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// NewPermanentError creates a new PermanentError with the given code and message.
func NewPermanentError(code, message string) *PermanentError {
	return &PermanentError{Code: code, Message: message}
}

// IsPermanent checks if an error is a PermanentError.
func IsPermanent(err error) bool {
	_, ok := err.(*PermanentError)
	return ok
}

// Common error codes for user creation
const (
	ErrCodeValidation    = "VALIDATION_ERROR"
	ErrCodeDuplicateUser = "DUPLICATE_USER"
	ErrCodeCreateFailed  = "CREATE_FAILED"
)
