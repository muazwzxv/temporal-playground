package repository

import (
	"context"

	"encore.app/db"
	"encore.app/entity"
	"encore.dev/storage/sqldb"
)

// CustomerRepo is the PostgreSQL implementation of CustomerRepository.
type CustomerRepo struct {
	DB *sqldb.Database
}

// Ensure CustomerRepo implements CustomerRepository.
var _ CustomerRepository = (*CustomerRepo)(nil)

func (r *CustomerRepo) FetchByUUID(ctx context.Context, uuid string) (*entity.CustomerEntity, error) {
	return db.FetchCustomerByUUID(ctx, r.DB, uuid)
}

func (r *CustomerRepo) FetchByEmail(ctx context.Context, email string) (*entity.CustomerEntity, error) {
	return db.FetchCustomerByEmail(ctx, r.DB, email)
}

func (r *CustomerRepo) Insert(ctx context.Context, customer *entity.CustomerEntity) error {
	return db.InsertCustomer(ctx, r.DB, customer)
}
