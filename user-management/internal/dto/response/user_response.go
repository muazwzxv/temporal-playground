// Package response contains structures for API response payloads.
package response

import (
	"time"

	"github.com/muazwzxv/user-management/internal/entity"
)

type UserResponse struct {
	ID          int64                        `json:"id"`
	Name        string                       `json:"name"`
	Description string                       `json:"description"`
	Status      entity.UserStatus `json:"status"`
	CreatedAt   time.Time                    `json:"created_at"`
	UpdatedAt   time.Time                    `json:"updated_at"`
	IsActive    bool                         `json:"is_active"`
}

type UserDetailResponse struct {
	UserResponse
}

type CreateUserResponse struct {
	ID      int64  `json:"id"`
	Message string `json:"message"`
}

type UpdateUserResponse struct {
	ID      int64  `json:"id"`
	Message string `json:"message"`
	Updated bool   `json:"updated"`
}

type UserListResponse struct {
	Users []UserResponse `json:"users"`
	Pagination         PaginationMetadata           `json:"pagination"`
}

type BulkUserResponse struct {
	UpdatedCount int     `json:"updated_count"`
	FailedIDs    []int64 `json:"failed_ids,omitempty"`
	Message      string  `json:"message"`
}

type UserStatsResponse struct {
	TotalCount    int64 `json:"total_count"`
	ActiveCount   int64 `json:"active_count"`
	InactiveCount int64 `json:"inactive_count"`
	ArchivedCount int64 `json:"archived_count"`
}
