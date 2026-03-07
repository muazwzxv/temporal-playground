package db

import (
	"context"
	"log/slog"
	"time"

	"encore.app/entity"
	"encore.dev/storage/sqldb"
)

func FetchLineItemByBillAndKey(ctx context.Context, db *sqldb.Database, billUUID string, idempotencyKey string) (*entity.LineItemEntity, error) {
	query := `
		SELECT
			uuid, bill_uuid, idempotency_key, fee_type,
			description, amount_cents, reference_uuid, created_at
		FROM line_items
			WHERE bill_uuid = $1 AND idempotency_key = $2
	`
	li := &entity.LineItemEntity{}

	err := db.QueryRow(ctx, query, billUUID, idempotencyKey).
		Scan(&li.UUID, &li.BillUUID, &li.IdempotencyKey, &li.FeeType, &li.Description, &li.AmountCents, &li.ReferenceUUID, &li.CreatedAt)
	if err != nil {
		return nil, err
	}
	return li, nil
}

func UpsertLineItem(ctx context.Context, db *sqldb.Database, lineItem *entity.LineItemEntity) (string, error) {
	var uuid string
	err := db.QueryRow(ctx, `
		INSERT INTO line_items
			(uuid, bill_uuid, idempotency_key, fee_type, description, amount_cents, reference_uuid)
		VALUES
			($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT
			(bill_uuid, idempotency_key) DO NOTHING
		RETURNING uuid
	`, lineItem.UUID, lineItem.BillUUID, lineItem.IdempotencyKey, lineItem.FeeType, lineItem.Description, lineItem.AmountCents, lineItem.ReferenceUUID).Scan(&uuid)

	if err != nil {
		// ON CONFLICT DO NOTHING returns no row — fetch existing
		err = db.QueryRow(ctx, `
			SELECT
				uuid
			FROM line_items
				WHERE bill_uuid = $1 AND idempotency_key = $2
		`, lineItem.BillUUID, lineItem.IdempotencyKey).Scan(&uuid)
		if err != nil {
			return "", err
		}
	}

	return uuid, nil
}

func InsertLineItem(ctx context.Context, db *sqldb.Database, lineItem *entity.LineItemEntity) error {
	_, insertErr := db.Exec(ctx, `
		INSERT INTO line_items
			(uuid, bill_uuid, idempotency_key, fee_type, description, amount_cents)
		VALUES
			($1, $2, $3, $4, $5, $6)
	`, lineItem.UUID, lineItem.BillUUID, lineItem.IdempotencyKey, lineItem.FeeType, lineItem.Description, lineItem.AmountCents)
	if insertErr != nil {
		slog.ErrorContext(ctx, "error inserting line item",
			"uuid", lineItem.UUID,
			"err", insertErr.Error)

		return insertErr
	}
	return nil
}

// InsertLineItemWithBillUpdate inserts a line item and atomically updates the bill's total_cents
// within a single transaction to maintain consistency. Uses ON CONFLICT DO NOTHING for idempotency.
func InsertLineItemWithBillUpdate(ctx context.Context, db *sqldb.Database, lineItem *entity.LineItemEntity) error {
	tx, err := db.Begin(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "error beginning transaction",
			"bill_uuid", lineItem.BillUUID,
			"err", err.Error())
		return err
	}
	defer tx.Rollback()

	// Insert the line item with ON CONFLICT DO NOTHING for idempotency
	result, err := tx.Exec(ctx, `
		INSERT INTO line_items
			(uuid, bill_uuid, idempotency_key, fee_type, description, amount_cents)
		VALUES
			($1, $2, $3, $4, $5, $6)
		ON CONFLICT (bill_uuid, idempotency_key) DO NOTHING
	`, lineItem.UUID, lineItem.BillUUID, lineItem.IdempotencyKey, lineItem.FeeType, lineItem.Description, lineItem.AmountCents)
	if err != nil {
		slog.ErrorContext(ctx, "error inserting line item in transaction",
			"uuid", lineItem.UUID,
			"err", err.Error())
		return err
	}

	// Only update bill total if the row was actually inserted (not a duplicate)
	rowsAffected := result.RowsAffected()

	if rowsAffected > 0 {
		// Update the bill's total_cents
		_, err = tx.Exec(ctx, `
			UPDATE bills
			SET total_cents = COALESCE(total_cents, 0) + $2,
			    updated_at = NOW()
			WHERE uuid = $1
		`, lineItem.BillUUID, lineItem.AmountCents)
		if err != nil {
			slog.ErrorContext(ctx, "error updating bill total_cents",
				"bill_uuid", lineItem.BillUUID,
				"err", err.Error())
			return err
		}
	}

	if err = tx.Commit(); err != nil {
		slog.ErrorContext(ctx, "error committing transaction",
			"bill_uuid", lineItem.BillUUID,
			"err", err.Error())
		return err
	}

	return nil
}

