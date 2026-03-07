package entity

// =============================================================================
// Bill Status (persisted in DB: OPEN, CLOSED only)
// =============================================================================

// BillStatus represents the lifecycle status of a bill
type BillStatus string

const (
	// BillStatusOpen - Bill is active, accepting line items
	BillStatusOpen BillStatus = "OPEN"

	// BillStatusClosing - Close signal sent, workflow processing
	// NOTE: API-only status, not persisted to database
	BillStatusClosing BillStatus = "CLOSING"

	// BillStatusClosed - Bill finalized, no modifications allowed
	BillStatusClosed BillStatus = "CLOSED"
)

// IsValid checks if the status is a valid bill status (for filtering)
func (s BillStatus) IsValid() bool {
	return s == BillStatusOpen || s == BillStatusClosed
}

// String returns the string representation of the status
func (s BillStatus) String() string {
	return string(s)
}

// =============================================================================
// Line Item Response Status (API-only, not persisted)
// =============================================================================

// LineItemResponseStatus represents the status in API responses
type LineItemResponseStatus string

const (
	// LineItemStatusPending - Signal sent to workflow, awaiting persistence
	LineItemStatusPending LineItemResponseStatus = "pending"

	// LineItemStatusPersisted - Already exists in database
	LineItemStatusPersisted LineItemResponseStatus = "persisted"
)

// String returns the string representation of the status
func (s LineItemResponseStatus) String() string {
	return string(s)
}

// =============================================================================
// Fee Types (strict enum)
// =============================================================================

// FeeType represents the type of fee charged
type FeeType string

const (
	FeeTypeWireTransfer FeeType = "WIRE_TRANSFER"
	FeeTypeACH          FeeType = "ACH"
	FeeTypeMonthlyFee   FeeType = "MONTHLY_FEE"
	FeeTypeReversal     FeeType = "REVERSAL"
	FeeTypeOther        FeeType = "OTHER"
)

// ValidFeeTypes is the list of allowed fee types
var ValidFeeTypes = []FeeType{
	FeeTypeWireTransfer,
	FeeTypeACH,
	FeeTypeMonthlyFee,
	FeeTypeReversal,
	FeeTypeOther,
}

// IsValid checks if the fee type is valid
func (f FeeType) IsValid() bool {
	for _, valid := range ValidFeeTypes {
		if f == valid {
			return true
		}
	}
	return false
}

// String returns the string representation of the fee type
func (f FeeType) String() string {
	return string(f)
}
