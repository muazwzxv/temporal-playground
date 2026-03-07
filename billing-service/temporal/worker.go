package temporal

import (
	"encore.app/db/repository"
	"encore.app/temporal/bill"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
)

func NewWorker(c client.Client, billRepo repository.BillRepository, lineItemRepo repository.LineItemRepository) worker.Worker {
	w := worker.New(c, TaskQueue, worker.Options{})

	billActivities := &bill.BillActivities{
		BillRepo:     billRepo,
		LineItemRepo: lineItemRepo,
	}
	w.RegisterActivity(billActivities)
	w.RegisterWorkflow(bill.BillWorkflow)

	return w
}
