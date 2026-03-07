package handlers

import (
	"context"
	"errors"
	"log/slog"

	"encore.app/db/repository"
	"encore.app/dto"
	"encore.app/entity"
	"encore.app/utils"
	"encore.dev/storage/sqldb"
	"github.com/google/uuid"
)

type CreateCustomerHandler struct {
	CustomerRepo repository.CustomerRepository
}

func (h *CreateCustomerHandler) Handle(ctx context.Context, req *dto.CreateCustomerRequest) (*dto.CreateCustomerResponse, error) {
	if validationErrs := validateCreateCustomer(req); len(validationErrs) != 0 {
		return nil, utils.ErrValidationFailedWithDetails(validationErrs)
	}

	customer, err := h.CustomerRepo.FetchByEmail(ctx, req.Email)
	if err != nil && !errors.Is(err, sqldb.ErrNoRows) {
		slog.ErrorContext(ctx, "error fetch customer",
			"email", req.Email,
			"err", err.Error())

		return nil, utils.ErrInternal
	}

	if customer != nil {
		return nil, utils.ErrEmailAlreadyUsed
	}

	cust := &entity.CustomerEntity{
		UUID:  uuid.New().String(),
		Name:  req.Name,
		Email: req.Email,
	}

	insertErr := h.CustomerRepo.Insert(ctx, cust)
	if insertErr != nil {
		return nil, insertErr
	}

	return &dto.CreateCustomerResponse{
		UUID:  cust.UUID,
		Name:  req.Name,
		Email: req.Email,
	}, nil
}

func validateCreateCustomer(req *dto.CreateCustomerRequest) []utils.ValidationError {
	var errs []utils.ValidationError
	if req.Name == "" {
		errs = append(errs, utils.ErrInvalidName)
	}
	if req.Email == "" {
		errs = append(errs, utils.ErrInvalidEmail)
	}
	return errs
}
