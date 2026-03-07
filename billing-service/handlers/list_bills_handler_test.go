package handlers

import (
	"context"
	"testing"
	"time"

	"encore.app/db"
	"encore.app/db/repository/mocks"
	"encore.app/dto"
	"encore.app/entity"
	"encore.app/utils"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestListBillsHandler_Handle(t *testing.T) {
	t.Run("success - returns bills with pagination", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockBillRepo := mocks.NewMockBillRepository(ctrl)

		handler := &ListBillsHandler{
			BillRepo: mockBillRepo,
		}

		totalCents := int64(1000)
		bills := []*entity.BillEntity{
			{
				ID:           1,
				UUID:         "bill-1",
				CustomerUUID: "customer-123",
				Status:       "OPEN",
				Currency:     "USD",
				TotalCents:   &totalCents,
				PeriodStart:  time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				PeriodEnd:    time.Date(2024, 1, 31, 23, 59, 59, 0, time.UTC),
				CreatedAt:    time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			},
			{
				ID:           2,
				UUID:         "bill-2",
				CustomerUUID: "customer-123",
				Status:       "CLOSED",
				Currency:     "USD",
				TotalCents:   &totalCents,
				PeriodStart:  time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC),
				PeriodEnd:    time.Date(2024, 2, 29, 23, 59, 59, 0, time.UTC),
				CreatedAt:    time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC),
			},
		}

		mockBillRepo.EXPECT().
			FetchAll(gomock.Any(), gomock.Any()).
			Return(bills, nil)

		resp, err := handler.Handle(context.Background(), &dto.ListBillsRequest{
			CustomerUUID: "customer-123",
			Limit:        10,
		})

		require.NoError(t, err)
		assert.Len(t, resp.Data, 2)
		assert.Equal(t, "bill-1", resp.Data[0].UUID)
		assert.Equal(t, "bill-2", resp.Data[1].UUID)
		assert.False(t, resp.Pagination.HasMore)
	})

	t.Run("success - returns bills with hasMore", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockBillRepo := mocks.NewMockBillRepository(ctrl)

		handler := &ListBillsHandler{
			BillRepo: mockBillRepo,
		}

		totalCents := int64(1000)
		// Return 3 bills when limit is 2 (limit+1 to check hasMore)
		bills := []*entity.BillEntity{
			{
				ID:           1,
				UUID:         "bill-1",
				CustomerUUID: "customer-123",
				Status:       "OPEN",
				Currency:     "USD",
				TotalCents:   &totalCents,
				PeriodStart:  time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				PeriodEnd:    time.Date(2024, 1, 31, 23, 59, 59, 0, time.UTC),
				CreatedAt:    time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			},
			{
				ID:           2,
				UUID:         "bill-2",
				CustomerUUID: "customer-123",
				Status:       "CLOSED",
				Currency:     "USD",
				TotalCents:   &totalCents,
				PeriodStart:  time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC),
				PeriodEnd:    time.Date(2024, 2, 29, 23, 59, 59, 0, time.UTC),
				CreatedAt:    time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC),
			},
			{
				ID:           3,
				UUID:         "bill-3",
				CustomerUUID: "customer-123",
				Status:       "OPEN",
				Currency:     "USD",
				TotalCents:   &totalCents,
				PeriodStart:  time.Date(2024, 3, 1, 0, 0, 0, 0, time.UTC),
				PeriodEnd:    time.Date(2024, 3, 31, 23, 59, 59, 0, time.UTC),
				CreatedAt:    time.Date(2024, 3, 1, 0, 0, 0, 0, time.UTC),
			},
		}

		mockBillRepo.EXPECT().
			FetchAll(gomock.Any(), gomock.Any()).
			Return(bills, nil)

		resp, err := handler.Handle(context.Background(), &dto.ListBillsRequest{
			CustomerUUID: "customer-123",
			Limit:        2, // Only want 2
		})

		require.NoError(t, err)
		assert.Len(t, resp.Data, 2)
		assert.True(t, resp.Pagination.HasMore)
		assert.NotEmpty(t, resp.Pagination.NextCursor)
	})

	t.Run("success - returns empty list", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockBillRepo := mocks.NewMockBillRepository(ctrl)

		handler := &ListBillsHandler{
			BillRepo: mockBillRepo,
		}

		mockBillRepo.EXPECT().
			FetchAll(gomock.Any(), gomock.Any()).
			Return([]*entity.BillEntity{}, nil)

		resp, err := handler.Handle(context.Background(), &dto.ListBillsRequest{
			CustomerUUID: "customer-123",
		})

		require.NoError(t, err)
		assert.Len(t, resp.Data, 0)
		assert.False(t, resp.Pagination.HasMore)
	})

	t.Run("success - applies default limit", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockBillRepo := mocks.NewMockBillRepository(ctrl)

		handler := &ListBillsHandler{
			BillRepo: mockBillRepo,
		}

		mockBillRepo.EXPECT().
			FetchAll(gomock.Any(), gomock.AssignableToTypeOf(db.BillQueryParams{})).
			DoAndReturn(func(_ context.Context, params db.BillQueryParams) ([]*entity.BillEntity, error) {
				// Default limit should be 20, but we request 21 (limit+1)
				assert.Equal(t, 21, params.Limit)
				return []*entity.BillEntity{}, nil
			})

		resp, err := handler.Handle(context.Background(), &dto.ListBillsRequest{
			CustomerUUID: "customer-123",
			Limit:        0, // Should use default
		})

		require.NoError(t, err)
		assert.NotNil(t, resp)
	})

	t.Run("success - filters by status", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockBillRepo := mocks.NewMockBillRepository(ctrl)

		handler := &ListBillsHandler{
			BillRepo: mockBillRepo,
		}

		mockBillRepo.EXPECT().
			FetchAll(gomock.Any(), gomock.AssignableToTypeOf(db.BillQueryParams{})).
			DoAndReturn(func(_ context.Context, params db.BillQueryParams) ([]*entity.BillEntity, error) {
				assert.Equal(t, "OPEN", params.Status)
				return []*entity.BillEntity{}, nil
			})

		resp, err := handler.Handle(context.Background(), &dto.ListBillsRequest{
			CustomerUUID: "customer-123",
			Status:       "OPEN",
		})

		require.NoError(t, err)
		assert.NotNil(t, resp)
	})

	t.Run("error - invalid cursor", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		handler := &ListBillsHandler{
			BillRepo: mocks.NewMockBillRepository(ctrl),
		}

		resp, err := handler.Handle(context.Background(), &dto.ListBillsRequest{
			CustomerUUID: "customer-123",
			Cursor:       "invalid-cursor-format",
		})

		assert.Nil(t, resp)
		assert.Equal(t, utils.ErrInvalidCursor, err)
	})

	t.Run("error - internal error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockBillRepo := mocks.NewMockBillRepository(ctrl)

		handler := &ListBillsHandler{
			BillRepo: mockBillRepo,
		}

		mockBillRepo.EXPECT().
			FetchAll(gomock.Any(), gomock.Any()).
			Return(nil, assert.AnError)

		resp, err := handler.Handle(context.Background(), &dto.ListBillsRequest{
			CustomerUUID: "customer-123",
		})

		assert.Nil(t, resp)
		assert.Equal(t, utils.ErrInternal, err)
	})

	t.Run("success - handles nil total cents", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockBillRepo := mocks.NewMockBillRepository(ctrl)

		handler := &ListBillsHandler{
			BillRepo: mockBillRepo,
		}

		bills := []*entity.BillEntity{
			{
				ID:           1,
				UUID:         "bill-1",
				CustomerUUID: "customer-123",
				Status:       "OPEN",
				Currency:     "USD",
				TotalCents:   nil, // No total yet
				PeriodStart:  time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				PeriodEnd:    time.Date(2024, 1, 31, 23, 59, 59, 0, time.UTC),
				CreatedAt:    time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			},
		}

		mockBillRepo.EXPECT().
			FetchAll(gomock.Any(), gomock.Any()).
			Return(bills, nil)

		resp, err := handler.Handle(context.Background(), &dto.ListBillsRequest{
			CustomerUUID: "customer-123",
		})

		require.NoError(t, err)
		assert.Len(t, resp.Data, 1)
		assert.Equal(t, int64(0), resp.Data[0].Total.Amount)
	})
}
