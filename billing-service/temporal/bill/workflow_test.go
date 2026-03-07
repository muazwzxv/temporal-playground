package bill

import (
	"context"
	"testing"
	"time"

	"encore.app/db/repository/mocks"
	"encore.app/entity"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.temporal.io/sdk/testsuite"
	"go.uber.org/mock/gomock"
)

func TestBillWorkflow(t *testing.T) {
	t.Run("success - workflow closes on timer", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockBillRepo := mocks.NewMockBillRepository(ctrl)
		mockLineItemRepo := mocks.NewMockLineItemRepository(ctrl)

		activities := &BillActivities{
			BillRepo:     mockBillRepo,
			LineItemRepo: mockLineItemRepo,
		}

		testSuite := &testsuite.WorkflowTestSuite{}
		env := testSuite.NewTestWorkflowEnvironment()
		env.RegisterActivity(activities.InsertLineItem)
		env.RegisterActivity(activities.CloseBill)

		billUUID := "bill-123"
		closedAt := time.Date(2024, 1, 31, 23, 59, 59, 0, time.UTC)

		// Expect close to be called
		mockBillRepo.EXPECT().
			Close(gomock.Any(), billUUID, gomock.Any()).
			Return(nil)

		mockBillRepo.EXPECT().
			FetchClosed(gomock.Any(), billUUID, gomock.Any()).
			Return(int64(0), closedAt, nil)

		// Start workflow with period end in the past
		input := BillWorkflowInput{
			BillUUID:  billUUID,
			PeriodEnd: time.Now().Add(-time.Hour), // Already expired
		}

		env.ExecuteWorkflow(BillWorkflow, input)

		require.True(t, env.IsWorkflowCompleted())
		require.NoError(t, env.GetWorkflowError())

		var result BillWorkflowResult
		require.NoError(t, env.GetWorkflowResult(&result))

		assert.Equal(t, billUUID, result.BillUUID)
		assert.Equal(t, int64(0), result.TotalCents)
		assert.Equal(t, 0, result.ItemCount)
	})

	t.Run("success - workflow processes line item and closes", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockBillRepo := mocks.NewMockBillRepository(ctrl)
		mockLineItemRepo := mocks.NewMockLineItemRepository(ctrl)

		activities := &BillActivities{
			BillRepo:     mockBillRepo,
			LineItemRepo: mockLineItemRepo,
		}

		testSuite := &testsuite.WorkflowTestSuite{}
		env := testSuite.NewTestWorkflowEnvironment()
		env.RegisterActivity(activities.InsertLineItem)
		env.RegisterActivity(activities.CloseBill)

		billUUID := "bill-123"
		closedAt := time.Date(2024, 1, 31, 23, 59, 59, 0, time.UTC)

		// Expect line item to be inserted
		mockLineItemRepo.EXPECT().
			InsertWithBillUpdate(gomock.Any(), gomock.AssignableToTypeOf(&entity.LineItemEntity{})).
			DoAndReturn(func(_ context.Context, li *entity.LineItemEntity) error {
				assert.Equal(t, "item-1", li.UUID)
				assert.Equal(t, billUUID, li.BillUUID)
				assert.Equal(t, "TRANSACTION", li.FeeType)
				assert.Equal(t, int64(1000), li.AmountCents)
				return nil
			})

		// Expect close to be called
		mockBillRepo.EXPECT().
			Close(gomock.Any(), billUUID, gomock.Any()).
			Return(nil)

		mockBillRepo.EXPECT().
			FetchClosed(gomock.Any(), billUUID, gomock.Any()).
			Return(int64(1000), closedAt, nil)

		// Send signal before workflow completes
		env.RegisterDelayedCallback(func() {
			env.SignalWorkflow(SignalAddLineItem, AddLineItemSignal{
				UUID:           "item-1",
				IdempotencyKey: "idem-1",
				FeeType:        "TRANSACTION",
				Description:    "Test transaction",
				AmountCents:    1000,
			})
		}, time.Millisecond*100)

		// Signal close after line item
		env.RegisterDelayedCallback(func() {
			env.SignalWorkflow(SignalCloseBill, nil)
		}, time.Millisecond*200)

		input := BillWorkflowInput{
			BillUUID:  billUUID,
			PeriodEnd: time.Now().Add(time.Hour * 24), // Far future
		}

		env.ExecuteWorkflow(BillWorkflow, input)

		require.True(t, env.IsWorkflowCompleted())
		require.NoError(t, env.GetWorkflowError())

		var result BillWorkflowResult
		require.NoError(t, env.GetWorkflowResult(&result))

		assert.Equal(t, billUUID, result.BillUUID)
		assert.Equal(t, int64(1000), result.TotalCents)
		assert.Equal(t, 1, result.ItemCount)
	})

	t.Run("success - workflow handles multiple line items", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockBillRepo := mocks.NewMockBillRepository(ctrl)
		mockLineItemRepo := mocks.NewMockLineItemRepository(ctrl)

		activities := &BillActivities{
			BillRepo:     mockBillRepo,
			LineItemRepo: mockLineItemRepo,
		}

		testSuite := &testsuite.WorkflowTestSuite{}
		env := testSuite.NewTestWorkflowEnvironment()
		env.RegisterActivity(activities.InsertLineItem)
		env.RegisterActivity(activities.CloseBill)

		billUUID := "bill-123"
		closedAt := time.Date(2024, 1, 31, 23, 59, 59, 0, time.UTC)

		// Expect two line items to be inserted
		mockLineItemRepo.EXPECT().
			InsertWithBillUpdate(gomock.Any(), gomock.Any()).
			Return(nil).
			Times(2)

		// Expect close to be called
		mockBillRepo.EXPECT().
			Close(gomock.Any(), billUUID, gomock.Any()).
			Return(nil)

		mockBillRepo.EXPECT().
			FetchClosed(gomock.Any(), billUUID, gomock.Any()).
			Return(int64(3000), closedAt, nil)

		// Send first line item
		env.RegisterDelayedCallback(func() {
			env.SignalWorkflow(SignalAddLineItem, AddLineItemSignal{
				UUID:           "item-1",
				IdempotencyKey: "idem-1",
				FeeType:        "TRANSACTION",
				AmountCents:    1000,
			})
		}, time.Millisecond*100)

		// Send second line item
		env.RegisterDelayedCallback(func() {
			env.SignalWorkflow(SignalAddLineItem, AddLineItemSignal{
				UUID:           "item-2",
				IdempotencyKey: "idem-2",
				FeeType:        "TRANSACTION",
				AmountCents:    2000,
			})
		}, time.Millisecond*200)

		// Signal close
		env.RegisterDelayedCallback(func() {
			env.SignalWorkflow(SignalCloseBill, nil)
		}, time.Millisecond*300)

		input := BillWorkflowInput{
			BillUUID:  billUUID,
			PeriodEnd: time.Now().Add(time.Hour * 24),
		}

		env.ExecuteWorkflow(BillWorkflow, input)

		require.True(t, env.IsWorkflowCompleted())
		require.NoError(t, env.GetWorkflowError())

		var result BillWorkflowResult
		require.NoError(t, env.GetWorkflowResult(&result))

		assert.Equal(t, 2, result.ItemCount)
		assert.Equal(t, int64(3000), result.TotalCents)
	})

	t.Run("success - workflow handles reversal", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockBillRepo := mocks.NewMockBillRepository(ctrl)
		mockLineItemRepo := mocks.NewMockLineItemRepository(ctrl)

		activities := &BillActivities{
			BillRepo:     mockBillRepo,
			LineItemRepo: mockLineItemRepo,
		}

		testSuite := &testsuite.WorkflowTestSuite{}
		env := testSuite.NewTestWorkflowEnvironment()
		env.RegisterActivity(activities.InsertLineItem)
		env.RegisterActivity(activities.CloseBill)

		billUUID := "bill-123"
		closedAt := time.Date(2024, 1, 31, 23, 59, 59, 0, time.UTC)
		originalUUID := "item-1"

		// Expect two line items to be inserted (original + reversal)
		mockLineItemRepo.EXPECT().
			InsertWithBillUpdate(gomock.Any(), gomock.Any()).
			Return(nil).
			Times(2)

		// Expect close to be called
		mockBillRepo.EXPECT().
			Close(gomock.Any(), billUUID, gomock.Any()).
			Return(nil)

		mockBillRepo.EXPECT().
			FetchClosed(gomock.Any(), billUUID, gomock.Any()).
			Return(int64(0), closedAt, nil) // Net zero

		// Send original line item
		env.RegisterDelayedCallback(func() {
			env.SignalWorkflow(SignalAddLineItem, AddLineItemSignal{
				UUID:           "item-1",
				IdempotencyKey: "idem-1",
				FeeType:        "TRANSACTION",
				AmountCents:    1000,
			})
		}, time.Millisecond*100)

		// Send reversal
		env.RegisterDelayedCallback(func() {
			env.SignalWorkflow(SignalAddLineItem, AddLineItemSignal{
				UUID:           "reversal-1",
				IdempotencyKey: "idem-reversal",
				FeeType:        "REVERSAL",
				AmountCents:    -1000,
				ReferenceUUID:  &originalUUID,
			})
		}, time.Millisecond*200)

		// Signal close
		env.RegisterDelayedCallback(func() {
			env.SignalWorkflow(SignalCloseBill, nil)
		}, time.Millisecond*300)

		input := BillWorkflowInput{
			BillUUID:  billUUID,
			PeriodEnd: time.Now().Add(time.Hour * 24),
		}

		env.ExecuteWorkflow(BillWorkflow, input)

		require.True(t, env.IsWorkflowCompleted())
		require.NoError(t, env.GetWorkflowError())

		var result BillWorkflowResult
		require.NoError(t, env.GetWorkflowResult(&result))

		assert.Equal(t, 2, result.ItemCount)
		// In-memory state will be 0, but DB will have actual calculation
		assert.Equal(t, int64(0), result.TotalCents)
	})

	t.Run("success - query returns correct state", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockBillRepo := mocks.NewMockBillRepository(ctrl)
		mockLineItemRepo := mocks.NewMockLineItemRepository(ctrl)

		activities := &BillActivities{
			BillRepo:     mockBillRepo,
			LineItemRepo: mockLineItemRepo,
		}

		testSuite := &testsuite.WorkflowTestSuite{}
		env := testSuite.NewTestWorkflowEnvironment()
		env.RegisterActivity(activities.InsertLineItem)
		env.RegisterActivity(activities.CloseBill)

		billUUID := "bill-123"
		closedAt := time.Date(2024, 1, 31, 23, 59, 59, 0, time.UTC)

		// Expect line item to be inserted
		mockLineItemRepo.EXPECT().
			InsertWithBillUpdate(gomock.Any(), gomock.Any()).
			Return(nil)

		// Expect close to be called
		mockBillRepo.EXPECT().
			Close(gomock.Any(), billUUID, gomock.Any()).
			Return(nil)

		mockBillRepo.EXPECT().
			FetchClosed(gomock.Any(), billUUID, gomock.Any()).
			Return(int64(1000), closedAt, nil)

		// Query state after line item added
		env.RegisterDelayedCallback(func() {
			env.SignalWorkflow(SignalAddLineItem, AddLineItemSignal{
				UUID:           "item-1",
				IdempotencyKey: "idem-1",
				FeeType:        "TRANSACTION",
				AmountCents:    1000,
			})
		}, time.Millisecond*100)

		// Query state (must be done before workflow completes)
		env.RegisterDelayedCallback(func() {
			result, err := env.QueryWorkflow(QueryGetBillState)
			require.NoError(t, err)

			var state BillStateQuery
			require.NoError(t, result.Get(&state))

			assert.Equal(t, "OPEN", state.Status)
			assert.Equal(t, int64(1000), state.TotalCents)
			assert.Equal(t, 1, state.ItemCount)

			// Now close
			env.SignalWorkflow(SignalCloseBill, nil)
		}, time.Millisecond*200)

		input := BillWorkflowInput{
			BillUUID:  billUUID,
			PeriodEnd: time.Now().Add(time.Hour * 24),
		}

		env.ExecuteWorkflow(BillWorkflow, input)

		require.True(t, env.IsWorkflowCompleted())
		require.NoError(t, env.GetWorkflowError())
	})
}

