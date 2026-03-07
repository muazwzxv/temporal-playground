package bill

import "time"

const (
	SignalAddLineItem = "add_line_item"
	SignalCloseBill   = "close_bill"
	QueryGetBillState = "get_bill_state"
)

type BillWorkflowInput struct {
	BillUUID  string
	PeriodEnd time.Time
}

type BillWorkflowResult struct {
	BillUUID   string
	TotalCents int64
	ItemCount  int
	ClosedAt   time.Time
}

type AddLineItemSignal struct {
	UUID           string
	IdempotencyKey string
	FeeType        string
	Description    string
	AmountCents    int64
	ReferenceUUID  *string
}

type BillStateQuery struct {
	Status     string
	TotalCents int64
	ItemCount  int
}

type InsertLineItemInput struct {
	UUID           string
	BillUUID       string
	IdempotencyKey string
	FeeType        string
	Description    string
	AmountCents    int64
	ReferenceUUID  *string
}

type InsertLineItemResult struct {
	UUID string
}

type CloseBillInput struct {
	BillUUID string
}

type CloseBillResult struct {
	TotalCents int64
	ClosedAt   time.Time
}
