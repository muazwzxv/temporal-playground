package bill

import "go.temporal.io/sdk/workflow"

func (w *billWorkflow) closeBill(ctx workflow.Context) (*BillWorkflowResult, error) {
	// manually cancel timer if the bill is closed manually
	w.timerCancel()

	// Process any remaining buffered signals
	w.drainPendingSignals(ctx)

	w.state.Status = "CLOSED"

	// creates a separate context that's not linked to parent to ensure activity ran without interruption
	// when parents context get's cancelled, AI proposed this, good to know
	disconnectedCtx, _ := workflow.NewDisconnectedContext(ctx)
	activityCtx := workflow.WithActivityOptions(disconnectedCtx, defaultActivityOptions())

	var closeResult CloseBillResult
	err := workflow.ExecuteActivity(activityCtx, (*BillActivities).CloseBill, CloseBillInput{
		BillUUID: w.input.BillUUID,
	}).Get(disconnectedCtx, &closeResult)

	if err != nil {
		return nil, err
	}

	return &BillWorkflowResult{
		BillUUID:   w.input.BillUUID,
		TotalCents: closeResult.TotalCents,
		ItemCount:  w.state.ItemCount,
		ClosedAt:   closeResult.ClosedAt,
	}, nil
}
