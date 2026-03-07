package db

import (
	"context"
	"log/slog"
	"time"

	"encore.app/entity"
	"encore.dev/storage/sqldb"
)

func FetchBillByUUID(ctx context.Context, db *sqldb.Database, uuid string) (*entity.BillEntity, error) {
	query := `
		SELECT
			uuid, customer_uuid, currency, status, period_start, period_end, closed_at, total_cents, created_at, updated_at
		FROM bills
			WHERE uuid = $1
	`
	b := &entity.BillEntity{}

	err := db.QueryRow(ctx, query, uuid).
		Scan(&b.UUID, &b.CustomerUUID, &b.Currency, &b.Status, &b.PeriodStart,
			&b.PeriodEnd, &b.ClosedAt, &b.TotalCents, &b.CreatedAt, &b.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func InsertBill(ctx context.Context, db *sqldb.Database, bill *entity.BillEntity) error {
	_, insertErr := db.Exec(ctx, `
		INSERT INTO bills
			(uuid, customer_uuid, currency, period_start, period_end, total_cents)
		VALUES
			($1, $2, $3, $4, $5, 0)
	`, bill.UUID, bill.CustomerUUID, bill.Currency, bill.PeriodStart, bill.PeriodEnd)
	if insertErr != nil {
		slog.ErrorContext(ctx, "error inserting bill",
			"uuid", bill.UUID,
			"err", insertErr.Error)

		return insertErr
	}
	return nil
}

func CloseBill(ctx context.Context, db *sqldb.Database, billUUID string, closedAt time.Time) error {
	_, err := db.Exec(ctx, `
		UPDATE bills
		SET status = 'CLOSED',
		    closed_at = $2,
		    total_cents = (
		        SELECT COALESCE(SUM(amount_cents), 0)
		        FROM line_items
		        WHERE bill_uuid = $1
		    ),
		    updated_at = $2
		WHERE
			uuid = $1 AND status = 'OPEN'
	`, billUUID, closedAt)
	if err != nil {
		slog.ErrorContext(ctx, "error closing bill",
			"uuid", billUUID,
			"err", err.Error)
		return err
	}
	return nil
}

func FetchClosedBill(ctx context.Context, db *sqldb.Database, billUUID string, fallbackClosedAt time.Time) (int64, time.Time, error) {
	var totalCents int64
	var closedAt time.Time
	err := db.QueryRow(ctx, `
		SELECT
			COALESCE(total_cents, 0),
			COALESCE(closed_at, $2)
		FROM bills
			WHERE uuid = $1
	`, billUUID, fallbackClosedAt).Scan(&totalCents, &closedAt)
	if err != nil {
		return 0, time.Time{}, err
	}
	return totalCents, closedAt, nil
}

func FetchBills(ctx context.Context, db *sqldb.Database, params BillQueryParams) ([]*entity.BillEntity, error) {
	var query string
	var args []interface{}

	// I don't know what to feel about this lmao, AI handrolled this but it seems alright for now
	// ideally some query builder ability would be nice
	// handling sort order and first vs non first page queries
	if params.SortDesc {
		if params.CursorID > 0 {
			// DESC with cursor
			query = `
				SELECT id, uuid, customer_uuid, currency, status, period_start,
				       period_end, closed_at, total_cents, created_at, updated_at
				FROM bills
				WHERE (created_at, id) < ($1, $2)
				  AND ($3 = '' OR customer_uuid = $3)
				  AND ($4 = '' OR status = $4)
				ORDER BY created_at DESC, id DESC
				LIMIT $5
			`
			args = []interface{}{params.CursorTime, params.CursorID, params.CustomerUUID, params.Status, params.Limit}
		} else {
			// DESC first page
			query = `
				SELECT id, uuid, customer_uuid, currency, status, period_start,
				       period_end, closed_at, total_cents, created_at, updated_at
				FROM bills
				WHERE ($1 = '' OR customer_uuid = $1)
				  AND ($2 = '' OR status = $2)
				ORDER BY created_at DESC, id DESC
				LIMIT $3
			`
			args = []interface{}{params.CustomerUUID, params.Status, params.Limit}
		}
	} else {
		if params.CursorID > 0 {
			// ASC with cursor
			query = `
				SELECT id, uuid, customer_uuid, currency, status, period_start,
				       period_end, closed_at, total_cents, created_at, updated_at
				FROM bills
				WHERE (created_at, id) > ($1, $2)
				  AND ($3 = '' OR customer_uuid = $3)
				  AND ($4 = '' OR status = $4)
				ORDER BY created_at ASC, id ASC
				LIMIT $5
			`
			args = []interface{}{params.CursorTime, params.CursorID, params.CustomerUUID, params.Status, params.Limit}
		} else {
			// ASC first page
			query = `
				SELECT id, uuid, customer_uuid, currency, status, period_start,
				       period_end, closed_at, total_cents, created_at, updated_at
				FROM bills
				WHERE ($1 = '' OR customer_uuid = $1)
				  AND ($2 = '' OR status = $2)
				ORDER BY created_at ASC, id ASC
				LIMIT $3
			`
			args = []interface{}{params.CustomerUUID, params.Status, params.Limit}
		}
	}

	rows, err := db.Query(ctx, query, args...)
	if err != nil {
		slog.ErrorContext(ctx, "error fetching bills", "err", err.Error())
		return nil, err
	}
	defer rows.Close()

	var bills []*entity.BillEntity
	for rows.Next() {
		b := &entity.BillEntity{}
		err := rows.Scan(&b.ID, &b.UUID, &b.CustomerUUID, &b.Currency, &b.Status,
			&b.PeriodStart, &b.PeriodEnd, &b.ClosedAt, &b.TotalCents,
			&b.CreatedAt, &b.UpdatedAt)
		if err != nil {
			slog.ErrorContext(ctx, "error scanning bill row", "err", err.Error())
			return nil, err
		}
		bills = append(bills, b)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return bills, nil
}
