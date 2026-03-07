package handlers

import (
	"context"
	"testing"

	"encore.app/db/repository/mocks"
	"encore.app/dto"
	"encore.app/entity"
	temporalmocks "encore.app/temporal/mocks"
	"encore.app/utils"

	"encore.dev/storage/sqldb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.temporal.io/api/serviceerror"
	"go.uber.org/mock/gomock"
)

func TestCloseBillHandler_Handle(t *testing.T) {
	t.Run("success - initiates bill close", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockBillRepo := mocks.NewMockBillRepository(ctrl)
		mockTemporalClient := temporalmocks.NewMockWorkflowClient(ctrl)

		handler := &CloseBillHandler{
			BillRepo:       mockBillRepo,
			TemporalClient: mockTemporalClient,
		}

		billUUID := "bill-123"
		bill := &entity.BillEntity{
			UUID:   billUUID,
			Status: "OPEN",
		}

		mockBillRepo.EXPECT().
			FetchByUUID(gomock.Any(), billUUID).
			Return(bill, nil)

		mockTemporalClient.EXPECT().
			SignalWorkflow(gomock.Any(), "bill-"+billUUID, "", "close_bill", nil).
			Return(nil)

		resp, err := handler.Handle(context.Background(), &dto.CloseBillRequest{UUID: billUUID})

		require.NoError(t, err)
		assert.Equal(t, billUUID, resp.UUID)
		assert.Equal(t, "CLOSING", resp.Status)
		assert.Contains(t, resp.Message, "Poll GET /v1/bill/get")
	})

	t.Run("error - missing UUID", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		handler := &CloseBillHandler{
			BillRepo:       mocks.NewMockBillRepository(ctrl),
			TemporalClient: temporalmocks.NewMockWorkflowClient(ctrl),
		}

		resp, err := handler.Handle(context.Background(), &dto.CloseBillRequest{UUID: ""})

		assert.Nil(t, resp)
		assert.Equal(t, utils.ErrUUIDMissing, err)
	})

	t.Run("error - bill not found", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockBillRepo := mocks.NewMockBillRepository(ctrl)
		mockTemporalClient := temporalmocks.NewMockWorkflowClient(ctrl)

		handler := &CloseBillHandler{
			BillRepo:       mockBillRepo,
			TemporalClient: mockTemporalClient,
		}

		billUUID := "nonexistent-bill"

		mockBillRepo.EXPECT().
			FetchByUUID(gomock.Any(), billUUID).
			Return(nil, sqldb.ErrNoRows)

		resp, err := handler.Handle(context.Background(), &dto.CloseBillRequest{UUID: billUUID})

		assert.Nil(t, resp)
		assert.Equal(t, utils.ErrBillNotFoundAPI, err)
	})

	t.Run("error - bill already closed", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockBillRepo := mocks.NewMockBillRepository(ctrl)
		mockTemporalClient := temporalmocks.NewMockWorkflowClient(ctrl)

		handler := &CloseBillHandler{
			BillRepo:       mockBillRepo,
			TemporalClient: mockTemporalClient,
		}

		billUUID := "bill-123"
		bill := &entity.BillEntity{
			UUID:   billUUID,
			Status: "CLOSED",
		}

		mockBillRepo.EXPECT().
			FetchByUUID(gomock.Any(), billUUID).
			Return(bill, nil)

		resp, err := handler.Handle(context.Background(), &dto.CloseBillRequest{UUID: billUUID})

		assert.Nil(t, resp)
		assert.Equal(t, utils.ErrBillAlreadyClosedAPI, err)
	})

	t.Run("error - workflow not found", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockBillRepo := mocks.NewMockBillRepository(ctrl)
		mockTemporalClient := temporalmocks.NewMockWorkflowClient(ctrl)

		handler := &CloseBillHandler{
			BillRepo:       mockBillRepo,
			TemporalClient: mockTemporalClient,
		}

		billUUID := "bill-123"
		bill := &entity.BillEntity{
			UUID:   billUUID,
			Status: "OPEN",
		}

		mockBillRepo.EXPECT().
			FetchByUUID(gomock.Any(), billUUID).
			Return(bill, nil)

		mockTemporalClient.EXPECT().
			SignalWorkflow(gomock.Any(), "bill-"+billUUID, "", "close_bill", nil).
			Return(&serviceerror.NotFound{})

		resp, err := handler.Handle(context.Background(), &dto.CloseBillRequest{UUID: billUUID})

		assert.Nil(t, resp)
		assert.Equal(t, utils.ErrBillAlreadyClosedAPI, err)
	})

	t.Run("error - workflow signal failed", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockBillRepo := mocks.NewMockBillRepository(ctrl)
		mockTemporalClient := temporalmocks.NewMockWorkflowClient(ctrl)

		handler := &CloseBillHandler{
			BillRepo:       mockBillRepo,
			TemporalClient: mockTemporalClient,
		}

		billUUID := "bill-123"
		bill := &entity.BillEntity{
			UUID:   billUUID,
			Status: "OPEN",
		}

		mockBillRepo.EXPECT().
			FetchByUUID(gomock.Any(), billUUID).
			Return(bill, nil)

		mockTemporalClient.EXPECT().
			SignalWorkflow(gomock.Any(), "bill-"+billUUID, "", "close_bill", nil).
			Return(assert.AnError)

		resp, err := handler.Handle(context.Background(), &dto.CloseBillRequest{UUID: billUUID})

		assert.Nil(t, resp)
		assert.Equal(t, utils.ErrWorkflowSignalFailed, err)
	})

	t.Run("error - internal error fetching bill", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockBillRepo := mocks.NewMockBillRepository(ctrl)
		mockTemporalClient := temporalmocks.NewMockWorkflowClient(ctrl)

		handler := &CloseBillHandler{
			BillRepo:       mockBillRepo,
			TemporalClient: mockTemporalClient,
		}

		billUUID := "bill-123"

		mockBillRepo.EXPECT().
			FetchByUUID(gomock.Any(), billUUID).
			Return(nil, assert.AnError)

		resp, err := handler.Handle(context.Background(), &dto.CloseBillRequest{UUID: billUUID})

		assert.Nil(t, resp)
		assert.Equal(t, utils.ErrInternal, err)
	})
}
