package service

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/gofiber/fiber/v2/log"
	"github.com/muazwzxv/user-management/internal/dto/cache"
	"github.com/muazwzxv/user-management/internal/dto/request"
	"github.com/muazwzxv/user-management/internal/dto/response"
	"github.com/muazwzxv/user-management/internal/entity"
	"github.com/muazwzxv/user-management/internal/redis"
	"github.com/muazwzxv/user-management/internal/repository"
	"github.com/muazwzxv/user-management/internal/worker"
	"github.com/muazwzxv/user-management/internal/worker/createuser"
	"github.com/samber/do/v2"
	"go.temporal.io/sdk/client"
)

type UserServiceImpl struct {
	repo        repository.UserRepository
	redisClient redis.Client
	temporalWf  *worker.Worker
}

func NewUserService(i do.Injector) (UserService, error) {
	repo := do.MustInvoke[repository.UserRepository](i)
	redisClient := do.MustInvoke[redis.Client](i)
	temporalWf := do.MustInvoke[*worker.Worker](i)

	return &UserServiceImpl{
		repo:        repo,
		redisClient: redisClient,
		temporalWf:  temporalWf,
	}, nil
}

func (s *UserServiceImpl) CreateUser(ctx context.Context, req request.CreateUserRequest) (*response.CreateUserResponse, error) {
	var (
		cacheResp *cache.CreateUserCacheResponse
	)
	getResp, err := s.redisClient.Get(ctx, req.ReferenceID)
	if err != nil {
		log.Errorw("error redis Exists, error: %+v", err)
		return nil, response.BuildError(http.StatusInternalServerError, "INTERNAL")
	}

	if getResp != "" {
		err := json.Unmarshal([]byte(getResp), &cacheResp)
		if err != nil {
			return nil, response.BuildError(http.StatusInternalServerError, "INTERNAL")
		}
		return &response.CreateUserResponse{
			ReferenceID: req.ReferenceID,
			User: response.UserResponse{
				Name:   cacheResp.Name,
				Status: entity.UserStatus(cacheResp.Status),
			},
		}, nil
	}

	cacheResp = &cache.CreateUserCacheResponse{
		Name:   req.Name,
		Status: string(entity.UserStatusProcessing),
	}

	if setErr := s.redisClient.Set(ctx, req.ReferenceID, cacheResp, time.Duration(7*time.Hour)); setErr != nil {
		log.Errorw("error redis Set, error: %+v", err)
		return nil, response.BuildError(http.StatusInternalServerError, "INTERNAL")
	}

	opts := client.StartWorkflowOptions{
		ID:        req.ReferenceID,
		TaskQueue: "", // TODO: fill this in
	}
	we, wfErr := s.temporalWf.Client().ExecuteWorkflow(ctx, opts, createuser.CreateUserWorkflow, createuser.CreateUserInput{
		ReferenceID: req.ReferenceID,
		Name:        req.Name,
	})
	if wfErr != nil {
		log.Errorw("error starting temporal workflow, error: %+v", err)
		return nil, response.BuildError(http.StatusInternalServerError, "INTERNAL")
	}
	log.Infow("Started workflow",
		"WorkflowID", we.GetID(),
		"RunID", we.GetRunID())

	return &response.CreateUserResponse{
		ReferenceID: req.ReferenceID,
	}, nil
}

func (s *UserServiceImpl) entityToResponse(item *entity.User) *response.UserResponse {
	return &response.UserResponse{
		UserUUID:  item.UserUUID,
		Name:      item.Name,
		Status:    item.Status,
		CreatedAt: item.CreatedAt,
		UpdatedAt: item.UpdatedAt,
	}
}
