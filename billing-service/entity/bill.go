package entity

import "time"

type BillEntity struct {
	ID           int64 `json:"-"` // Internal use only, excluded from JSON
	UUID         string
	CustomerUUID string
	Currency     string
	Status       string
	PeriodStart  time.Time
	PeriodEnd    time.Time
	ClosedAt     *time.Time
	TotalCents   *int64

	CreatedAt time.Time
	UpdatedAt time.Time
}

func (b *BillEntity) IsOpen() bool {
	return b.Status == "OPEN"
}
