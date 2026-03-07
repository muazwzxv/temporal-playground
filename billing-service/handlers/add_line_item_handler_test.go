package handlers

import (
	"context"
	"testing"
	"time"

	"encore.app/db/repository/mocks"
	"encore.app/dto"
	"encore.app/entity"
	tbill "encore.app/temporal/bill"
	temporalmocks "encore.app/temporal/mocks"
	"encore.app/utils"

	"encore.dev/storage/sqldb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.temporal.io/api/serviceerror"
	"go.uber.org/mock/gomock"
)

func TestAddLineItemHandler_Handle(t *testing.T) {
	t.Run("success - adds line item", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockBillRepo := mocks.NewMockBillRepository(ctrl)
		mockLineItemRepo := mocks.NewMockLineItemRepository(ctrl)
		mockTemporalClient := temporalmocks.NewMockWorkflowClient(ctrl)

		handler := &AddLineItemHandler{
			BillRepo:       mockBillRepo,
			LineItemRepo:   mockLineItemRepo,
			TemporalClient: mockTemporalClient,
		}

		billUUID := "bill-123"
		bill := &entity.BillEntity{
			UUID:     billUUID,
			Status:   "OPEN",
			Currency: "USD",
		}

		mockBillRepo.EXPECT().
			FetchByUUID(gomock.Any(), billUUID).
			Return(bill, nil)

		mockLineItemRepo.EXPECT().
			FetchByBillAndKey(gomock.Any(), billUUID, "idem-key").
			Return(nil, sqldb.ErrNoRows)

		mockTemporalClient.EXPECT().
			QueryWorkflow(gomock.Any(), "bill-"+billUUID, "", tbill.QueryGetBillState).
			Return(newMockEncodedValue(tbill.BillStateQuery{Status: "OPEN"}), nil)

		mockTemporalClient.EXPECT().
			SignalWorkflow(gomock.Any(), "bill-"+billUUID, "", tbill.SignalAddLineItem, gomock.Any()).
			Return(nil)

		resp, err := handler.Handle(context.Background(), &dto.AddLineItemRequest{
			BillUUID:       billUUID,
			IdempotencyKey: "idem-key",
			FeeType:        "TRANSACTION",
			Description:    "Test transaction",
			Amount: dto.Money{
				Amount:   1000,
				Currency: "USD",
			},
		})

		require.NoError(t, err)
		assert.NotEmpty(t, resp.UUID)
		assert.Equal(t, "TRANSACTION", resp.FeeType)
		assert.Equal(t, int64(1000), resp.Amount.Amount)
		assert.Equal(t, "USD", resp.Amount.Currency)
		assert.Equal(t, "pending", resp.Status)
	})

	t.Run("error - validation fails - missing bill UUID", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		handler := &AddLineItemHandler{
			BillRepo:       mocks.NewMockBillRepository(ctrl),
			LineItemRepo:   mocks.NewMockLineItemRepository(ctrl),
			TemporalClient: temporalmocks.NewMockWorkflowClient(ctrl),
		}

		resp, err := handler.Handle(context.Background(), &dto.AddLineItemRequest{
			BillUUID:       "",
			IdempotencyKey: "idem-key",
			FeeType:        "TRANSACTION",
			Amount: dto.Money{
				Amount:   1000,
				Currency: "USD",
			},
		})

		assert.Nil(t, resp)
		assert.NotNil(t, err)
	})

	t.Run("error - validation fails - missing idempotency key", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		handler := &AddLineItemHandler{
			BillRepo:       mocks.NewMockBillRepository(ctrl),
			LineItemRepo:   mocks.NewMockLineItemRepository(ctrl),
			TemporalClient: temporalmocks.NewMockWorkflowClient(ctrl),
		}

		resp, err := handler.Handle(context.Background(), &dto.AddLineItemRequest{
			BillUUID:       "bill-123",
			IdempotencyKey: "",
			FeeType:        "TRANSACTION",
			Amount: dto.Money{
				Amount:   1000,
				Currency: "USD",
			},
		})

		assert.Nil(t, resp)
		assert.NotNil(t, err)
	})

	t.Run("error - validation fails - invalid amount", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		handler := &AddLineItemHandler{
			BillRepo:       mocks.NewMockBillRepository(ctrl),
			LineItemRepo:   mocks.NewMockLineItemRepository(ctrl),
			TemporalClient: temporalmocks.NewMockWorkflowClient(ctrl),
		}

		resp, err := handler.Handle(context.Background(), &dto.AddLineItemRequest{
			BillUUID:       "bill-123",
			IdempotencyKey: "idem-key",
			FeeType:        "TRANSACTION",
			Amount: dto.Money{
				Amount:   0, // Invalid amount
				Currency: "USD",
			},
		})

		assert.Nil(t, resp)
		assert.NotNil(t, err)
	})

	t.Run("error - bill not found", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockBillRepo := mocks.NewMockBillRepository(ctrl)
		mockLineItemRepo := mocks.NewMockLineItemRepository(ctrl)
		mockTemporalClient := temporalmocks.NewMockWorkflowClient(ctrl)

		handler := &AddLineItemHandler{
			BillRepo:       mockBillRepo,
			LineItemRepo:   mockLineItemRepo,
			TemporalClient: mockTemporalClient,
		}

		mockBillRepo.EXPECT().
			FetchByUUID(gomock.Any(), "bill-123").
			Return(nil, sqldb.ErrNoRows)

		resp, err := handler.Handle(context.Background(), &dto.AddLineItemRequest{
			BillUUID:       "bill-123",
			IdempotencyKey: "idem-key",
			FeeType:        "TRANSACTION",
			Amount: dto.Money{
				Amount:   1000,
				Currency: "USD",
			},
		})

		assert.Nil(t, resp)
		assert.Equal(t, utils.ErrBillNotFoundAPI, err)
	})

	t.Run("error - bill closed", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockBillRepo := mocks.NewMockBillRepository(ctrl)
		mockLineItemRepo := mocks.NewMockLineItemRepository(ctrl)
		mockTemporalClient := temporalmocks.NewMockWorkflowClient(ctrl)

		handler := &AddLineItemHandler{
			BillRepo:       mockBillRepo,
			LineItemRepo:   mockLineItemRepo,
			TemporalClient: mockTemporalClient,
		}

		bill := &entity.BillEntity{
			UUID:   "bill-123",
			Status: "CLOSED",
		}

		mockBillRepo.EXPECT().
			FetchByUUID(gomock.Any(), "bill-123").
			Return(bill, nil)

		resp, err := handler.Handle(context.Background(), &dto.AddLineItemRequest{
			BillUUID:       "bill-123",
			IdempotencyKey: "idem-key",
			FeeType:        "TRANSACTION",
			Amount: dto.Money{
				Amount:   1000,
				Currency: "USD",
			},
		})

		assert.Nil(t, resp)
		assert.Equal(t, utils.ErrBillClosed, err)
	})

	t.Run("error - currency mismatch", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockBillRepo := mocks.NewMockBillRepository(ctrl)
		mockLineItemRepo := mocks.NewMockLineItemRepository(ctrl)
		mockTemporalClient := temporalmocks.NewMockWorkflowClient(ctrl)

		handler := &AddLineItemHandler{
			BillRepo:       mockBillRepo,
			LineItemRepo:   mockLineItemRepo,
			TemporalClient: mockTemporalClient,
		}

		bill := &entity.BillEntity{
			UUID:     "bill-123",
			Status:   "OPEN",
			Currency: "USD",
		}

		mockBillRepo.EXPECT().
			FetchByUUID(gomock.Any(), "bill-123").
			Return(bill, nil)

		resp, err := handler.Handle(context.Background(), &dto.AddLineItemRequest{
			BillUUID:       "bill-123",
			IdempotencyKey: "idem-key",
			FeeType:        "TRANSACTION",
			Amount: dto.Money{
				Amount:   1000,
				Currency: "GEL", // Mismatched currency
			},
		})

		assert.Nil(t, resp)
		assert.Equal(t, utils.ErrCurrencyMismatch, err)
	})

	t.Run("idempotent - returns existing line item", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockBillRepo := mocks.NewMockBillRepository(ctrl)
		mockLineItemRepo := mocks.NewMockLineItemRepository(ctrl)
		mockTemporalClient := temporalmocks.NewMockWorkflowClient(ctrl)

		handler := &AddLineItemHandler{
			BillRepo:       mockBillRepo,
			LineItemRepo:   mockLineItemRepo,
			TemporalClient: mockTemporalClient,
		}

		billUUID := "bill-123"
		bill := &entity.BillEntity{
			UUID:     billUUID,
			Status:   "OPEN",
			Currency: "USD",
		}

		existingLineItem := &entity.LineItemEntity{
			UUID:        "existing-line-item",
			BillUUID:    billUUID,
			FeeType:     "TRANSACTION",
			AmountCents: 1000,
			CreatedAt:   time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		}

		mockBillRepo.EXPECT().
			FetchByUUID(gomock.Any(), billUUID).
			Return(bill, nil)

		mockLineItemRepo.EXPECT().
			FetchByBillAndKey(gomock.Any(), billUUID, "idem-key").
			Return(existingLineItem, nil)

		resp, err := handler.Handle(context.Background(), &dto.AddLineItemRequest{
			BillUUID:       billUUID,
			IdempotencyKey: "idem-key",
			FeeType:        "TRANSACTION",
			Amount: dto.Money{
				Amount:   1000,
				Currency: "USD",
			},
		})

		require.NoError(t, err)
		assert.Equal(t, "existing-line-item", resp.UUID)
		assert.Equal(t, "TRANSACTION", resp.FeeType)
		assert.Equal(t, int64(1000), resp.Amount.Amount)
		assert.Equal(t, "persisted", resp.Status)
	})

	t.Run("error - workflow not found", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockBillRepo := mocks.NewMockBillRepository(ctrl)
		mockLineItemRepo := mocks.NewMockLineItemRepository(ctrl)
		mockTemporalClient := temporalmocks.NewMockWorkflowClient(ctrl)

		handler := &AddLineItemHandler{
			BillRepo:       mockBillRepo,
			LineItemRepo:   mockLineItemRepo,
			TemporalClient: mockTemporalClient,
		}

		billUUID := "bill-123"
		bill := &entity.BillEntity{
			UUID:     billUUID,
			Status:   "OPEN",
			Currency: "USD",
		}

		mockBillRepo.EXPECT().
			FetchByUUID(gomock.Any(), billUUID).
			Return(bill, nil)

		mockLineItemRepo.EXPECT().
			FetchByBillAndKey(gomock.Any(), billUUID, "idem-key").
			Return(nil, sqldb.ErrNoRows)

		mockTemporalClient.EXPECT().
			QueryWorkflow(gomock.Any(), "bill-"+billUUID, "", tbill.QueryGetBillState).
			Return(nil, &serviceerror.NotFound{})

		resp, err := handler.Handle(context.Background(), &dto.AddLineItemRequest{
			BillUUID:       billUUID,
			IdempotencyKey: "idem-key",
			FeeType:        "TRANSACTION",
			Amount: dto.Money{
				Amount:   1000,
				Currency: "USD",
			},
		})

		assert.Nil(t, resp)
		assert.Equal(t, utils.ErrWorkflowNotFound, err)
	})

	t.Run("error - workflow query failed", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockBillRepo := mocks.NewMockBillRepository(ctrl)
		mockLineItemRepo := mocks.NewMockLineItemRepository(ctrl)
		mockTemporalClient := temporalmocks.NewMockWorkflowClient(ctrl)

		handler := &AddLineItemHandler{
			BillRepo:       mockBillRepo,
			LineItemRepo:   mockLineItemRepo,
			TemporalClient: mockTemporalClient,
		}

		billUUID := "bill-123"
		bill := &entity.BillEntity{
			UUID:     billUUID,
			Status:   "OPEN",
			Currency: "USD",
		}

		mockBillRepo.EXPECT().
			FetchByUUID(gomock.Any(), billUUID).
			Return(bill, nil)

		mockLineItemRepo.EXPECT().
			FetchByBillAndKey(gomock.Any(), billUUID, "idem-key").
			Return(nil, sqldb.ErrNoRows)

		mockTemporalClient.EXPECT().
			QueryWorkflow(gomock.Any(), "bill-"+billUUID, "", tbill.QueryGetBillState).
			Return(nil, assert.AnError)

		resp, err := handler.Handle(context.Background(), &dto.AddLineItemRequest{
			BillUUID:       billUUID,
			IdempotencyKey: "idem-key",
			FeeType:        "TRANSACTION",
			Amount: dto.Money{
				Amount:   1000,
				Currency: "USD",
			},
		})

		assert.Nil(t, resp)
		assert.Equal(t, utils.ErrWorkflowQueryFailed, err)
	})

	t.Run("error - workflow signal failed", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockBillRepo := mocks.NewMockBillRepository(ctrl)
		mockLineItemRepo := mocks.NewMockLineItemRepository(ctrl)
		mockTemporalClient := temporalmocks.NewMockWorkflowClient(ctrl)

		handler := &AddLineItemHandler{
			BillRepo:       mockBillRepo,
			LineItemRepo:   mockLineItemRepo,
			TemporalClient: mockTemporalClient,
		}

		billUUID := "bill-123"
		bill := &entity.BillEntity{
			UUID:     billUUID,
			Status:   "OPEN",
			Currency: "USD",
		}

		mockBillRepo.EXPECT().
			FetchByUUID(gomock.Any(), billUUID).
			Return(bill, nil)

		mockLineItemRepo.EXPECT().
			FetchByBillAndKey(gomock.Any(), billUUID, "idem-key").
			Return(nil, sqldb.ErrNoRows)

		mockTemporalClient.EXPECT().
			QueryWorkflow(gomock.Any(), "bill-"+billUUID, "", tbill.QueryGetBillState).
			Return(newMockEncodedValue(tbill.BillStateQuery{Status: "OPEN"}), nil)

		mockTemporalClient.EXPECT().
			SignalWorkflow(gomock.Any(), "bill-"+billUUID, "", tbill.SignalAddLineItem, gomock.Any()).
			Return(assert.AnError)

		resp, err := handler.Handle(context.Background(), &dto.AddLineItemRequest{
			BillUUID:       billUUID,
			IdempotencyKey: "idem-key",
			FeeType:        "TRANSACTION",
			Amount: dto.Money{
				Amount:   1000,
				Currency: "USD",
			},
		})

		assert.Nil(t, resp)
		assert.Equal(t, utils.ErrWorkflowSignalFailed, err)
	})

	t.Run("error - workflow completed between query and signal", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockBillRepo := mocks.NewMockBillRepository(ctrl)
		mockLineItemRepo := mocks.NewMockLineItemRepository(ctrl)
		mockTemporalClient := temporalmocks.NewMockWorkflowClient(ctrl)

		handler := &AddLineItemHandler{
			BillRepo:       mockBillRepo,
			LineItemRepo:   mockLineItemRepo,
			TemporalClient: mockTemporalClient,
		}

		billUUID := "bill-123"
		bill := &entity.BillEntity{
			UUID:     billUUID,
			Status:   "OPEN",
			Currency: "USD",
		}

		mockBillRepo.EXPECT().
			FetchByUUID(gomock.Any(), billUUID).
			Return(bill, nil)

		mockLineItemRepo.EXPECT().
			FetchByBillAndKey(gomock.Any(), billUUID, "idem-key").
			Return(nil, sqldb.ErrNoRows)

		mockTemporalClient.EXPECT().
			QueryWorkflow(gomock.Any(), "bill-"+billUUID, "", tbill.QueryGetBillState).
			Return(newMockEncodedValue(tbill.BillStateQuery{Status: "OPEN"}), nil)

		// Workflow completed between query and signal
		mockTemporalClient.EXPECT().
			SignalWorkflow(gomock.Any(), "bill-"+billUUID, "", tbill.SignalAddLineItem, gomock.Any()).
			Return(&serviceerror.NotFound{})

		resp, err := handler.Handle(context.Background(), &dto.AddLineItemRequest{
			BillUUID:       billUUID,
			IdempotencyKey: "idem-key",
			FeeType:        "TRANSACTION",
			Amount: dto.Money{
				Amount:   1000,
				Currency: "USD",
			},
		})

		assert.Nil(t, resp)
		assert.Equal(t, utils.ErrBillClosed, err)
	})
}
