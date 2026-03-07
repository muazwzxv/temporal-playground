package dto

import "time"

type CreateCustomerRequest struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

type CreateCustomerResponse struct {
	UUID  string `json:"uuid"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

type GetCustomerRequest struct {
	UUID string `json:"uuid"`
}

type GetCustomerResponse struct {
	UUID      string    `json:"uuid"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"createdAt"`
}
