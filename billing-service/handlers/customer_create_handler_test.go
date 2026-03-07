package handlers

import (
	"context"
	"testing"

	"encore.app/db/repository/mocks"
	"encore.app/dto"
	"encore.app/entity"
	"encore.app/utils"

	"encore.dev/storage/sqldb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestCreateCustomerHandler_Handle(t *testing.T) {
	t.Run("success - creates customer", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockCustomerRepo := mocks.NewMockCustomerRepository(ctrl)

		handler := &CreateCustomerHandler{
			CustomerRepo: mockCustomerRepo,
		}

		mockCustomerRepo.EXPECT().
			FetchByEmail(gomock.Any(), "test@example.com").
			Return(nil, sqldb.ErrNoRows)

		mockCustomerRepo.EXPECT().
			Insert(gomock.Any(), gomock.Any()).
			Return(nil)

		resp, err := handler.Handle(context.Background(), &dto.CreateCustomerRequest{
			Name:  "Test User",
			Email: "test@example.com",
		})

		require.NoError(t, err)
		assert.NotEmpty(t, resp.UUID)
		assert.Equal(t, "Test User", resp.Name)
		assert.Equal(t, "test@example.com", resp.Email)
	})

	t.Run("error - validation fails - missing name", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		handler := &CreateCustomerHandler{
			CustomerRepo: mocks.NewMockCustomerRepository(ctrl),
		}

		resp, err := handler.Handle(context.Background(), &dto.CreateCustomerRequest{
			Name:  "",
			Email: "test@example.com",
		})

		assert.Nil(t, resp)
		assert.NotNil(t, err)
	})

	t.Run("error - validation fails - missing email", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		handler := &CreateCustomerHandler{
			CustomerRepo: mocks.NewMockCustomerRepository(ctrl),
		}

		resp, err := handler.Handle(context.Background(), &dto.CreateCustomerRequest{
			Name:  "Test User",
			Email: "",
		})

		assert.Nil(t, resp)
		assert.NotNil(t, err)
	})

	t.Run("error - email already used", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockCustomerRepo := mocks.NewMockCustomerRepository(ctrl)

		handler := &CreateCustomerHandler{
			CustomerRepo: mockCustomerRepo,
		}

		existingCustomer := &entity.CustomerEntity{
			UUID:  "existing-uuid",
			Name:  "Existing User",
			Email: "test@example.com",
		}

		mockCustomerRepo.EXPECT().
			FetchByEmail(gomock.Any(), "test@example.com").
			Return(existingCustomer, nil)

		resp, err := handler.Handle(context.Background(), &dto.CreateCustomerRequest{
			Name:  "Test User",
			Email: "test@example.com",
		})

		assert.Nil(t, resp)
		assert.Equal(t, utils.ErrEmailAlreadyUsed, err)
	})

	t.Run("error - internal error checking email", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockCustomerRepo := mocks.NewMockCustomerRepository(ctrl)

		handler := &CreateCustomerHandler{
			CustomerRepo: mockCustomerRepo,
		}

		mockCustomerRepo.EXPECT().
			FetchByEmail(gomock.Any(), "test@example.com").
			Return(nil, assert.AnError)

		resp, err := handler.Handle(context.Background(), &dto.CreateCustomerRequest{
			Name:  "Test User",
			Email: "test@example.com",
		})

		assert.Nil(t, resp)
		assert.Equal(t, utils.ErrInternal, err)
	})

	t.Run("error - insert failed", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockCustomerRepo := mocks.NewMockCustomerRepository(ctrl)

		handler := &CreateCustomerHandler{
			CustomerRepo: mockCustomerRepo,
		}

		mockCustomerRepo.EXPECT().
			FetchByEmail(gomock.Any(), "test@example.com").
			Return(nil, sqldb.ErrNoRows)

		mockCustomerRepo.EXPECT().
			Insert(gomock.Any(), gomock.Any()).
			Return(assert.AnError)

		resp, err := handler.Handle(context.Background(), &dto.CreateCustomerRequest{
			Name:  "Test User",
			Email: "test@example.com",
		})

		assert.Nil(t, resp)
		assert.Equal(t, assert.AnError, err)
	})
}
