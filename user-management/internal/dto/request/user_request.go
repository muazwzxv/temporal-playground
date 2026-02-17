// Package request contains structures for API request payloads.
package request

type CreateUserRequest struct {
	Name        string `json:"name" validate:"required,min=1,max=255"`
	Description string `json:"description" validate:"max=1000"`
	Status      string `json:"status" validate:"omitempty,oneof=active inactive archived"`
}

type UpdateUserRequest struct {
	Name        *string `json:"name,omitempty" validate:"omitempty,min=1,max=255"`
	Description *string `json:"description,omitempty" validate:"omitempty,max=1000"`
	Status      *string `json:"status,omitempty" validate:"omitempty,oneof=active inactive archived"`
}

type BulkUserRequest struct {
	IDs    []int64 `json:"ids" validate:"required,min=1,max=100"`
	Status string  `json:"status" validate:"required,oneof=active inactive archived"`
}
