package repository

import (
	"context"
	"time"

	"encore.app/db"
	"encore.app/entity"
	"encore.dev/storage/sqldb"
)

// BillRepo is the PostgreSQL implementation of BillRepository.
type BillRepo struct {
	DB *sqldb.Database
}

// Ensure BillRepo implements BillRepository.
var _ BillRepository = (*BillRepo)(nil)

func (r *BillRepo) FetchByUUID(ctx context.Context, uuid string) (*entity.BillEntity, error) {
	return db.FetchBillByUUID(ctx, r.DB, uuid)
}

func (r *BillRepo) Insert(ctx context.Context, bill *entity.BillEntity) error {
	return db.InsertBill(ctx, r.DB, bill)
}

func (r *BillRepo) Close(ctx context.Context, billUUID string, closedAt time.Time) error {
	return db.CloseBill(ctx, r.DB, billUUID, closedAt)
}

func (r *BillRepo) FetchClosed(ctx context.Context, billUUID string, fallbackClosedAt time.Time) (int64, time.Time, error) {
	return db.FetchClosedBill(ctx, r.DB, billUUID, fallbackClosedAt)
}

func (r *BillRepo) FetchAll(ctx context.Context, params db.BillQueryParams) ([]*entity.BillEntity, error) {
	return db.FetchBills(ctx, r.DB, params)
}
