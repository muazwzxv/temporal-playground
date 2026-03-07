package utils

import "fmt"

// ValidationError represents an error returned by API validator
type ValidationError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type ValidationErrors []ValidationError

func (e *ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func (e ValidationErrors) ErrDetails() {}
func (e ValidationError) ErrDetails()  {}

// Validation error
var (
	ErrCustomerNotFound = ValidationError{Code: "CUSTOMER_NOT_FOUND", Message: "Customer not found"}
	ErrInvalidName      = ValidationError{Code: "INVALID_NAME", Message: "Name is required"}
	ErrInvalidEmail     = ValidationError{Code: "INVALID_EMAIL", Message: "Email is required"}

	ErrInvalidCustomerUUID = ValidationError{Code: "INVALID_CUSTOMER_UUID", Message: "Customer UUID is required"}
	ErrInvalidUUID         = ValidationError{Code: "INVALID_UUID", Message: "UUID is required"}
	ErrInvalidCurrency     = ValidationError{Code: "INVALID_CURRENCY", Message: "Currency must be USD or GEL"}
	ErrInvalidPeriodStart  = ValidationError{Code: "INVALID_PERIOD_START", Message: "Period start is required"}
	ErrInvalidPeriodEnd    = ValidationError{Code: "INVALID_PERIOD_END", Message: "Period end is required"}
	ErrInvalidPeriod       = ValidationError{Code: "INVALID_PERIOD", Message: "Period end must be after period start"}

	ErrBillNotFound          = ValidationError{Code: "BILL_NOT_FOUND", Message: "Bill not found"}
	ErrInvalidAmount         = ValidationError{Code: "INVALID_AMOUNT", Message: "Amount must be positive"}
	ErrInvalidFeeType        = ValidationError{Code: "INVALID_FEE_TYPE", Message: "Fee type is required"}
	ErrInvalidIdempotencyKey = ValidationError{Code: "INVALID_IDEMPOTENCY_KEY", Message: "Idempotency key is required"}
	ErrInvalidBillUUID       = ValidationError{Code: "INVALID_BILL_UUID", Message: "Bill UUID is required"}

	// New validation errors for scaffolded endpoints
	ErrInvalidStatus       = ValidationError{Code: "INVALID_STATUS", Message: "Status must be OPEN or CLOSED"}
	ErrInvalidFeeTypeValue = ValidationError{Code: "INVALID_FEE_TYPE_VALUE", Message: "Invalid fee type value"}
	ErrInvalidLineItemUUID = ValidationError{Code: "INVALID_LINE_ITEM_UUID", Message: "Line item UUID is required"}
	ErrLineItemNotFound    = ValidationError{Code: "LINE_ITEM_NOT_FOUND", Message: "Line item not found"}
	ErrAlreadyReversed     = ValidationError{Code: "ALREADY_REVERSED", Message: "Line item already reversed"}
	ErrBillAlreadyClosed   = ValidationError{Code: "BILL_ALREADY_CLOSED", Message: "Bill is already closed"}
)
