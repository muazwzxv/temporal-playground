package handlers

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"encore.app/db/repository"
	"encore.app/dto"
	"encore.app/utils"
	"encore.dev/storage/sqldb"
)

type GetBillHandler struct {
	BillRepo repository.BillRepository
}

func (h *GetBillHandler) Handle(ctx context.Context, req *dto.GetBillRequest) (*dto.GetBillResponse, error) {
	if req.UUID == "" {
		return nil, utils.ErrUUIDMissing
	}

	bill, err := h.BillRepo.FetchByUUID(ctx, req.UUID)
	if err != nil {
		slog.ErrorContext(ctx, "error fetching bill",
			"uuid", req.UUID,
			"err", err)

		if errors.Is(err, sqldb.ErrNoRows) {
			return nil, utils.ErrNotFound
		}

		return nil, utils.ErrInternal
	}

	response := &dto.GetBillResponse{
		UUID:         bill.UUID,
		CustomerUUID: bill.CustomerUUID,
		Status:       bill.Status,
		Currency:     bill.Currency,
		TotalCents:   0,
		PeriodStart:  bill.PeriodStart.Format(time.RFC3339),
		PeriodEnd:    bill.PeriodEnd.Format(time.RFC3339),
		CreatedAt:    bill.CreatedAt.Format(time.RFC3339),
		UpdatedAt:    bill.UpdatedAt.Format(time.RFC3339),
	}

	if bill.TotalCents != nil {
		response.TotalCents = *bill.TotalCents
	}

	if bill.ClosedAt != nil {
		response.ClosedAt = bill.ClosedAt.Format(time.RFC3339)
	}

	return response, nil
}
