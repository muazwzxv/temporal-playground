package db

import (
	"context"
	"log/slog"

	"encore.app/entity"
	"encore.dev/storage/sqldb"
)

func InsertCustomer(ctx context.Context, db *sqldb.Database, customer *entity.CustomerEntity) error {
	_, insertErr := db.Exec(ctx, `
		INSERT INTO customers (uuid, name, email)
		VALUES ($1, $2, $3)
	`, customer.UUID, customer.Name, customer.Email)
	if insertErr != nil {
		slog.ErrorContext(ctx, "error inserting customer",
			"uuid", customer.UUID,
			"err", insertErr.Error)

		return insertErr
	}
	return nil
}

func FetchCustomerByEmail(ctx context.Context, db *sqldb.Database, email string) (*entity.CustomerEntity, error) {
	query := `
		SELECT uuid, name, email, created_at, updated_at
		FROM customers WHERE email = $1
  `
	c := &entity.CustomerEntity{}

	err := db.QueryRow(ctx, query, email).
		Scan(&c.UUID, &c.Name, &c.Email, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return c, nil
}

func FetchCustomerByUUID(ctx context.Context, db *sqldb.Database, uuid string) (*entity.CustomerEntity, error) {
	query := `
		SELECT uuid, name, email, created_at, updated_at
		FROM customers WHERE uuid = $1
  `
	c := &entity.CustomerEntity{}

	err := db.QueryRow(ctx, query, uuid).
		Scan(&c.UUID, &c.Name, &c.Email, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return c, nil
}
