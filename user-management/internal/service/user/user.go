package service

import (
	"context"

	"github.com/muazwzxv/user-management/internal/dto/request"
	"github.com/muazwzxv/user-management/internal/dto/response"
)

type UserService interface {
	CreateUser(ctx context.Context, req request.CreateUserRequest) (*response.CreateUserResponse, error)
}
