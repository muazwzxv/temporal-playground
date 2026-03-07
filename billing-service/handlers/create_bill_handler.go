package handlers

import (
	"context"
	"errors"
	"time"

	"encore.app/db/repository"
	"encore.app/dto"
	"encore.app/entity"
	"encore.app/utils"

	t "encore.app/temporal"
	tbill "encore.app/temporal/bill"

	"encore.dev/rlog"
	"encore.dev/storage/sqldb"
	"go.temporal.io/api/serviceerror"
	tclient "go.temporal.io/sdk/client"
)

type CreateBillHandler struct {
	BillRepo       repository.BillRepository
	CustomerRepo   repository.CustomerRepository
	TemporalClient t.WorkflowClient
}

func (h *CreateBillHandler) Handle(ctx context.Context, req *dto.CreateBillRequest) (*dto.CreateBillResponse, error) {
	if validationErrors := validateCreateBill(req); len(validationErrors) != 0 {
		return nil, utils.ErrValidationFailedWithDetails(validationErrors)
	}

	// check if customer exists
	_, err := h.CustomerRepo.FetchByUUID(ctx, req.CustomerUUID)
	if err != nil {
		if errors.Is(err, sqldb.ErrNoRows) {
			return nil, utils.ErrCustomerNotFoundAPI
		}
		return nil, utils.ErrInternal
	}

	existing, err := h.BillRepo.FetchByUUID(ctx, req.UUID)
	if err != nil && !errors.Is(err, sqldb.ErrNoRows) {
		return nil, utils.ErrInternal
	}
	if existing != nil {
		return &dto.CreateBillResponse{
			UUID:        existing.UUID,
			Status:      existing.Status,
			Currency:    existing.Currency,
			PeriodStart: existing.PeriodStart.Format(time.RFC3339),
			PeriodEnd:   existing.PeriodEnd.Format(time.RFC3339),
		}, nil
	}

	periodStart, _ := time.Parse(time.RFC3339, req.PeriodStart)
	periodEnd, _ := time.Parse(time.RFC3339, req.PeriodEnd)

	bill := &entity.BillEntity{
		UUID:         req.UUID,
		CustomerUUID: req.CustomerUUID,
		Currency:     req.Currency,
		PeriodStart:  periodStart,
		PeriodEnd:    periodEnd,
	}

	if err := h.BillRepo.Insert(ctx, bill); err != nil {
		return nil, utils.ErrInternal
	}

	workflowOptions := tclient.StartWorkflowOptions{
		ID:        t.BillWorkflowIDPrefix + req.UUID,
		TaskQueue: t.TaskQueue,
	}
	_, err = h.TemporalClient.ExecuteWorkflow(ctx, workflowOptions, tbill.BillWorkflow, tbill.BillWorkflowInput{
		BillUUID:  req.UUID,
		PeriodEnd: periodEnd,
	})
	if err != nil {
		var alreadyStarted *serviceerror.WorkflowExecutionAlreadyStarted
		if errors.As(err, &alreadyStarted) {
			return &dto.CreateBillResponse{
				UUID:        req.UUID,
				Status:      "OPEN",
				Currency:    req.Currency,
				PeriodStart: req.PeriodStart,
				PeriodEnd:   req.PeriodEnd,
			}, nil
		}
		rlog.Error("workflow start failed",
			"workflow_id", workflowOptions.ID,
			"bill_uuid", req.UUID)
		return nil, utils.ErrWorkflowStartFailed
	}

	return &dto.CreateBillResponse{
		UUID:        req.UUID,
		Status:      "OPEN",
		Currency:    req.Currency,
		PeriodStart: req.PeriodStart,
		PeriodEnd:   req.PeriodEnd,
	}, nil
}

func validateCreateBill(req *dto.CreateBillRequest) []utils.ValidationError {
	var validationErrors []utils.ValidationError

	if req.UUID == "" {
		validationErrors = append(validationErrors, utils.ErrInvalidUUID)
	}
	if req.CustomerUUID == "" {
		validationErrors = append(validationErrors, utils.ErrInvalidCustomerUUID)
	}
	if req.Currency == "" || (req.Currency != "USD" && req.Currency != "GEL") {
		validationErrors = append(validationErrors, utils.ErrInvalidCurrency)
	}
	if req.PeriodStart == "" {
		validationErrors = append(validationErrors, utils.ErrInvalidPeriodStart)
	}
	if req.PeriodEnd == "" {
		validationErrors = append(validationErrors, utils.ErrInvalidPeriodEnd)
	}

	if req.PeriodStart != "" && req.PeriodEnd != "" {
		start, errStart := time.Parse(time.RFC3339, req.PeriodStart)
		end, errEnd := time.Parse(time.RFC3339, req.PeriodEnd)
		if errStart != nil {
			validationErrors = append(validationErrors, utils.ErrInvalidPeriodStart)
		}
		if errEnd != nil {
			validationErrors = append(validationErrors, utils.ErrInvalidPeriodEnd)
		}
		if errStart == nil && errEnd == nil && !end.After(start) {
			validationErrors = append(validationErrors, utils.ErrInvalidPeriod)
		}
	}

	return validationErrors
}
