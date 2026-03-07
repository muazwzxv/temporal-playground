package utils

import (
	"encore.dev/beta/errs"
)

var (
	ErrInternal = &errs.Error{
		Code:    errs.Internal,
		Message: "INTERNAL",
	}

	ErrUUIDMissing = &errs.Error{
		Code:    errs.InvalidArgument,
		Message: "UUID_MISSING",
	}

	ErrNotFound = &errs.Error{
		Code:    errs.NotFound,
		Message: "NOT_FOUND",
	}
)

// customer API errors
var (
	ErrCustomerNotFoundAPI = &errs.Error{Code: errs.NotFound, Message: "CUSTOMER_NOT_FOUND"}
	ErrEmailAlreadyUsed    = &errs.Error{Code: errs.InvalidArgument, Message: "EMAIL_USED"}
)

// bill API errors
var (
	ErrBillNotFoundAPI      = &errs.Error{Code: errs.NotFound, Message: "BILL_NOT_FOUND"}
	ErrBillClosed           = &errs.Error{Code: errs.FailedPrecondition, Message: "BILL_CLOSED"}
	ErrBillAlreadyClosedAPI = &errs.Error{Code: errs.FailedPrecondition, Message: "BILL_ALREADY_CLOSED"}
	ErrCurrencyMismatch     = &errs.Error{Code: errs.InvalidArgument, Message: "CURRENCY_MISMATCH"}
)

// line item API errors
var (
	ErrLineItemNotFoundAPI   = &errs.Error{Code: errs.NotFound, Message: "LINE_ITEM_NOT_FOUND"}
	ErrAlreadyReversedAPI    = &errs.Error{Code: errs.FailedPrecondition, Message: "ALREADY_REVERSED"}
	ErrCannotReverseReversal = &errs.Error{Code: errs.InvalidArgument, Message: "CANNOT_REVERSE_REVERSAL"}
)

// workflow API errors
var (
	ErrWorkflowNotFound     = &errs.Error{Code: errs.Internal, Message: "WORKFLOW_NOT_FOUND"}
	ErrWorkflowQueryFailed  = &errs.Error{Code: errs.Internal, Message: "WORKFLOW_QUERY_FAILED"}
	ErrWorkflowSignalFailed = &errs.Error{Code: errs.Internal, Message: "WORKFLOW_SIGNAL_FAILED"}
	ErrWorkflowStartFailed  = &errs.Error{Code: errs.Internal, Message: "WORKFLOW_START_FAILED"}
)

// pagination errors
var (
	ErrInvalidCursor = &errs.Error{Code: errs.InvalidArgument, Message: "INVALID_CURSOR"}
)

func ErrValidationFailedWithDetails(details ValidationErrors) *errs.Error {
	return &errs.Error{
		Code:    errs.InvalidArgument,
		Message: "VALIDATION_FAILED",
		Details: details,
	}
}
