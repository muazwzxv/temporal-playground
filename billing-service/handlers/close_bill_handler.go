package handlers

import (
	"context"
	"errors"

	"encore.app/db/repository"
	"encore.app/dto"
	t "encore.app/temporal"
	tbill "encore.app/temporal/bill"
	"encore.app/utils"

	"encore.dev/storage/sqldb"
	"go.temporal.io/api/serviceerror"
)

type CloseBillHandler struct {
	BillRepo       repository.BillRepository
	TemporalClient t.WorkflowClient
}

func (h *CloseBillHandler) Handle(ctx context.Context, req *dto.CloseBillRequest) (*dto.CloseBillResponse, error) {
	if req.UUID == "" {
		return nil, utils.ErrUUIDMissing
	}

	bill, err := h.BillRepo.FetchByUUID(ctx, req.UUID)
	if err != nil {
		if errors.Is(err, sqldb.ErrNoRows) {
			return nil, utils.ErrBillNotFoundAPI
		}
		return nil, utils.ErrInternal
	}

	if !bill.IsOpen() {
		return nil, utils.ErrBillAlreadyClosedAPI
	}

	workflowID := t.BillWorkflowIDPrefix + req.UUID
	err = h.TemporalClient.SignalWorkflow(ctx, workflowID, "", tbill.SignalCloseBill, nil)

	if err != nil {
		var notFound *serviceerror.NotFound
		if errors.As(err, &notFound) {
			return nil, utils.ErrBillAlreadyClosedAPI
		}
		return nil, utils.ErrWorkflowSignalFailed
	}

	return &dto.CloseBillResponse{
		UUID:    req.UUID,
		Status:  "CLOSING",
		Message: "Bill close initiated. Poll GET /v1/bill/get for final state.",
	}, nil
}
