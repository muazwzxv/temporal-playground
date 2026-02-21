// Package repository contains interfaces for data access operations.
package repository

import (
	"context"

	"github.com/muazwzxv/user-management/internal/entity"
	"github.com/samber/do/v2"
)

type UserRepository interface {
	Create(ctx context.Context, item *entity.User) error
	//GetByID(ctx context.Context, id int64) (*entity.User, error)
}

type DatabaseRepository interface {
	Ping(ctx context.Context) error
	Close() error
}

func InjectRepository(i do.Injector) {
	// Provide repositories
	do.Provide(i, NewUserRepository)
}