func TestBillActivities(t *testing.T) {
	t.Run("InsertLineItem - success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockBillRepo := mocks.NewMockBillRepository(ctrl)
		mockLineItemRepo := mocks.NewMockLineItemRepository(ctrl)

		activities := &BillActivities{
			BillRepo:     mockBillRepo,
			LineItemRepo: mockLineItemRepo,
		}

		mockLineItemRepo.EXPECT().
			InsertWithBillUpdate(gomock.Any(), gomock.AssignableToTypeOf(&entity.LineItemEntity{})).
			DoAndReturn(func(_ context.Context, li *entity.LineItemEntity) error {
				assert.Equal(t, "item-123", li.UUID)
				assert.Equal(t, "bill-123", li.BillUUID)
				assert.Equal(t, "TRANSACTION", li.FeeType)
				assert.Equal(t, int64(1000), li.AmountCents)
				return nil
			})

		result, err := activities.InsertLineItem(context.Background(), InsertLineItemInput{
			UUID:           "item-123",
			BillUUID:       "bill-123",
			IdempotencyKey: "idem-123",
			FeeType:        "TRANSACTION",
			Description:    "Test transaction",
			AmountCents:    1000,
		})

		require.NoError(t, err)
		assert.Equal(t, "item-123", result.UUID)
	})

	t.Run("InsertLineItem - error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockBillRepo := mocks.NewMockBillRepository(ctrl)
		mockLineItemRepo := mocks.NewMockLineItemRepository(ctrl)

		activities := &BillActivities{
			BillRepo:     mockBillRepo,
			LineItemRepo: mockLineItemRepo,
		}

		mockLineItemRepo.EXPECT().
			InsertWithBillUpdate(gomock.Any(), gomock.Any()).
			Return(assert.AnError)

		result, err := activities.InsertLineItem(context.Background(), InsertLineItemInput{
			UUID:        "item-123",
			BillUUID:    "bill-123",
			AmountCents: 1000,
		})

		assert.Nil(t, result)
		assert.Error(t, err)
	})

	t.Run("CloseBill - success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockBillRepo := mocks.NewMockBillRepository(ctrl)
		mockLineItemRepo := mocks.NewMockLineItemRepository(ctrl)

		activities := &BillActivities{
			BillRepo:     mockBillRepo,
			LineItemRepo: mockLineItemRepo,
		}

		closedAt := time.Date(2024, 1, 31, 23, 59, 59, 0, time.UTC)

		mockBillRepo.EXPECT().
			Close(gomock.Any(), "bill-123", gomock.Any()).
			Return(nil)

		mockBillRepo.EXPECT().
			FetchClosed(gomock.Any(), "bill-123", gomock.Any()).
			Return(int64(5000), closedAt, nil)

		result, err := activities.CloseBill(context.Background(), CloseBillInput{
			BillUUID: "bill-123",
		})

		require.NoError(t, err)
		assert.Equal(t, int64(5000), result.TotalCents)
		assert.Equal(t, closedAt, result.ClosedAt)
	})

	t.Run("CloseBill - close error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockBillRepo := mocks.NewMockBillRepository(ctrl)
		mockLineItemRepo := mocks.NewMockLineItemRepository(ctrl)

		activities := &BillActivities{
			BillRepo:     mockBillRepo,
			LineItemRepo: mockLineItemRepo,
		}

		mockBillRepo.EXPECT().
			Close(gomock.Any(), "bill-123", gomock.Any()).
			Return(assert.AnError)

		result, err := activities.CloseBill(context.Background(), CloseBillInput{
			BillUUID: "bill-123",
		})

		assert.Nil(t, result)
		assert.Error(t, err)
	})

	t.Run("CloseBill - fetch closed error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockBillRepo := mocks.NewMockBillRepository(ctrl)
		mockLineItemRepo := mocks.NewMockLineItemRepository(ctrl)

		activities := &BillActivities{
			BillRepo:     mockBillRepo,
			LineItemRepo: mockLineItemRepo,
		}

		mockBillRepo.EXPECT().
			Close(gomock.Any(), "bill-123", gomock.Any()).
			Return(nil)

		mockBillRepo.EXPECT().
			FetchClosed(gomock.Any(), "bill-123", gomock.Any()).
			Return(int64(0), time.Time{}, assert.AnError)

		result, err := activities.CloseBill(context.Background(), CloseBillInput{
			BillUUID: "bill-123",
		})

		assert.Nil(t, result)
		assert.Error(t, err)
	})
}
