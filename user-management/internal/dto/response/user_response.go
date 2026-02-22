// Package response contains structures for API response payloads.
package response

import (
	"github.com/muazwzxv/user-management/internal/entity"
)

type UserResponse struct {
	UserUUID string            `json:"user_uuid"`
	Name     string            `json:"name"`
	Status   entity.UserStatus `json:"status"`
	IsActive bool              `json:"is_active,omitempty"`
}

type UserDetailResponse struct {
	UserResponse
}

type CreateUserResponse struct {
	ReferenceID string        `json:"reference_id"`
	User        *UserResponse `json:"user,omitempty"`
}

type UpdateUserResponse struct {
	ID      int64  `json:"id"`
	Message string `json:"message"`
	Updated bool   `json:"updated"`
}

type UserListResponse struct {
	Users      []UserResponse     `json:"users"`
	Pagination PaginationMetadata `json:"pagination"`
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
