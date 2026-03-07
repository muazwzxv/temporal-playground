package db

import "time"

// BillQueryParams contains filters and pagination options for fetching bills
type BillQueryParams struct {
	// Filters
	CustomerUUID string
	Status       string

	// Cursor (decoded values)
	CursorTime time.Time
	CursorID   int64

	// Pagination
	Limit    int
	SortDesc bool // true = newest first (default), false = oldest first
}

// LineItemQueryParams contains filters and pagination options for fetching line items
type LineItemQueryParams struct {
	// Filters
	BillUUID string

	// Cursor (decoded values)
	CursorTime time.Time
	CursorID   int64

	// Pagination
	Limit    int
	SortDesc bool // true = newest first (default), false = oldest first
}
