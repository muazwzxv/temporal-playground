package repository

import (
	"context"
	"time"

	"encore.app/db"
	"encore.app/entity"
	"encore.dev/storage/sqldb"
)

// LineItemRepo is the PostgreSQL implementation of LineItemRepository.
type LineItemRepo struct {
	DB *sqldb.Database
}

// Ensure LineItemRepo implements LineItemRepository.
var _ LineItemRepository = (*LineItemRepo)(nil)

func (r *LineItemRepo) FetchByUUID(ctx context.Context, uuid string) (*entity.LineItemEntity, error) {
	return db.FetchLineItemByUUID(ctx, r.DB, uuid)
}

func (r *LineItemRepo) FetchByBillAndKey(ctx context.Context, billUUID, idempotencyKey string) (*entity.LineItemEntity, error) {
	return db.FetchLineItemByBillAndKey(ctx, r.DB, billUUID, idempotencyKey)
}

func (r *LineItemRepo) FetchByBillUUID(ctx context.Context, billUUID string, cursorTime time.Time, cursorID int64, limit int) ([]*entity.LineItemEntity, error) {
	return db.FetchLineItemsByBillUUID(ctx, r.DB, billUUID, cursorTime, cursorID, limit)
}

func (r *LineItemRepo) FetchReversalByOriginalUUID(ctx context.Context, originalUUID string) (*entity.LineItemEntity, error) {
	return db.FetchReversalByOriginalUUID(ctx, r.DB, originalUUID)
}

func (r *LineItemRepo) InsertWithBillUpdate(ctx context.Context, lineItem *entity.LineItemEntity) error {
	return db.InsertLineItemWithBillUpdate(ctx, r.DB, lineItem)
}
