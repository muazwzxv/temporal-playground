package handlers

import (
	"context"
	"testing"
	"time"

	"encore.app/db/repository/mocks"
	"encore.app/dto"
	"encore.app/entity"
	"encore.app/utils"

	"encore.dev/storage/sqldb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestGetCustomerHandler_Handle(t *testing.T) {
	t.Run("success - returns customer", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockCustomerRepo := mocks.NewMockCustomerRepository(ctrl)

		handler := &GetCustomerHandler{
			CustomerRepo: mockCustomerRepo,
		}

		customerUUID := "customer-123"
		customer := &entity.CustomerEntity{
			UUID:      customerUUID,
			Name:      "Test User",
			Email:     "test@example.com",
			CreatedAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		}

		mockCustomerRepo.EXPECT().
			FetchByUUID(gomock.Any(), customerUUID).
			Return(customer, nil)

		resp, err := handler.Handle(context.Background(), &dto.GetCustomerRequest{UUID: customerUUID})

		require.NoError(t, err)
		assert.Equal(t, customerUUID, resp.UUID)
		assert.Equal(t, "Test User", resp.Name)
		assert.Equal(t, "test@example.com", resp.Email)
		assert.Equal(t, time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), resp.CreatedAt)
	})

	t.Run("error - missing UUID", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		handler := &GetCustomerHandler{
			CustomerRepo: mocks.NewMockCustomerRepository(ctrl),
		}

		resp, err := handler.Handle(context.Background(), &dto.GetCustomerRequest{UUID: ""})

		assert.Nil(t, resp)
		assert.Equal(t, utils.ErrUUIDMissing, err)
	})

	t.Run("error - customer not found", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockCustomerRepo := mocks.NewMockCustomerRepository(ctrl)

		handler := &GetCustomerHandler{
			CustomerRepo: mockCustomerRepo,
		}

		mockCustomerRepo.EXPECT().
			FetchByUUID(gomock.Any(), "nonexistent").
			Return(nil, sqldb.ErrNoRows)

		resp, err := handler.Handle(context.Background(), &dto.GetCustomerRequest{UUID: "nonexistent"})

		assert.Nil(t, resp)
		assert.Equal(t, utils.ErrNotFound, err)
	})

	t.Run("error - internal error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockCustomerRepo := mocks.NewMockCustomerRepository(ctrl)

		handler := &GetCustomerHandler{
			CustomerRepo: mockCustomerRepo,
		}

		mockCustomerRepo.EXPECT().
			FetchByUUID(gomock.Any(), "customer-123").
			Return(nil, assert.AnError)

		resp, err := handler.Handle(context.Background(), &dto.GetCustomerRequest{UUID: "customer-123"})

		assert.Nil(t, resp)
		assert.Equal(t, utils.ErrInternal, err)
	})
}
