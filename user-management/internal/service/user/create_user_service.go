package service

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/gofiber/fiber/v2/log"
	"github.com/google/uuid"
	"github.com/muazwzxv/user-management/internal/dto/request"
	"github.com/muazwzxv/user-management/internal/dto/response"
	"github.com/muazwzxv/user-management/internal/entity"
	"github.com/muazwzxv/user-management/internal/redis"
	"github.com/muazwzxv/user-management/internal/repository"
	"github.com/samber/do/v2"
	"go.temporal.io/sdk/client"
)

type UserServiceImpl struct {
	repo           repository.UserRepository
	redisClient    redis.Client
	temporalClient client.Client
}

func NewUserService(i do.Injector) (UserService, error) {
	repo := do.MustInvoke[repository.UserRepository](i)
	redisClient := do.MustInvoke[redis.Client](i)
	temporalClient := do.MustInvoke[client.Client](i)

	return &UserServiceImpl{
		repo:           repo,
		redisClient:    redisClient,
		temporalClient: temporalClient,
	}, nil
}

func (s *UserServiceImpl) CreateUser(ctx context.Context, req request.CreateUserRequest) (*response.UserResponse, error) {
	// TODO: logic
	// 1 - check idempotency exists or not
	// 		- if exist status processing, return payload
	// 		- if exist status completed, return cached payload
	// 		- not exist, set key with payload and status
	var (
		resp *response.UserResponse
	)
	getResp, err := s.redisClient.Get(ctx, req.ReferenceID)
	if err != nil {
		log.Errorw("error redis Exists, error: %+v", err)
		return nil, response.BuildError(http.StatusInternalServerError, "INTERNAL")
	}

	if getResp != "" {
		err := json.Unmarshal([]byte(getResp), &resp)
		if err != nil {
			return nil, response.BuildError(http.StatusInternalServerError, "INTERNAL")
		}
		return resp, nil
	}

	resp = &response.UserResponse{
		Name:   req.Name,
		Status: entity.UserStatusProcessing,
	}

	if setErr := s.redisClient.Set(ctx, req.ReferenceID, resp, time.Duration(7*time.Hour)); setErr != nil {
		log.Errorw("error redis Set, error: %+v", err)
		return nil, response.BuildError(http.StatusInternalServerError, "INTERNAL")
	}

	// TODO: proceed to handle inserts
	user := &entity.User{
		UserUUID: uuid.New().String(),
		Name:     req.Name,
		Status:   entity.UserStatusProcessing,
	}
	createErr := s.repo.Create(ctx, user)
	if createErr != nil {
		return nil, response.BuildError(http.StatusInternalServerError, "INTERNAL")
	}

	// TODO: update cache on successful inserts, TTL 3 mins

	return nil, nil
}

func (s *UserServiceImpl) entityToResponse(item *entity.User) *response.UserResponse {
	return &response.UserResponse{
		ID:        item.ID,
		UserUUID:  item.UserUUID,
		Name:      item.Name,
		Status:    item.Status,
		CreatedAt: item.CreatedAt,
		UpdatedAt: item.UpdatedAt,
		IsActive:  item.IsActive(),
	}
}
