package repository

import (
	"context"
	"time"

	"encore.app/db"
	"encore.app/entity"
)

//go:generate mockgen -source=interfaces.go -destination=mocks/mocks.go -package=mocks

// BillRepository defines operations for bill persistence.
// All methods return raw database errors; callers are responsible for
// translating them to domain-specific errors.
type BillRepository interface {
	FetchByUUID(ctx context.Context, uuid string) (*entity.BillEntity, error)
	Insert(ctx context.Context, bill *entity.BillEntity) error
	Close(ctx context.Context, billUUID string, closedAt time.Time) error
	FetchClosed(ctx context.Context, billUUID string, fallbackClosedAt time.Time) (int64, time.Time, error)
	FetchAll(ctx context.Context, params db.BillQueryParams) ([]*entity.BillEntity, error)
}

// LineItemRepository defines operations for line item persistence.
// All methods return raw database errors; callers are responsible for
// translating them to domain-specific errors.
type LineItemRepository interface {
	FetchByUUID(ctx context.Context, uuid string) (*entity.LineItemEntity, error)
	FetchByBillAndKey(ctx context.Context, billUUID, idempotencyKey string) (*entity.LineItemEntity, error)
	FetchByBillUUID(ctx context.Context, billUUID string, cursorTime time.Time, cursorID int64, limit int) ([]*entity.LineItemEntity, error)
	FetchReversalByOriginalUUID(ctx context.Context, originalUUID string) (*entity.LineItemEntity, error)
	InsertWithBillUpdate(ctx context.Context, lineItem *entity.LineItemEntity) error
}

// CustomerRepository defines operations for customer persistence.
// All methods return raw database errors; callers are responsible for
// translating them to domain-specific errors.
type CustomerRepository interface {
	FetchByUUID(ctx context.Context, uuid string) (*entity.CustomerEntity, error)
	FetchByEmail(ctx context.Context, email string) (*entity.CustomerEntity, error)
	Insert(ctx context.Context, customer *entity.CustomerEntity) error
}
