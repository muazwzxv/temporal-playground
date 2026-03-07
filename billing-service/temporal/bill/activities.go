package bill

import (
	"context"
	"time"

	"encore.app/db/repository"
	"encore.app/entity"
)

type BillActivities struct {
	BillRepo     repository.BillRepository
	LineItemRepo repository.LineItemRepository
}

func (a *BillActivities) InsertLineItem(ctx context.Context, input InsertLineItemInput) (*InsertLineItemResult, error) {
	err := a.LineItemRepo.InsertWithBillUpdate(ctx, &entity.LineItemEntity{
		UUID:           input.UUID,
		BillUUID:       input.BillUUID,
		IdempotencyKey: input.IdempotencyKey,
		FeeType:        input.FeeType,
		Description:    input.Description,
		AmountCents:    input.AmountCents,
		ReferenceUUID:  input.ReferenceUUID,
	})
	if err != nil {
		return nil, err
	}
	return &InsertLineItemResult{UUID: input.UUID}, nil
}

func (a *BillActivities) CloseBill(ctx context.Context, input CloseBillInput) (*CloseBillResult, error) {
	now := time.Now().UTC()

	if err := a.BillRepo.Close(ctx, input.BillUUID, now); err != nil {
		return nil, err
	}

	totalCents, closedAt, err := a.BillRepo.FetchClosed(ctx, input.BillUUID, now)
	if err != nil {
		return nil, err
	}

	return &CloseBillResult{
		TotalCents: totalCents,
		ClosedAt:   closedAt,
	}, nil
}
