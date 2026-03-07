package entity

import "time"

type LineItemEntity struct {
	ID             int64 `json:"-"` // Internal use only, excluded from JSON
	UUID           string
	BillUUID       string
	IdempotencyKey string
	FeeType        string
	Description    string
	AmountCents    int64
	ReferenceUUID  *string
	CreatedAt      time.Time
}
