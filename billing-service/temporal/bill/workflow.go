package bill

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

func defaultActivityOptions() workflow.ActivityOptions {
	return workflow.ActivityOptions{
		StartToCloseTimeout: 30 * time.Second,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0,
			MaximumInterval:    30 * time.Second,
			MaximumAttempts:    5,
		},
	}
}

type billWorkflow struct {
	state billWorkflowState
	input BillWorkflowInput

	addItemChan workflow.ReceiveChannel
	closeChan   workflow.ReceiveChannel

	closed      bool
	timerCancel workflow.CancelFunc
}

func newBillWorkflow(ctx workflow.Context, input BillWorkflowInput) *billWorkflow {
	return &billWorkflow{
		state: billWorkflowState{
			Status: "OPEN",
		},
		input:       input,
		addItemChan: workflow.GetSignalChannel(ctx, SignalAddLineItem),
		closeChan:   workflow.GetSignalChannel(ctx, SignalCloseBill),
	}
}

func BillWorkflow(ctx workflow.Context, input BillWorkflowInput) (*BillWorkflowResult, error) {
	w := newBillWorkflow(ctx, input)
	return w.run(ctx)
}

func (w *billWorkflow) run(ctx workflow.Context) (*BillWorkflowResult, error) {
	if err := w.registerQueryHandlers(ctx); err != nil {
		return nil, err
	}

	timerFuture := w.startTimer(ctx)
	w.eventLoop(ctx, timerFuture)
	return w.closeBill(ctx)
}

func (w *billWorkflow) registerQueryHandlers(ctx workflow.Context) error {
	return workflow.SetQueryHandler(ctx, QueryGetBillState, func() (*BillStateQuery, error) {
		return &BillStateQuery{
			Status:     w.state.Status,
			TotalCents: w.state.TotalCents,
			ItemCount:  w.state.ItemCount,
		}, nil
	})
}

func (w *billWorkflow) startTimer(ctx workflow.Context) workflow.Future {
	timerDuration := w.input.PeriodEnd.Sub(workflow.Now(ctx))
	if timerDuration < 0 {
		timerDuration = 0
	}

	timerCtx, cancel := workflow.WithCancel(ctx)
	w.timerCancel = cancel

	return workflow.NewTimer(timerCtx, timerDuration)
}

// drainPendingSignals processes any buffered signals before workflow completion.
// This ensures no line items are lost if they arrived just before the close event.
func (w *billWorkflow) drainPendingSignals(ctx workflow.Context) {
	for {
		var signal AddLineItemSignal
		if !w.addItemChan.ReceiveAsync(&signal) {
			break
		}
		w.processLineItem(ctx, signal)
	}
}
