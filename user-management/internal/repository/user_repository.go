package repository

import (
	"context"

	"github.com/muazwzxv/user-management/internal/database"
	"github.com/muazwzxv/user-management/internal/database/store"
	"github.com/muazwzxv/user-management/internal/entity"
	"github.com/samber/do/v2"
)

type UserRepositoryImpl struct {
	queries *store.Queries
	db      store.DBTX
}

func NewUserRepository(i do.Injector) (UserRepository, error) {
	queries := do.MustInvoke[*store.Queries](i)
	db := do.MustInvoke[*database.Database](i)

	return &UserRepositoryImpl{
		queries: queries,
		db:      db.DB, // Extract *sqlx.DB from Database wrapper
	}, nil
}

func (r *UserRepositoryImpl) Create(ctx context.Context, item *entity.User) error {
	_, err := r.queries.CreateUser(ctx, r.db, store.CreateUserParams{
		UserUuid: item.UserUUID,
		Name:     item.Name,
		Status:   string(item.Status),
	})
	if err != nil {
		return err
	}
	return nil
}

// func (r *UserRepositoryImpl) GetByID(ctx context.Context, id int64) (*entity.User, error) {
// 	row, err := r.queries.GetUserByID(ctx, r.db, id)
// 	if err != nil {
// 		return nil, err
// 	}

// 	return r.toEntity(row), nil
// }

// func (r *UserRepositoryImpl) toEntity(row *store.User) *entity.User {
// 	result := &entity.User{
// 		ID:     row.ID,
// 		Name:   row.Name,
// 		Status: entity.UserStatus(row.Status),
// 	}

// 	if row.Description.Valid {
// 		result.Description = row.Description.String
// 	}
// 	if row.CreatedAt.Valid {
// 		result.CreatedAt = row.CreatedAt.Time
// 	}
// 	if row.UpdatedAt.Valid {
// 		result.UpdatedAt = row.UpdatedAt.Time
// 	}

// 	return result
// }
