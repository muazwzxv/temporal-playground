package service

import (
	"context"

	"github.com/gofiber/fiber/v2"
	"github.com/samber/do/v2"
	"github.com/muazwzxv/user-management/internal/dto/request"
	"github.com/muazwzxv/user-management/internal/dto/response"
	"github.com/muazwzxv/user-management/internal/entity"
	"github.com/muazwzxv/user-management/internal/repository"
)

type UserServiceImpl struct {
	repo repository.UserRepository
}

func NewUserService(i do.Injector) (UserService, error) {
	repo := do.MustInvoke[repository.UserRepository](i)

	return &UserServiceImpl{
		repo: repo,
	}, nil
}

func (s *UserServiceImpl) CreateUser(ctx context.Context, req request.CreateUserRequest) (*response.UserResponse, error) {
	status := entity.UserStatusActive
	if req.Status != "" {
		status = entity.UserStatus(req.Status)
	}

	item := &entity.User{
		Name:        req.Name,
		Description: req.Description,
		Status:      status,
	}

	if err := s.repo.Create(ctx, item); err != nil {
		return nil, response.BuildErrorWithCode(
			fiber.StatusInternalServerError,
			"Failed to create user",
			"CREATE_ERROR",
		)
	}

	return s.entityToResponse(item), nil
}

func (s *UserServiceImpl) entityToResponse(item *entity.User) *response.UserResponse {
	return &response.UserResponse{
		ID:          item.ID,
		Name:        item.Name,
		Description: item.Description,
		Status:      item.Status,
		CreatedAt:   item.CreatedAt,
		UpdatedAt:   item.UpdatedAt,
		IsActive:    item.IsActive(),
	}
}
