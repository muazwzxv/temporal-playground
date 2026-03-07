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

func TestListLineItemsHandler_Handle(t *testing.T) {
	t.Run("success - returns line items with pagination", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockBillRepo := mocks.NewMockBillRepository(ctrl)
		mockLineItemRepo := mocks.NewMockLineItemRepository(ctrl)

		handler := &ListLineItemsHandler{
			BillRepo:     mockBillRepo,
			LineItemRepo: mockLineItemRepo,
		}

		billUUID := "bill-123"
		bill := &entity.BillEntity{
			UUID:     billUUID,
			Currency: "USD",
		}

		lineItems := []*entity.LineItemEntity{
			{
				ID:          1,
				UUID:        "item-1",
				BillUUID:    billUUID,
				FeeType:     "TRANSACTION",
				Description: "Test transaction 1",
				AmountCents: 1000,
				CreatedAt:   time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			},
			{
				ID:          2,
				UUID:        "item-2",
				BillUUID:    billUUID,
				FeeType:     "TRANSACTION",
				Description: "Test transaction 2",
				AmountCents: 2000,
				CreatedAt:   time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
			},
		}

		mockBillRepo.EXPECT().
			FetchByUUID(gomock.Any(), billUUID).
			Return(bill, nil)

		// First page: zero time and zero ID (no cursor)
		mockLineItemRepo.EXPECT().
			FetchByBillUUID(gomock.Any(), billUUID, time.Time{}, int64(0), 21). // limit+1
			Return(lineItems, nil)

		resp, err := handler.Handle(context.Background(), &dto.ListLineItemsRequest{
			BillUUID: billUUID,
			Limit:    20,
		})

		require.NoError(t, err)
		assert.Len(t, resp.Data, 2)
		assert.Equal(t, "item-1", resp.Data[0].UUID)
		assert.Equal(t, "item-2", resp.Data[1].UUID)
		assert.Equal(t, "USD", resp.Data[0].Amount.Currency)
		assert.False(t, resp.Pagination.HasMore)
	})

	t.Run("success - returns line items with hasMore and encoded cursor", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockBillRepo := mocks.NewMockBillRepository(ctrl)
		mockLineItemRepo := mocks.NewMockLineItemRepository(ctrl)

		handler := &ListLineItemsHandler{
			BillRepo:     mockBillRepo,
			LineItemRepo: mockLineItemRepo,
		}

		billUUID := "bill-123"
		bill := &entity.BillEntity{
			UUID:     billUUID,
			Currency: "USD",
		}

		item2CreatedAt := time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC)

		// Return 3 items when limit is 2
		lineItems := []*entity.LineItemEntity{
			{
				ID:          1,
				UUID:        "item-1",
				BillUUID:    billUUID,
				FeeType:     "TRANSACTION",
				AmountCents: 1000,
				CreatedAt:   time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			},
			{
				ID:          2,
				UUID:        "item-2",
				BillUUID:    billUUID,
				FeeType:     "TRANSACTION",
				AmountCents: 2000,
				CreatedAt:   item2CreatedAt,
			},
			{
				ID:          3,
				UUID:        "item-3",
				BillUUID:    billUUID,
				FeeType:     "TRANSACTION",
				AmountCents: 3000,
				CreatedAt:   time.Date(2024, 1, 3, 0, 0, 0, 0, time.UTC),
			},
		}

		mockBillRepo.EXPECT().
			FetchByUUID(gomock.Any(), billUUID).
			Return(bill, nil)

		// First page: zero time and zero ID
		mockLineItemRepo.EXPECT().
			FetchByBillUUID(gomock.Any(), billUUID, time.Time{}, int64(0), 3). // limit+1 = 3
			Return(lineItems, nil)

		resp, err := handler.Handle(context.Background(), &dto.ListLineItemsRequest{
			BillUUID: billUUID,
			Limit:    2,
		})

		require.NoError(t, err)
		assert.Len(t, resp.Data, 2)
		assert.True(t, resp.Pagination.HasMore)
		// NextCursor should be encoded from item-2's createdAt and ID
		expectedCursor := utils.EncodeCursor(item2CreatedAt, 2)
		assert.Equal(t, expectedCursor, resp.Pagination.NextCursor)
	})

	t.Run("success - returns line item with reference UUID (reversal)", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockBillRepo := mocks.NewMockBillRepository(ctrl)
		mockLineItemRepo := mocks.NewMockLineItemRepository(ctrl)

		handler := &ListLineItemsHandler{
			BillRepo:     mockBillRepo,
			LineItemRepo: mockLineItemRepo,
		}

		billUUID := "bill-123"
		originalUUID := "item-1"
		bill := &entity.BillEntity{
			UUID:     billUUID,
			Currency: "USD",
		}

		lineItems := []*entity.LineItemEntity{
			{
				ID:            1,
				UUID:          "item-reversal",
				BillUUID:      billUUID,
				FeeType:       "REVERSAL",
				AmountCents:   -1000,
				ReferenceUUID: &originalUUID,
				CreatedAt:     time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
			},
		}

		mockBillRepo.EXPECT().
			FetchByUUID(gomock.Any(), billUUID).
			Return(bill, nil)

		mockLineItemRepo.EXPECT().
			FetchByBillUUID(gomock.Any(), billUUID, time.Time{}, int64(0), 21).
			Return(lineItems, nil)

		resp, err := handler.Handle(context.Background(), &dto.ListLineItemsRequest{
			BillUUID: billUUID,
		})

		require.NoError(t, err)
		assert.Len(t, resp.Data, 1)
		assert.Equal(t, "REVERSAL", resp.Data[0].FeeType)
		assert.Equal(t, int64(-1000), resp.Data[0].Amount.Amount)
		assert.Equal(t, originalUUID, resp.Data[0].ReferenceUUID)
	})

	t.Run("error - missing bill UUID", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		handler := &ListLineItemsHandler{
			BillRepo:     mocks.NewMockBillRepository(ctrl),
			LineItemRepo: mocks.NewMockLineItemRepository(ctrl),
		}

		resp, err := handler.Handle(context.Background(), &dto.ListLineItemsRequest{
			BillUUID: "",
		})

		assert.Nil(t, resp)
		assert.Equal(t, utils.ErrUUIDMissing, err)
	})

	t.Run("error - invalid cursor", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		handler := &ListLineItemsHandler{
			BillRepo:     mocks.NewMockBillRepository(ctrl),
			LineItemRepo: mocks.NewMockLineItemRepository(ctrl),
		}

		resp, err := handler.Handle(context.Background(), &dto.ListLineItemsRequest{
			BillUUID: "bill-123",
			Cursor:   "invalid-not-base64-json",
		})

		assert.Nil(t, resp)
		assert.Equal(t, utils.ErrInvalidCursor, err)
	})

	t.Run("error - bill not found", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockBillRepo := mocks.NewMockBillRepository(ctrl)
		mockLineItemRepo := mocks.NewMockLineItemRepository(ctrl)

		handler := &ListLineItemsHandler{
			BillRepo:     mockBillRepo,
			LineItemRepo: mockLineItemRepo,
		}

		mockBillRepo.EXPECT().
			FetchByUUID(gomock.Any(), "nonexistent").
			Return(nil, sqldb.ErrNoRows)

		resp, err := handler.Handle(context.Background(), &dto.ListLineItemsRequest{
			BillUUID: "nonexistent",
		})

		assert.Nil(t, resp)
		assert.Equal(t, utils.ErrBillNotFoundAPI, err)
	})

	t.Run("error - internal error fetching bill", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockBillRepo := mocks.NewMockBillRepository(ctrl)
		mockLineItemRepo := mocks.NewMockLineItemRepository(ctrl)

		handler := &ListLineItemsHandler{
			BillRepo:     mockBillRepo,
			LineItemRepo: mockLineItemRepo,
		}

		mockBillRepo.EXPECT().
			FetchByUUID(gomock.Any(), "bill-123").
			Return(nil, assert.AnError)

		resp, err := handler.Handle(context.Background(), &dto.ListLineItemsRequest{
			BillUUID: "bill-123",
		})

		assert.Nil(t, resp)
		assert.Equal(t, utils.ErrInternal, err)
	})

	t.Run("error - internal error fetching line items", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockBillRepo := mocks.NewMockBillRepository(ctrl)
		mockLineItemRepo := mocks.NewMockLineItemRepository(ctrl)

		handler := &ListLineItemsHandler{
			BillRepo:     mockBillRepo,
			LineItemRepo: mockLineItemRepo,
		}

		billUUID := "bill-123"
		bill := &entity.BillEntity{
			UUID:     billUUID,
			Currency: "USD",
		}

		mockBillRepo.EXPECT().
			FetchByUUID(gomock.Any(), billUUID).
			Return(bill, nil)

		mockLineItemRepo.EXPECT().
			FetchByBillUUID(gomock.Any(), billUUID, time.Time{}, int64(0), 21).
			Return(nil, assert.AnError)

		resp, err := handler.Handle(context.Background(), &dto.ListLineItemsRequest{
			BillUUID: billUUID,
		})

		assert.Nil(t, resp)
		assert.Equal(t, utils.ErrInternal, err)
	})

	t.Run("success - returns empty list", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockBillRepo := mocks.NewMockBillRepository(ctrl)
		mockLineItemRepo := mocks.NewMockLineItemRepository(ctrl)

		handler := &ListLineItemsHandler{
			BillRepo:     mockBillRepo,
			LineItemRepo: mockLineItemRepo,
		}

		billUUID := "bill-123"
		bill := &entity.BillEntity{
			UUID:     billUUID,
			Currency: "USD",
		}

		mockBillRepo.EXPECT().
			FetchByUUID(gomock.Any(), billUUID).
			Return(bill, nil)

		mockLineItemRepo.EXPECT().
			FetchByBillUUID(gomock.Any(), billUUID, time.Time{}, int64(0), 21).
			Return([]*entity.LineItemEntity{}, nil)

		resp, err := handler.Handle(context.Background(), &dto.ListLineItemsRequest{
			BillUUID: billUUID,
		})

		require.NoError(t, err)
		assert.Len(t, resp.Data, 0)
		assert.False(t, resp.Pagination.HasMore)
	})

	t.Run("success - uses encoded cursor", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockBillRepo := mocks.NewMockBillRepository(ctrl)
		mockLineItemRepo := mocks.NewMockLineItemRepository(ctrl)

		handler := &ListLineItemsHandler{
			BillRepo:     mockBillRepo,
			LineItemRepo: mockLineItemRepo,
		}

		billUUID := "bill-123"
		cursorTime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
		cursorID := int64(1)
		cursor := utils.EncodeCursor(cursorTime, cursorID)

		bill := &entity.BillEntity{
			UUID:     billUUID,
			Currency: "USD",
		}

		lineItems := []*entity.LineItemEntity{
			{
				ID:          2,
				UUID:        "item-2",
				BillUUID:    billUUID,
				FeeType:     "TRANSACTION",
				AmountCents: 2000,
				CreatedAt:   time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
			},
		}

		mockBillRepo.EXPECT().
			FetchByUUID(gomock.Any(), billUUID).
			Return(bill, nil)

		mockLineItemRepo.EXPECT().
			FetchByBillUUID(gomock.Any(), billUUID, cursorTime, cursorID, 21).
			Return(lineItems, nil)

		resp, err := handler.Handle(context.Background(), &dto.ListLineItemsRequest{
			BillUUID: billUUID,
			Cursor:   cursor,
		})

		require.NoError(t, err)
		assert.Len(t, resp.Data, 1)
		assert.Equal(t, "item-2", resp.Data[0].UUID)
	})
}
