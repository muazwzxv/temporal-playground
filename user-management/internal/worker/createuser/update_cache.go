package createuser

import (
	"context"
	"encoding/json"
	"time"

	"github.com/muazwzxv/user-management/internal/dto/cache"
)

func (a *Activities) UpdateCacheSuccess(ctx context.Context, input CacheSuccessInput) error {
	cacheData := cache.CreateUserCacheResponse{
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

func (a *Activities) UpdateCacheFailure(ctx context.Context, input CacheFailureInput) error {
	cacheData := cache.CreateUserCacheResponse{
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
