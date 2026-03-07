package entity

import "time"

type CustomerEntity struct {
	UUID      string
	Name      string
	Email     string
	CreatedAt time.Time
	UpdatedAt time.Time
}
