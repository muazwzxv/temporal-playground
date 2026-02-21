package createuser

import (
	"github.com/muazwzxv/user-management/internal/redis"
	"github.com/muazwzxv/user-management/internal/repository"
	"github.com/samber/do/v2"
)

type Activities struct {
	repo        repository.UserRepository
	redisClient *redis.Client
}

func NewUserActivities(i do.Injector) *Activities {
	return &Activities{
		repo:        do.MustInvoke[repository.UserRepository](i),
		redisClient: do.MustInvoke[*redis.Client](i),
	}
}
