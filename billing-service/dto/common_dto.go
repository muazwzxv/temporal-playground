package dto

type Money struct {
	Amount   int64  `json:"amount"`
	Currency string `json:"currency"`
}

type PaginationResponse struct {
	NextCursor string `json:"nextCursor,omitempty"`
	HasMore    bool   `json:"hasMore"`
}

// SortOrder for list endpoints
type SortOrder string

const (
	SortOrderAsc  SortOrder = "asc"
	SortOrderDesc SortOrder = "desc"
)
