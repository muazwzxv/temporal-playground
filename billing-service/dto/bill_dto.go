package dto

type CreateBillRequest struct {
	UUID         string `json:"uuid"`
	CustomerUUID string `json:"customerUuid"`
	Currency     string `json:"currency"`
	PeriodStart  string `json:"periodStart"`
	PeriodEnd    string `json:"periodEnd"`
}

type CreateBillResponse struct {
	UUID        string `json:"uuid"`
	Status      string `json:"status"`
	Currency    string `json:"currency"`
	PeriodStart string `json:"periodStart"`
	PeriodEnd   string `json:"periodEnd"`
}

type GetBillRequest struct {
	UUID string `json:"uuid"`
}

type GetBillResponse struct {
	UUID         string `json:"uuid"`
	CustomerUUID string `json:"customerUuid"`
	Status       string `json:"status"`
	Currency     string `json:"currency"`
	TotalCents   int64  `json:"totalCents"`
	PeriodStart  string `json:"periodStart"`
	PeriodEnd    string `json:"periodEnd"`
	ClosedAt     string `json:"closedAt,omitempty"`
	CreatedAt    string `json:"createdAt"`
	UpdatedAt    string `json:"updatedAt"`
}

// CloseBillRequest for POST /v1/bill/close
type CloseBillRequest struct {
	UUID string `json:"uuid"`
}

// CloseBillResponse - async response, client should poll GetBill for final state
type CloseBillResponse struct {
	UUID    string `json:"uuid"`
	Status  string `json:"status"` // "CLOSING"
	Message string `json:"message,omitempty"`
}

// ListBillsRequest for POST /v1/bill/list
type ListBillsRequest struct {
	CustomerUUID string    `json:"customerUuid,omitempty"`
	Status       string    `json:"status,omitempty"` // "OPEN" or "CLOSED"
	Cursor       string    `json:"cursor,omitempty"`
	Limit        int       `json:"limit,omitempty"`     // default 20, max 20
	SortOrder    SortOrder `json:"sortOrder,omitempty"` // "asc" or "desc", default "desc"
}

// BillSummary for list responses
type BillSummary struct {
	UUID         string `json:"uuid"`
	CustomerUUID string `json:"customerUuid"`
	Status       string `json:"status"`
	Currency     string `json:"currency"`
	Total        Money  `json:"total"`
	PeriodStart  string `json:"periodStart"`
	PeriodEnd    string `json:"periodEnd"`
}

// ListBillsResponse for POST /v1/bill/list
type ListBillsResponse struct {
	Data       []BillSummary      `json:"data"`
	Pagination PaginationResponse `json:"pagination"`
}
