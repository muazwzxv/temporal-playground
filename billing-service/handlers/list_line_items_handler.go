package handlers

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"encore.app/db/repository"
	"encore.app/dto"
	"encore.app/entity"
	"encore.app/utils"

	"encore.dev/storage/sqldb"
)

const (
	defaultLineItemLimit = 20
	maxLineItemLimit     = 20
)

type ListLineItemsHandler struct {
	BillRepo     repository.BillRepository
	LineItemRepo repository.LineItemRepository
}

func (h *ListLineItemsHandler) Handle(ctx context.Context, req *dto.ListLineItemsRequest) (*dto.ListLineItemsResponse, error) {
	if req.BillUUID == "" {
		return nil, utils.ErrUUIDMissing
	}

	limit := req.Limit
	if limit <= 0 || limit > maxLineItemLimit {
		limit = defaultLineItemLimit
	}

	// Decode cursor (same convention as list bills API)
	cursorTime, cursorID, err := utils.DecodeCursor(req.Cursor)
	if err != nil {
		slog.ErrorContext(ctx, "invalid cursor", "cursor", req.Cursor, "err", err)
		return nil, utils.ErrInvalidCursor
	}

	bill, err := h.BillRepo.FetchByUUID(ctx, req.BillUUID)
	if err != nil {
		if errors.Is(err, sqldb.ErrNoRows) {
			return nil, utils.ErrBillNotFoundAPI
		}
		slog.ErrorContext(ctx, "error fetching bill", "bill_uuid", req.BillUUID, "err", err)
		return nil, utils.ErrInternal
	}

	// Fetch line items (limit+1 to determine has_more)
	lineItems, err := h.LineItemRepo.FetchByBillUUID(ctx, req.BillUUID, cursorTime, cursorID, limit+1)
	if err != nil {
		slog.ErrorContext(ctx, "error fetching line items", "bill_uuid", req.BillUUID, "err", err)
		return nil, utils.ErrInternal
	}

	hasMore := len(lineItems) > limit
	if hasMore {
		lineItems = lineItems[:limit]
	}

	// Encode next cursor using (created_at, id) like bills API
	var nextCursor string
	if hasMore && len(lineItems) > 0 {
		last := lineItems[len(lineItems)-1]
		nextCursor = utils.EncodeCursor(last.CreatedAt, last.ID)
	}

	data := make([]dto.LineItemSummary, len(lineItems))
	for i, li := range lineItems {
		data[i] = mapLineItemToSummary(li, bill.Currency)
	}

	return &dto.ListLineItemsResponse{
		Data: data,
		Pagination: dto.PaginationResponse{
			NextCursor: nextCursor,
			HasMore:    hasMore,
		},
	}, nil
}

func mapLineItemToSummary(li *entity.LineItemEntity, currency string) dto.LineItemSummary {
	summary := dto.LineItemSummary{
		UUID:        li.UUID,
		FeeType:     li.FeeType,
		Description: li.Description,
		Amount: dto.Money{
			Amount:   li.AmountCents,
			Currency: currency,
		},
		CreatedAt: li.CreatedAt.Format(time.RFC3339),
	}
	if li.ReferenceUUID != nil {
		summary.ReferenceUUID = *li.ReferenceUUID
	}
	return summary
}
