package handlers

import (
	"context"
	"testing"
	"time"

	"encore.app/db/repository/mocks"
	"encore.app/dto"
	"encore.app/entity"
	temporalmocks "encore.app/temporal/mocks"
	"encore.app/utils"

	"encore.dev/storage/sqldb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.temporal.io/api/serviceerror"
	"go.temporal.io/sdk/client"
	"go.uber.org/mock/gomock"
)

// mockWorkflowRun implements client.WorkflowRun for testing
type mockWorkflowRun struct {
	workflowID string
	runID      string
}

func (m *mockWorkflowRun) GetID() string {
	return m.workflowID
}

func (m *mockWorkflowRun) GetRunID() string {
	return m.runID
}

func (m *mockWorkflowRun) Get(ctx context.Context, valuePtr interface{}) error {
	return nil
}

func (m *mockWorkflowRun) GetWithOptions(ctx context.Context, valuePtr interface{}, options client.WorkflowRunGetOptions) error {
	return nil
}

func TestCreateBillHandler_Handle(t *testing.T) {
	t.Run("success - creates bill and starts workflow", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockBillRepo := mocks.NewMockBillRepository(ctrl)
		mockCustomerRepo := mocks.NewMockCustomerRepository(ctrl)
		mockTemporalClient := temporalmocks.NewMockWorkflowClient(ctrl)

		handler := &CreateBillHandler{
			BillRepo:       mockBillRepo,
			CustomerRepo:   mockCustomerRepo,
			TemporalClient: mockTemporalClient,
		}

		customerUUID := "customer-123"
		billUUID := "bill-123"

		mockCustomerRepo.EXPECT().
			FetchByUUID(gomock.Any(), customerUUID).
			Return(&entity.CustomerEntity{UUID: customerUUID}, nil)

		mockBillRepo.EXPECT().
			FetchByUUID(gomock.Any(), billUUID).
			Return(nil, sqldb.ErrNoRows)

		mockBillRepo.EXPECT().
			Insert(gomock.Any(), gomock.Any()).
			Return(nil)

		mockTemporalClient.EXPECT().
			ExecuteWorkflow(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Return(&mockWorkflowRun{workflowID: "bill-" + billUUID}, nil)

		resp, err := handler.Handle(context.Background(), &dto.CreateBillRequest{
			UUID:         billUUID,
			CustomerUUID: customerUUID,
			Currency:     "USD",
			PeriodStart:  "2024-01-01T00:00:00Z",
			PeriodEnd:    "2024-01-31T23:59:59Z",
		})

		require.NoError(t, err)
		assert.Equal(t, billUUID, resp.UUID)
		assert.Equal(t, "OPEN", resp.Status)
		assert.Equal(t, "USD", resp.Currency)
	})

	t.Run("error - validation fails - missing UUID", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		handler := &CreateBillHandler{
			BillRepo:       mocks.NewMockBillRepository(ctrl),
			CustomerRepo:   mocks.NewMockCustomerRepository(ctrl),
			TemporalClient: temporalmocks.NewMockWorkflowClient(ctrl),
		}

		resp, err := handler.Handle(context.Background(), &dto.CreateBillRequest{
			UUID:         "",
			CustomerUUID: "customer-123",
			Currency:     "USD",
			PeriodStart:  "2024-01-01T00:00:00Z",
			PeriodEnd:    "2024-01-31T23:59:59Z",
		})

		assert.Nil(t, resp)
		assert.NotNil(t, err)
	})

	t.Run("error - validation fails - invalid currency", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		handler := &CreateBillHandler{
			BillRepo:       mocks.NewMockBillRepository(ctrl),
			CustomerRepo:   mocks.NewMockCustomerRepository(ctrl),
			TemporalClient: temporalmocks.NewMockWorkflowClient(ctrl),
		}

		resp, err := handler.Handle(context.Background(), &dto.CreateBillRequest{
			UUID:         "bill-123",
			CustomerUUID: "customer-123",
			Currency:     "EUR", // Invalid - only USD and GEL supported
			PeriodStart:  "2024-01-01T00:00:00Z",
			PeriodEnd:    "2024-01-31T23:59:59Z",
		})

		assert.Nil(t, resp)
		assert.NotNil(t, err)
	})

	t.Run("error - validation fails - period end before period start", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		handler := &CreateBillHandler{
			BillRepo:       mocks.NewMockBillRepository(ctrl),
			CustomerRepo:   mocks.NewMockCustomerRepository(ctrl),
			TemporalClient: temporalmocks.NewMockWorkflowClient(ctrl),
		}

		resp, err := handler.Handle(context.Background(), &dto.CreateBillRequest{
			UUID:         "bill-123",
			CustomerUUID: "customer-123",
			Currency:     "USD",
			PeriodStart:  "2024-01-31T23:59:59Z", // Start after end
			PeriodEnd:    "2024-01-01T00:00:00Z",
		})

		assert.Nil(t, resp)
		assert.NotNil(t, err)
	})

	t.Run("error - customer not found", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockBillRepo := mocks.NewMockBillRepository(ctrl)
		mockCustomerRepo := mocks.NewMockCustomerRepository(ctrl)
		mockTemporalClient := temporalmocks.NewMockWorkflowClient(ctrl)

		handler := &CreateBillHandler{
			BillRepo:       mockBillRepo,
			CustomerRepo:   mockCustomerRepo,
			TemporalClient: mockTemporalClient,
		}

		mockCustomerRepo.EXPECT().
			FetchByUUID(gomock.Any(), "customer-123").
			Return(nil, sqldb.ErrNoRows)

		resp, err := handler.Handle(context.Background(), &dto.CreateBillRequest{
			UUID:         "bill-123",
			CustomerUUID: "customer-123",
			Currency:     "USD",
			PeriodStart:  "2024-01-01T00:00:00Z",
			PeriodEnd:    "2024-01-31T23:59:59Z",
		})

		assert.Nil(t, resp)
		assert.Equal(t, utils.ErrCustomerNotFoundAPI, err)
	})

	t.Run("idempotent - returns existing bill", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockBillRepo := mocks.NewMockBillRepository(ctrl)
		mockCustomerRepo := mocks.NewMockCustomerRepository(ctrl)
		mockTemporalClient := temporalmocks.NewMockWorkflowClient(ctrl)

		handler := &CreateBillHandler{
			BillRepo:       mockBillRepo,
			CustomerRepo:   mockCustomerRepo,
			TemporalClient: mockTemporalClient,
		}

		customerUUID := "customer-123"
		billUUID := "bill-123"

		existingBill := &entity.BillEntity{
			UUID:        billUUID,
			Status:      "OPEN",
			Currency:    "USD",
			PeriodStart: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			PeriodEnd:   time.Date(2024, 1, 31, 23, 59, 59, 0, time.UTC),
		}

		mockCustomerRepo.EXPECT().
			FetchByUUID(gomock.Any(), customerUUID).
			Return(&entity.CustomerEntity{UUID: customerUUID}, nil)

		mockBillRepo.EXPECT().
			FetchByUUID(gomock.Any(), billUUID).
			Return(existingBill, nil)

		resp, err := handler.Handle(context.Background(), &dto.CreateBillRequest{
			UUID:         billUUID,
			CustomerUUID: customerUUID,
			Currency:     "USD",
			PeriodStart:  "2024-01-01T00:00:00Z",
			PeriodEnd:    "2024-01-31T23:59:59Z",
		})

		require.NoError(t, err)
		assert.Equal(t, billUUID, resp.UUID)
		assert.Equal(t, "OPEN", resp.Status)
	})

	t.Run("error - bill insert failed", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockBillRepo := mocks.NewMockBillRepository(ctrl)
		mockCustomerRepo := mocks.NewMockCustomerRepository(ctrl)
		mockTemporalClient := temporalmocks.NewMockWorkflowClient(ctrl)

		handler := &CreateBillHandler{
			BillRepo:       mockBillRepo,
			CustomerRepo:   mockCustomerRepo,
			TemporalClient: mockTemporalClient,
		}

		customerUUID := "customer-123"
		billUUID := "bill-123"

		mockCustomerRepo.EXPECT().
			FetchByUUID(gomock.Any(), customerUUID).
			Return(&entity.CustomerEntity{UUID: customerUUID}, nil)

		mockBillRepo.EXPECT().
			FetchByUUID(gomock.Any(), billUUID).
			Return(nil, sqldb.ErrNoRows)

		mockBillRepo.EXPECT().
			Insert(gomock.Any(), gomock.Any()).
			Return(assert.AnError)

		resp, err := handler.Handle(context.Background(), &dto.CreateBillRequest{
			UUID:         billUUID,
			CustomerUUID: customerUUID,
			Currency:     "USD",
			PeriodStart:  "2024-01-01T00:00:00Z",
			PeriodEnd:    "2024-01-31T23:59:59Z",
		})

		assert.Nil(t, resp)
		assert.Equal(t, utils.ErrInternal, err)
	})

	t.Run("success - workflow already started returns success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockBillRepo := mocks.NewMockBillRepository(ctrl)
		mockCustomerRepo := mocks.NewMockCustomerRepository(ctrl)
		mockTemporalClient := temporalmocks.NewMockWorkflowClient(ctrl)

		handler := &CreateBillHandler{
			BillRepo:       mockBillRepo,
			CustomerRepo:   mockCustomerRepo,
			TemporalClient: mockTemporalClient,
		}

		customerUUID := "customer-123"
		billUUID := "bill-123"

		mockCustomerRepo.EXPECT().
			FetchByUUID(gomock.Any(), customerUUID).
			Return(&entity.CustomerEntity{UUID: customerUUID}, nil)

		mockBillRepo.EXPECT().
			FetchByUUID(gomock.Any(), billUUID).
			Return(nil, sqldb.ErrNoRows)

		mockBillRepo.EXPECT().
			Insert(gomock.Any(), gomock.Any()).
			Return(nil)

		// Workflow already started - this is a race condition scenario
		mockTemporalClient.EXPECT().
			ExecuteWorkflow(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Return(nil, &serviceerror.WorkflowExecutionAlreadyStarted{})

		resp, err := handler.Handle(context.Background(), &dto.CreateBillRequest{
			UUID:         billUUID,
			CustomerUUID: customerUUID,
			Currency:     "USD",
			PeriodStart:  "2024-01-01T00:00:00Z",
			PeriodEnd:    "2024-01-31T23:59:59Z",
		})

		require.NoError(t, err)
		assert.Equal(t, billUUID, resp.UUID)
		assert.Equal(t, "OPEN", resp.Status)
	})

	t.Run("error - workflow start failed", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockBillRepo := mocks.NewMockBillRepository(ctrl)
		mockCustomerRepo := mocks.NewMockCustomerRepository(ctrl)
		mockTemporalClient := temporalmocks.NewMockWorkflowClient(ctrl)

		handler := &CreateBillHandler{
			BillRepo:       mockBillRepo,
			CustomerRepo:   mockCustomerRepo,
			TemporalClient: mockTemporalClient,
		}

		customerUUID := "customer-123"
		billUUID := "bill-123"

		mockCustomerRepo.EXPECT().
			FetchByUUID(gomock.Any(), customerUUID).
			Return(&entity.CustomerEntity{UUID: customerUUID}, nil)

		mockBillRepo.EXPECT().
			FetchByUUID(gomock.Any(), billUUID).
			Return(nil, sqldb.ErrNoRows)

		mockBillRepo.EXPECT().
			Insert(gomock.Any(), gomock.Any()).
			Return(nil)

		mockTemporalClient.EXPECT().
			ExecuteWorkflow(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Return(nil, assert.AnError)

		resp, err := handler.Handle(context.Background(), &dto.CreateBillRequest{
			UUID:         billUUID,
			CustomerUUID: customerUUID,
			Currency:     "USD",
			PeriodStart:  "2024-01-01T00:00:00Z",
			PeriodEnd:    "2024-01-31T23:59:59Z",
		})

		assert.Nil(t, resp)
		assert.Equal(t, utils.ErrWorkflowStartFailed, err)
	})

	t.Run("success - creates bill with GEL currency", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockBillRepo := mocks.NewMockBillRepository(ctrl)
		mockCustomerRepo := mocks.NewMockCustomerRepository(ctrl)
		mockTemporalClient := temporalmocks.NewMockWorkflowClient(ctrl)

		handler := &CreateBillHandler{
			BillRepo:       mockBillRepo,
			CustomerRepo:   mockCustomerRepo,
			TemporalClient: mockTemporalClient,
		}

		customerUUID := "customer-123"
		billUUID := "bill-123"

		mockCustomerRepo.EXPECT().
			FetchByUUID(gomock.Any(), customerUUID).
			Return(&entity.CustomerEntity{UUID: customerUUID}, nil)

		mockBillRepo.EXPECT().
			FetchByUUID(gomock.Any(), billUUID).
			Return(nil, sqldb.ErrNoRows)

		mockBillRepo.EXPECT().
			Insert(gomock.Any(), gomock.Any()).
			Return(nil)

		mockTemporalClient.EXPECT().
			ExecuteWorkflow(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Return(&mockWorkflowRun{workflowID: "bill-" + billUUID}, nil)

		resp, err := handler.Handle(context.Background(), &dto.CreateBillRequest{
			UUID:         billUUID,
			CustomerUUID: customerUUID,
			Currency:     "GEL",
			PeriodStart:  "2024-01-01T00:00:00Z",
			PeriodEnd:    "2024-01-31T23:59:59Z",
		})

		require.NoError(t, err)
		assert.Equal(t, "GEL", resp.Currency)
	})
}
