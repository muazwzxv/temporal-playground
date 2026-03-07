package handlers

import (
	"context"
	"errors"
	"log/slog"

	"encore.app/db/repository"
	"encore.app/dto"
	"encore.app/utils"
	"encore.dev/storage/sqldb"
)

type GetCustomerHandler struct {
	CustomerRepo repository.CustomerRepository
}

func (h *GetCustomerHandler) Handle(ctx context.Context, req *dto.GetCustomerRequest) (*dto.GetCustomerResponse, error) {
	if req.UUID == "" {
		return nil, utils.ErrUUIDMissing
	}

	customer, err := h.CustomerRepo.FetchByUUID(ctx, req.UUID)
	if err != nil {
		slog.ErrorContext(ctx, "error fetching customer",
			"uuid", req.UUID,
			"err", err)

		if errors.Is(err, sqldb.ErrNoRows) {
			return nil, utils.ErrNotFound
		}

		return nil, utils.ErrInternal
	}

	return &dto.GetCustomerResponse{
		UUID:      customer.UUID,
		Name:      customer.Name,
		Email:     customer.Email,
		CreatedAt: customer.CreatedAt,
	}, nil
}
