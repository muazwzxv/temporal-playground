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

func TestGetBillHandler_Handle(t *testing.T) {
	t.Run("success - returns bill", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockBillRepo := mocks.NewMockBillRepository(ctrl)

		handler := &GetBillHandler{
			BillRepo: mockBillRepo,
		}

		billUUID := "bill-123"
		totalCents := int64(5000)
		closedAt := time.Date(2024, 1, 31, 23, 59, 59, 0, time.UTC)

		bill := &entity.BillEntity{
			UUID:         billUUID,
			CustomerUUID: "customer-123",
			Status:       "CLOSED",
			Currency:     "USD",
			TotalCents:   &totalCents,
			PeriodStart:  time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			PeriodEnd:    time.Date(2024, 1, 31, 23, 59, 59, 0, time.UTC),
			ClosedAt:     &closedAt,
			CreatedAt:    time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			UpdatedAt:    time.Date(2024, 1, 31, 23, 59, 59, 0, time.UTC),
		}

		mockBillRepo.EXPECT().
			FetchByUUID(gomock.Any(), billUUID).
			Return(bill, nil)

		resp, err := handler.Handle(context.Background(), &dto.GetBillRequest{UUID: billUUID})

		require.NoError(t, err)
		assert.Equal(t, billUUID, resp.UUID)
		assert.Equal(t, "customer-123", resp.CustomerUUID)
		assert.Equal(t, "CLOSED", resp.Status)
		assert.Equal(t, "USD", resp.Currency)
		assert.Equal(t, int64(5000), resp.TotalCents)
		assert.NotEmpty(t, resp.ClosedAt)
	})

	t.Run("success - returns open bill without total", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockBillRepo := mocks.NewMockBillRepository(ctrl)

		handler := &GetBillHandler{
			BillRepo: mockBillRepo,
		}

		billUUID := "bill-123"
		bill := &entity.BillEntity{
			UUID:         billUUID,
			CustomerUUID: "customer-123",
			Status:       "OPEN",
			Currency:     "USD",
			TotalCents:   nil, // No total yet
			PeriodStart:  time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			PeriodEnd:    time.Date(2024, 1, 31, 23, 59, 59, 0, time.UTC),
			ClosedAt:     nil,
			CreatedAt:    time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			UpdatedAt:    time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		}

		mockBillRepo.EXPECT().
			FetchByUUID(gomock.Any(), billUUID).
			Return(bill, nil)

		resp, err := handler.Handle(context.Background(), &dto.GetBillRequest{UUID: billUUID})

		require.NoError(t, err)
		assert.Equal(t, billUUID, resp.UUID)
		assert.Equal(t, "OPEN", resp.Status)
		assert.Equal(t, int64(0), resp.TotalCents)
		assert.Empty(t, resp.ClosedAt)
	})

	t.Run("error - missing UUID", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		handler := &GetBillHandler{
			BillRepo: mocks.NewMockBillRepository(ctrl),
		}

		resp, err := handler.Handle(context.Background(), &dto.GetBillRequest{UUID: ""})

		assert.Nil(t, resp)
		assert.Equal(t, utils.ErrUUIDMissing, err)
	})

	t.Run("error - bill not found", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockBillRepo := mocks.NewMockBillRepository(ctrl)

		handler := &GetBillHandler{
			BillRepo: mockBillRepo,
		}

		mockBillRepo.EXPECT().
			FetchByUUID(gomock.Any(), "nonexistent").
			Return(nil, sqldb.ErrNoRows)

		resp, err := handler.Handle(context.Background(), &dto.GetBillRequest{UUID: "nonexistent"})

		assert.Nil(t, resp)
		assert.Equal(t, utils.ErrNotFound, err)
	})

	t.Run("error - internal error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockBillRepo := mocks.NewMockBillRepository(ctrl)

		handler := &GetBillHandler{
			BillRepo: mockBillRepo,
		}

		mockBillRepo.EXPECT().
			FetchByUUID(gomock.Any(), "bill-123").
			Return(nil, assert.AnError)

		resp, err := handler.Handle(context.Background(), &dto.GetBillRequest{UUID: "bill-123"})

		assert.Nil(t, resp)
		assert.Equal(t, utils.ErrInternal, err)
	})
}
