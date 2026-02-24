package createuser

import (
	"context"

	"github.com/muazwzxv/user-management/internal/entity"
	"go.temporal.io/sdk/temporal"
)

type insertUserInput struct {
	uuid   string
	name   string
	status entity.UserStatus
}

func (a *Activities) CreateUserInDB(ctx context.Context, in insertUserInput) (*CreateUserOutput, error) {
	user := &entity.User{
		UserUUID: in.uuid,
		Name:     in.name,
		Status:   in.status,
	}

	if err := a.repo.Create(ctx, user); err != nil {
		return nil, temporal.NewApplicationErrorWithCause("error inserting user", "DB_ERROR", err)
	}

	return &CreateUserOutput{
		UserUUID: user.UserUUID,
		Name:     user.Name,
		Status:   string(user.Status),
	}, nil
}
