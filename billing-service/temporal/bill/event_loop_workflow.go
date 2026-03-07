package bill

import "go.temporal.io/sdk/workflow"

func (w *billWorkflow) eventLoop(ctx workflow.Context, timerFuture workflow.Future) {
	for !w.closed {
		selector := workflow.NewSelector(ctx)

		// handles adding line item
		selector.AddReceive(w.addItemChan, func(c workflow.ReceiveChannel, more bool) {
			var signal AddLineItemSignal
			c.Receive(ctx, &signal)
			w.processLineItem(ctx, signal)
		})

		// handles manual close signal
		selector.AddReceive(w.closeChan, func(c workflow.ReceiveChannel, more bool) {
			c.Receive(ctx, nil)
			// receiving signal sets closed as true and breaks the event loop
			w.closed = true
		})

		// handles timer expiration, bill closing on configured day
		selector.AddFuture(timerFuture, func(f workflow.Future) {
			_ = f.Get(ctx, nil)
			w.closed = true
		})

		selector.Select(ctx)
	}
}
