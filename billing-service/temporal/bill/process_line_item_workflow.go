package bill

import "go.temporal.io/sdk/workflow"

func (w *billWorkflow) processLineItem(ctx workflow.Context, signal AddLineItemSignal) {
	logger := workflow.GetLogger(ctx)
	activityCtx := workflow.WithActivityOptions(ctx, defaultActivityOptions())

	var result InsertLineItemResult
	err := workflow.ExecuteActivity(activityCtx, (*BillActivities).InsertLineItem, InsertLineItemInput{
		UUID:           signal.UUID,
		BillUUID:       w.input.BillUUID,
		IdempotencyKey: signal.IdempotencyKey,
		FeeType:        signal.FeeType,
		Description:    signal.Description,
		AmountCents:    signal.AmountCents,
		ReferenceUUID:  signal.ReferenceUUID,
	}).Get(ctx, &result)

	if err != nil {
		// Log but continue - retry policy exhausted, bill will still close
		logger.Error("failed to insert line item", "error", err, "uuid", signal.UUID)
	}

	// Update in-memory counters regardless (for query handler accuracy)
	w.state.TotalCents += signal.AmountCents
	w.state.ItemCount++
}