// FetchLineItemsByBillUUID fetches line items for a bill with cursor-based pagination.
// Uses (created_at, id) tuple for stable cursor-based pagination, matching the bills API convention.
func FetchLineItemsByBillUUID(ctx context.Context, db *sqldb.Database, billUUID string, cursorTime time.Time, cursorID int64, limit int) ([]*entity.LineItemEntity, error) {
	var query string
	var args []any

	subsequentPage := cursorID > 0

	if subsequentPage {
		query = `
			SELECT
				id, uuid, bill_uuid, idempotency_key, fee_type, description, amount_cents, reference_uuid, created_at
			FROM line_items
			WHERE bill_uuid = $1
				AND (created_at, id) > ($2, $3)
			ORDER BY created_at ASC, id ASC
			LIMIT $4
		`
		args = []any{billUUID, cursorTime, cursorID, limit}
	} else {
		// first page
		query = `
			SELECT
				id, uuid, bill_uuid, idempotency_key, fee_type, description, amount_cents, reference_uuid, created_at
			FROM line_items
			WHERE bill_uuid = $1
			ORDER BY created_at ASC, id ASC
			LIMIT $2
		`
		args = []any{billUUID, limit}
	}

	rows, err := db.Query(ctx, query, args...)
	if err != nil {
		slog.ErrorContext(ctx, "error fetching line items", "bill_uuid", billUUID, "err", err.Error())
		return nil, err
	}
	defer rows.Close()

	var lineItems []*entity.LineItemEntity
	for rows.Next() {
		li := &entity.LineItemEntity{}
		err := rows.Scan(&li.ID, &li.UUID, &li.BillUUID, &li.IdempotencyKey, &li.FeeType, &li.Description, &li.AmountCents, &li.ReferenceUUID, &li.CreatedAt)
		if err != nil {
			slog.ErrorContext(ctx, "error scanning line item row", "err", err.Error())
			return nil, err
		}
		lineItems = append(lineItems, li)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return lineItems, nil
}

// FetchLineItemByUUID fetches a single line item by UUID
func FetchLineItemByUUID(ctx context.Context, db *sqldb.Database, uuid string) (*entity.LineItemEntity, error) {
	query := `
		SELECT
			uuid, bill_uuid, idempotency_key, fee_type, description, amount_cents, reference_uuid, created_at
		FROM line_items
		WHERE uuid = $1
	`
	li := &entity.LineItemEntity{}

	err := db.QueryRow(ctx, query, uuid).
		Scan(&li.UUID, &li.BillUUID, &li.IdempotencyKey, &li.FeeType, &li.Description, &li.AmountCents, &li.ReferenceUUID, &li.CreatedAt)
	if err != nil {
		return nil, err
	}
	return li, nil
}

// FetchReversalByOriginalUUID checks if a line item has been reversed.
// Returns the reversal line item if found, or sqldb.ErrNoRows if not reversed.
func FetchReversalByOriginalUUID(ctx context.Context, db *sqldb.Database, originalUUID string) (*entity.LineItemEntity, error) {
	query := `
		SELECT
			uuid, bill_uuid, idempotency_key, fee_type, description, amount_cents, reference_uuid, created_at
		FROM line_items
		WHERE reference_uuid = $1
		LIMIT 1
	`
	li := &entity.LineItemEntity{}

	err := db.QueryRow(ctx, query, originalUUID).
		Scan(&li.UUID, &li.BillUUID, &li.IdempotencyKey, &li.FeeType, &li.Description, &li.AmountCents, &li.ReferenceUUID, &li.CreatedAt)
	if err != nil {
		return nil, err
	}
	return li, nil
}
