package handlers

import (
	"context"
	"log/slog"
	"time"

	"encore.app/db"
	"encore.app/db/repository"
	"encore.app/dto"
	"encore.app/entity"
	"encore.app/utils"
)

const (
	defaultLimit = 20
	maxLimit     = 20
)

type ListBillsHandler struct {
	BillRepo repository.BillRepository
}

func (h *ListBillsHandler) Handle(ctx context.Context, req *dto.ListBillsRequest) (*dto.ListBillsResponse, error) {
	// 1. Apply default/max limit
	limit := req.Limit
	if limit <= 0 || limit > maxLimit {
		limit = defaultLimit
	}

	// 2. Determine sort order (default: desc)
	sortDesc := req.SortOrder != dto.SortOrderAsc

	// 3. Decode cursor
	cursorTime, cursorID, err := utils.DecodeCursor(req.Cursor)
	if err != nil {
		slog.ErrorContext(ctx, "invalid cursor", "cursor", req.Cursor, "err", err)
		return nil, utils.ErrInvalidCursor
	}

	// 4. Fetch bills from DB (fetch limit+1 to determine has_more)
	bills, err := h.BillRepo.FetchAll(ctx, db.BillQueryParams{
		CustomerUUID: req.CustomerUUID,
		Status:       req.Status,
		CursorTime:   cursorTime,
		CursorID:     cursorID,
		Limit:        limit + 1,
		SortDesc:     sortDesc,
	})
	if err != nil {
		slog.ErrorContext(ctx, "error fetching bills",
			"customer_uuid", req.CustomerUUID,
			"status", req.Status,
			"cursor", req.Cursor,
			"err", err)
		return nil, utils.ErrInternal
	}

	// 5. Determine pagination
	hasMore := len(bills) > limit
	if hasMore {
		bills = bills[:limit]
	}

	var nextCursor string
	if hasMore && len(bills) > 0 {
		last := bills[len(bills)-1]
		nextCursor = utils.EncodeCursor(last.CreatedAt, last.ID)
	}

	// 6. Map entities to DTOs
	data := make([]dto.BillSummary, len(bills))
	for i, bill := range bills {
		data[i] = mapBillToSummary(bill)
	}

	// 7. Return response
	return &dto.ListBillsResponse{
		Data: data,
		Pagination: dto.PaginationResponse{
			NextCursor: nextCursor,
			HasMore:    hasMore,
		},
	}, nil
}

func mapBillToSummary(bill *entity.BillEntity) dto.BillSummary {
	totalCents := int64(0)
	if bill.TotalCents != nil {
		totalCents = *bill.TotalCents
	}

	return dto.BillSummary{
		UUID:         bill.UUID,
		CustomerUUID: bill.CustomerUUID,
		Status:       bill.Status,
		Currency:     bill.Currency,
		Total: dto.Money{
			Amount:   totalCents,
			Currency: bill.Currency,
		},
		PeriodStart: bill.PeriodStart.Format(time.RFC3339),
		PeriodEnd:   bill.PeriodEnd.Format(time.RFC3339),
	}
}
