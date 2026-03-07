package dto

type AddLineItemRequest struct {
	BillUUID       string `json:"billUuid"`
	IdempotencyKey string `json:"idempotencyKey"`
	FeeType        string `json:"feeType"`
	Description    string `json:"description"`
	Amount         Money  `json:"amount"`
}

type AddLineItemResponse struct {
	UUID      string `json:"uuid"`
	FeeType   string `json:"feeType"`
	Amount    Money  `json:"amount"`
	Status    string `json:"status"` // "pending" or "persisted"
	CreatedAt string `json:"createdAt"`
}

// ListLineItemsRequest for POST /v1/bill/list-line-items
type ListLineItemsRequest struct {
	BillUUID string `json:"billUuid"`
	Cursor   string `json:"cursor,omitempty"`
	Limit    int    `json:"limit,omitempty"` // default 50
}

// LineItemSummary for list responses
type LineItemSummary struct {
	UUID          string `json:"uuid"`
	FeeType       string `json:"feeType"`
	Description   string `json:"description,omitempty"`
	Amount        Money  `json:"amount"`
	ReferenceUUID string `json:"referenceUuid,omitempty"`
	CreatedAt     string `json:"createdAt"`
}

// ListLineItemsResponse for POST /v1/bill/list-line-items
type ListLineItemsResponse struct {
	Data       []LineItemSummary  `json:"data"`
	Pagination PaginationResponse `json:"pagination"`
}

// ReverseLineItemRequest for POST /v1/bill/reverse-line-item
type ReverseLineItemRequest struct {
	BillUUID       string `json:"billUuid"`
	LineItemUUID   string `json:"lineItemUuid"`
	IdempotencyKey string `json:"idempotencyKey"`
	Reason         string `json:"reason,omitempty"`
}

// ReverseLineItemResponse for POST /v1/bill/reverse-line-item
type ReverseLineItemResponse struct {
	UUID          string `json:"uuid"`
	FeeType       string `json:"feeType"` // "REVERSAL"
	ReferenceUUID string `json:"referenceUuid"`
	Amount        Money  `json:"amount"` // negative
	CreatedAt     string `json:"createdAt"`
}
