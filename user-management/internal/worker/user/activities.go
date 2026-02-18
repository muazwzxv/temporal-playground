package user

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/muazwzxv/user-management/internal/entity"
	"github.com/muazwzxv/user-management/internal/redis"
	"github.com/muazwzxv/user-management/internal/repository"
	"github.com/samber/do/v2"
)

type Activities struct {
	repo        repository.UserRepository
	redisClient *redis.Client
}

func NewActivities(i do.Injector) *Activities {
	return &Activities{
		repo:        do.MustInvoke[repository.UserRepository](i),
		redisClient: do.MustInvoke[*redis.Client](i),
	}
}

type CreateUserInput struct {
	ReferenceID string
	Name        string
}

type CreateUserOutput struct {
	UserUUID string
	Name     string
	Status   string
}

func (a *Activities) CreateUserInDB(ctx context.Context, input CreateUserInput) (*CreateUserOutput, error) {
	user := &entity.User{
		UserUUID: uuid.New().String(),
		Name:     input.Name,
		Status:   entity.UserStatusActive,
	}

	if err := a.repo.Create(ctx, user); err != nil {
		// TODO: Inspect error type to determine if it's permanent or transient
		return nil, err
	}

	return &CreateUserOutput{
		UserUUID: user.UserUUID,
		Name:     user.Name,
		Status:   string(user.Status),
	}, nil
}

type CacheSuccessInput struct {
	ReferenceID string
	UserUUID    string
	Name        string
}

type CacheResponse struct {
	Status       string `json:"status"`
	UserUUID     string `json:"userUUID,omitempty"`
	Name         string `json:"name,omitempty"`
	ErrorCode    string `json:"errorCode,omitempty"`
	ErrorMessage string `json:"errorMessage,omitempty"`
}

func (a *Activities) UpdateCacheSuccess(ctx context.Context, input CacheSuccessInput) error {
	cacheData := CacheResponse{
		Status:   "completed",
		UserUUID: input.UserUUID,
		Name:     input.Name,
	}

	data, err := json.Marshal(cacheData)
	if err != nil {
		return err
	}

	return a.redisClient.Set(ctx, input.ReferenceID, string(data), 24*time.Hour)
}

type CacheFailureInput struct {
	ReferenceID  string
	ErrorCode    string
	ErrorMessage string
}

func (a *Activities) UpdateCacheFailure(ctx context.Context, input CacheFailureInput) error {
	cacheData := CacheResponse{
		Status:       "failed",
		ErrorCode:    input.ErrorCode,
		ErrorMessage: input.ErrorMessage,
	}

	data, err := json.Marshal(cacheData)
	if err != nil {
		return err
	}

	return a.redisClient.Set(ctx, input.ReferenceID, string(data), 1*time.Hour)
}
