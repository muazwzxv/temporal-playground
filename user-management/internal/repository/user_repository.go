package repository

import (
	"context"
	"database/sql"

	"github.com/samber/do/v2"
	"github.com/muazwzxv/user-management/internal/database"
	"github.com/muazwzxv/user-management/internal/database/store"
	"github.com/muazwzxv/user-management/internal/entity"
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
	result, err := r.queries.CreateUser(ctx, r.db, store.CreateUserParams{
		Name: item.Name,
		Description: sql.NullString{
			String: item.Description,
			Valid:  item.Description != "",
		},
		Status: string(item.Status),
	})
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}

	item.ID = id
	return nil
}

func (r *UserRepositoryImpl) GetByID(ctx context.Context, id int64) (*entity.User, error) {
	row, err := r.queries.GetUserByID(ctx, r.db, id)
	if err != nil {
		return nil, err
	}

	return r.toEntity(row), nil
}

func (r *UserRepositoryImpl) toEntity(row *store.User) *entity.User {
	result := &entity.User{
		ID:     row.ID,
		Name:   row.Name,
		Status: entity.UserStatus(row.Status),
	}
	
	if row.Description.Valid {
		result.Description = row.Description.String
	}
	if row.CreatedAt.Valid {
		result.CreatedAt = row.CreatedAt.Time
	}
	if row.UpdatedAt.Valid {
		result.UpdatedAt = row.UpdatedAt.Time
	}
	
	return result
}
