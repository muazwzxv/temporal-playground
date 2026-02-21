package createuser

import (
	"context"

	"github.com/google/uuid"
	"github.com/muazwzxv/user-management/internal/entity"
)

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
