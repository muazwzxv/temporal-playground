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
	"go.temporal.io/sdk/converter"
	"go.uber.org/mock/gomock"
)

// mockEncodedValue implements converter.EncodedValue for testing QueryWorkflow
type mockEncodedValue struct {
	value interface{}
}

func (m *mockEncodedValue) Get(valuePtr interface{}) error {
	// Type assert and assign the value
	if v, ok := valuePtr.(*tbill.BillStateQuery); ok && m.value != nil {
		*v = m.value.(tbill.BillStateQuery)
	}
	return nil
}

func (m *mockEncodedValue) HasValue() bool {
	return m.value != nil
}

func newMockEncodedValue(value interface{}) converter.EncodedValue {
	return &mockEncodedValue{value: value}
}

func TestReverseLineItemHandler_Handle(t *testing.T) {
	t.Run("success - reverses line item", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockBillRepo := mocks.NewMockBillRepository(ctrl)
		mockLineItemRepo := mocks.NewMockLineItemRepository(ctrl)
		mockTemporalClient := temporalmocks.NewMockWorkflowClient(ctrl)

		handler := &ReverseLineItemHandler{
			BillRepo:       mockBillRepo,
			LineItemRepo:   mockLineItemRepo,
			TemporalClient: mockTemporalClient,
		}

		billUUID := "bill-123"
		lineItemUUID := "line-item-456"

		bill := &entity.BillEntity{
			UUID:     billUUID,
			Status:   "OPEN",
			Currency: "USD",
		}

		originalLineItem := &entity.LineItemEntity{
			UUID:        lineItemUUID,
			BillUUID:    billUUID,
			FeeType:     "TRANSACTION",
			AmountCents: 1000,
			CreatedAt:   time.Now(),
		}

		// Mock expectations in order of execution
		mockBillRepo.EXPECT().
			FetchByUUID(gomock.Any(), billUUID).
			Return(bill, nil)

		mockLineItemRepo.EXPECT().
			FetchByUUID(gomock.Any(), lineItemUUID).
			Return(originalLineItem, nil)

		mockLineItemRepo.EXPECT().
			FetchReversalByOriginalUUID(gomock.Any(), lineItemUUID).
			Return(nil, sqldb.ErrNoRows)

		mockLineItemRepo.EXPECT().
			FetchByBillAndKey(gomock.Any(), billUUID, "idem-key").
			Return(nil, sqldb.ErrNoRows)

		mockTemporalClient.EXPECT().
			QueryWorkflow(gomock.Any(), "bill-"+billUUID, "", tbill.QueryGetBillState).
			Return(newMockEncodedValue(tbill.BillStateQuery{Status: "OPEN"}), nil)

		mockTemporalClient.EXPECT().
			SignalWorkflow(gomock.Any(), "bill-"+billUUID, "", tbill.SignalAddLineItem, gomock.Any()).
			Return(nil)

		resp, err := handler.Handle(context.Background(), &dto.ReverseLineItemRequest{
			BillUUID:       billUUID,
			LineItemUUID:   lineItemUUID,
			IdempotencyKey: "idem-key",
			Reason:         "test reason",
		})

		require.NoError(t, err)
		assert.NotEmpty(t, resp.UUID)
		assert.Equal(t, "REVERSAL", resp.FeeType)
		assert.Equal(t, lineItemUUID, resp.ReferenceUUID)
		assert.Equal(t, int64(-1000), resp.Amount.Amount)
		assert.Equal(t, "USD", resp.Amount.Currency)
	})

	t.Run("error - validation fails - missing bill UUID", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		handler := &ReverseLineItemHandler{
			BillRepo:       mocks.NewMockBillRepository(ctrl),
			LineItemRepo:   mocks.NewMockLineItemRepository(ctrl),
			TemporalClient: temporalmocks.NewMockWorkflowClient(ctrl),
		}

		resp, err := handler.Handle(context.Background(), &dto.ReverseLineItemRequest{
			BillUUID:       "",
			LineItemUUID:   "line-item-456",
			IdempotencyKey: "idem-key",
		})

		assert.Nil(t, resp)
		assert.NotNil(t, err)
	})

	t.Run("error - validation fails - missing line item UUID", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		handler := &ReverseLineItemHandler{
			BillRepo:       mocks.NewMockBillRepository(ctrl),
			LineItemRepo:   mocks.NewMockLineItemRepository(ctrl),
			TemporalClient: temporalmocks.NewMockWorkflowClient(ctrl),
		}

		resp, err := handler.Handle(context.Background(), &dto.ReverseLineItemRequest{
			BillUUID:       "bill-123",
			LineItemUUID:   "",
			IdempotencyKey: "idem-key",
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

		handler := &ReverseLineItemHandler{
			BillRepo:       mockBillRepo,
			LineItemRepo:   mockLineItemRepo,
			TemporalClient: mockTemporalClient,
		}

		mockBillRepo.EXPECT().
			FetchByUUID(gomock.Any(), "bill-123").
			Return(nil, sqldb.ErrNoRows)

		resp, err := handler.Handle(context.Background(), &dto.ReverseLineItemRequest{
			BillUUID:       "bill-123",
			LineItemUUID:   "line-item-456",
			IdempotencyKey: "idem-key",
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

		handler := &ReverseLineItemHandler{
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

		resp, err := handler.Handle(context.Background(), &dto.ReverseLineItemRequest{
			BillUUID:       "bill-123",
			LineItemUUID:   "line-item-456",
			IdempotencyKey: "idem-key",
		})

		assert.Nil(t, resp)
		assert.Equal(t, utils.ErrBillClosed, err)
	})

	t.Run("error - line item not found", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockBillRepo := mocks.NewMockBillRepository(ctrl)
		mockLineItemRepo := mocks.NewMockLineItemRepository(ctrl)
		mockTemporalClient := temporalmocks.NewMockWorkflowClient(ctrl)

		handler := &ReverseLineItemHandler{
			BillRepo:       mockBillRepo,
			LineItemRepo:   mockLineItemRepo,
			TemporalClient: mockTemporalClient,
		}

		bill := &entity.BillEntity{
			UUID:   "bill-123",
			Status: "OPEN",
		}

		mockBillRepo.EXPECT().
			FetchByUUID(gomock.Any(), "bill-123").
			Return(bill, nil)

		mockLineItemRepo.EXPECT().
			FetchByUUID(gomock.Any(), "line-item-456").
			Return(nil, sqldb.ErrNoRows)

		resp, err := handler.Handle(context.Background(), &dto.ReverseLineItemRequest{
			BillUUID:       "bill-123",
			LineItemUUID:   "line-item-456",
			IdempotencyKey: "idem-key",
		})

		assert.Nil(t, resp)
		assert.Equal(t, utils.ErrLineItemNotFoundAPI, err)
	})

	t.Run("error - line item belongs to different bill", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockBillRepo := mocks.NewMockBillRepository(ctrl)
		mockLineItemRepo := mocks.NewMockLineItemRepository(ctrl)
		mockTemporalClient := temporalmocks.NewMockWorkflowClient(ctrl)

		handler := &ReverseLineItemHandler{
			BillRepo:       mockBillRepo,
			LineItemRepo:   mockLineItemRepo,
			TemporalClient: mockTemporalClient,
		}

		bill := &entity.BillEntity{
			UUID:   "bill-123",
			Status: "OPEN",
		}

		lineItem := &entity.LineItemEntity{
			UUID:     "line-item-456",
			BillUUID: "different-bill", // Belongs to a different bill
		}

		mockBillRepo.EXPECT().
			FetchByUUID(gomock.Any(), "bill-123").
			Return(bill, nil)

		mockLineItemRepo.EXPECT().
			FetchByUUID(gomock.Any(), "line-item-456").
			Return(lineItem, nil)

		resp, err := handler.Handle(context.Background(), &dto.ReverseLineItemRequest{
			BillUUID:       "bill-123",
			LineItemUUID:   "line-item-456",
			IdempotencyKey: "idem-key",
		})

		assert.Nil(t, resp)
		assert.Equal(t, utils.ErrLineItemNotFoundAPI, err)
	})

	t.Run("error - cannot reverse a reversal", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockBillRepo := mocks.NewMockBillRepository(ctrl)
		mockLineItemRepo := mocks.NewMockLineItemRepository(ctrl)
		mockTemporalClient := temporalmocks.NewMockWorkflowClient(ctrl)

		handler := &ReverseLineItemHandler{
			BillRepo:       mockBillRepo,
			LineItemRepo:   mockLineItemRepo,
			TemporalClient: mockTemporalClient,
		}

		bill := &entity.BillEntity{
			UUID:   "bill-123",
			Status: "OPEN",
		}

		lineItem := &entity.LineItemEntity{
			UUID:     "line-item-456",
			BillUUID: "bill-123",
			FeeType:  "REVERSAL", // Already a reversal
		}

		mockBillRepo.EXPECT().
			FetchByUUID(gomock.Any(), "bill-123").
			Return(bill, nil)

		mockLineItemRepo.EXPECT().
			FetchByUUID(gomock.Any(), "line-item-456").
			Return(lineItem, nil)

		resp, err := handler.Handle(context.Background(), &dto.ReverseLineItemRequest{
			BillUUID:       "bill-123",
			LineItemUUID:   "line-item-456",
			IdempotencyKey: "idem-key",
		})

		assert.Nil(t, resp)
		assert.Equal(t, utils.ErrCannotReverseReversal, err)
	})

	t.Run("error - already reversed", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockBillRepo := mocks.NewMockBillRepository(ctrl)
		mockLineItemRepo := mocks.NewMockLineItemRepository(ctrl)
		mockTemporalClient := temporalmocks.NewMockWorkflowClient(ctrl)

		handler := &ReverseLineItemHandler{
			BillRepo:       mockBillRepo,
			LineItemRepo:   mockLineItemRepo,
			TemporalClient: mockTemporalClient,
		}

		bill := &entity.BillEntity{
			UUID:   "bill-123",
			Status: "OPEN",
		}

		lineItem := &entity.LineItemEntity{
			UUID:     "line-item-456",
			BillUUID: "bill-123",
			FeeType:  "TRANSACTION",
		}

		existingReversal := &entity.LineItemEntity{
			UUID:     "reversal-789",
			BillUUID: "bill-123",
			FeeType:  "REVERSAL",
		}

		mockBillRepo.EXPECT().
			FetchByUUID(gomock.Any(), "bill-123").
			Return(bill, nil)

		mockLineItemRepo.EXPECT().
			FetchByUUID(gomock.Any(), "line-item-456").
			Return(lineItem, nil)

		mockLineItemRepo.EXPECT().
			FetchReversalByOriginalUUID(gomock.Any(), "line-item-456").
			Return(existingReversal, nil)

		resp, err := handler.Handle(context.Background(), &dto.ReverseLineItemRequest{
			BillUUID:       "bill-123",
			LineItemUUID:   "line-item-456",
			IdempotencyKey: "idem-key",
		})

		assert.Nil(t, resp)
		assert.Equal(t, utils.ErrAlreadyReversedAPI, err)
	})

	t.Run("idempotent - returns existing reversal", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockBillRepo := mocks.NewMockBillRepository(ctrl)
		mockLineItemRepo := mocks.NewMockLineItemRepository(ctrl)
		mockTemporalClient := temporalmocks.NewMockWorkflowClient(ctrl)

		handler := &ReverseLineItemHandler{
			BillRepo:       mockBillRepo,
			LineItemRepo:   mockLineItemRepo,
			TemporalClient: mockTemporalClient,
		}

		billUUID := "bill-123"
		lineItemUUID := "line-item-456"
		originalUUID := lineItemUUID

		bill := &entity.BillEntity{
			UUID:     billUUID,
			Status:   "OPEN",
			Currency: "USD",
		}

		originalLineItem := &entity.LineItemEntity{
			UUID:        lineItemUUID,
			BillUUID:    billUUID,
			FeeType:     "TRANSACTION",
			AmountCents: 1000,
		}

		existingReversal := &entity.LineItemEntity{
			UUID:          "existing-reversal",
			BillUUID:      billUUID,
			FeeType:       "REVERSAL",
			AmountCents:   -1000,
			ReferenceUUID: &originalUUID,
			CreatedAt:     time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		}

		mockBillRepo.EXPECT().
			FetchByUUID(gomock.Any(), billUUID).
			Return(bill, nil)

		mockLineItemRepo.EXPECT().
			FetchByUUID(gomock.Any(), lineItemUUID).
			Return(originalLineItem, nil)

		mockLineItemRepo.EXPECT().
			FetchReversalByOriginalUUID(gomock.Any(), lineItemUUID).
			Return(nil, sqldb.ErrNoRows)

		mockLineItemRepo.EXPECT().
			FetchByBillAndKey(gomock.Any(), billUUID, "idem-key").
			Return(existingReversal, nil)

		resp, err := handler.Handle(context.Background(), &dto.ReverseLineItemRequest{
			BillUUID:       billUUID,
			LineItemUUID:   lineItemUUID,
			IdempotencyKey: "idem-key",
		})

		require.NoError(t, err)
		assert.Equal(t, "existing-reversal", resp.UUID)
		assert.Equal(t, "REVERSAL", resp.FeeType)
		assert.Equal(t, lineItemUUID, resp.ReferenceUUID)
		assert.Equal(t, int64(-1000), resp.Amount.Amount)
	})

	t.Run("error - workflow not found", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockBillRepo := mocks.NewMockBillRepository(ctrl)
		mockLineItemRepo := mocks.NewMockLineItemRepository(ctrl)
		mockTemporalClient := temporalmocks.NewMockWorkflowClient(ctrl)

		handler := &ReverseLineItemHandler{
			BillRepo:       mockBillRepo,
			LineItemRepo:   mockLineItemRepo,
			TemporalClient: mockTemporalClient,
		}

		billUUID := "bill-123"
		lineItemUUID := "line-item-456"

		bill := &entity.BillEntity{
			UUID:     billUUID,
			Status:   "OPEN",
			Currency: "USD",
		}

		originalLineItem := &entity.LineItemEntity{
			UUID:        lineItemUUID,
			BillUUID:    billUUID,
			FeeType:     "TRANSACTION",
			AmountCents: 1000,
		}

		mockBillRepo.EXPECT().
			FetchByUUID(gomock.Any(), billUUID).
			Return(bill, nil)

		mockLineItemRepo.EXPECT().
			FetchByUUID(gomock.Any(), lineItemUUID).
			Return(originalLineItem, nil)

		mockLineItemRepo.EXPECT().
			FetchReversalByOriginalUUID(gomock.Any(), lineItemUUID).
			Return(nil, sqldb.ErrNoRows)

		mockLineItemRepo.EXPECT().
			FetchByBillAndKey(gomock.Any(), billUUID, "idem-key").
			Return(nil, sqldb.ErrNoRows)

		mockTemporalClient.EXPECT().
			QueryWorkflow(gomock.Any(), "bill-"+billUUID, "", tbill.QueryGetBillState).
			Return(nil, &serviceerror.NotFound{})

		resp, err := handler.Handle(context.Background(), &dto.ReverseLineItemRequest{
			BillUUID:       billUUID,
			LineItemUUID:   lineItemUUID,
			IdempotencyKey: "idem-key",
		})

		assert.Nil(t, resp)
		assert.Equal(t, utils.ErrWorkflowNotFound, err)
	})

	t.Run("error - workflow signal failed", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockBillRepo := mocks.NewMockBillRepository(ctrl)
		mockLineItemRepo := mocks.NewMockLineItemRepository(ctrl)
		mockTemporalClient := temporalmocks.NewMockWorkflowClient(ctrl)

		handler := &ReverseLineItemHandler{
			BillRepo:       mockBillRepo,
			LineItemRepo:   mockLineItemRepo,
			TemporalClient: mockTemporalClient,
		}

		billUUID := "bill-123"
		lineItemUUID := "line-item-456"

		bill := &entity.BillEntity{
			UUID:     billUUID,
			Status:   "OPEN",
			Currency: "USD",
		}

		originalLineItem := &entity.LineItemEntity{
			UUID:        lineItemUUID,
			BillUUID:    billUUID,
			FeeType:     "TRANSACTION",
			AmountCents: 1000,
		}

		mockBillRepo.EXPECT().
			FetchByUUID(gomock.Any(), billUUID).
			Return(bill, nil)

		mockLineItemRepo.EXPECT().
			FetchByUUID(gomock.Any(), lineItemUUID).
			Return(originalLineItem, nil)

		mockLineItemRepo.EXPECT().
			FetchReversalByOriginalUUID(gomock.Any(), lineItemUUID).
			Return(nil, sqldb.ErrNoRows)

		mockLineItemRepo.EXPECT().
			FetchByBillAndKey(gomock.Any(), billUUID, "idem-key").
			Return(nil, sqldb.ErrNoRows)

		mockTemporalClient.EXPECT().
			QueryWorkflow(gomock.Any(), "bill-"+billUUID, "", tbill.QueryGetBillState).
			Return(newMockEncodedValue(tbill.BillStateQuery{Status: "OPEN"}), nil)

		mockTemporalClient.EXPECT().
			SignalWorkflow(gomock.Any(), "bill-"+billUUID, "", tbill.SignalAddLineItem, gomock.Any()).
			Return(assert.AnError)

		resp, err := handler.Handle(context.Background(), &dto.ReverseLineItemRequest{
			BillUUID:       billUUID,
			LineItemUUID:   lineItemUUID,
			IdempotencyKey: "idem-key",
		})

		assert.Nil(t, resp)
		assert.Equal(t, utils.ErrWorkflowSignalFailed, err)
	})
}
