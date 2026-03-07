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
	"github.com/google/uuid"
	"go.temporal.io/api/serviceerror"
)

type AddLineItemHandler struct {
	BillRepo       repository.BillRepository
	LineItemRepo   repository.LineItemRepository
	TemporalClient t.WorkflowClient
}

func (h *AddLineItemHandler) Handle(ctx context.Context, req *dto.AddLineItemRequest) (*dto.AddLineItemResponse, error) {
	if validationErrors := validateAddLineItem(req); len(validationErrors) != 0 {
		return nil, utils.ErrValidationFailedWithDetails(validationErrors)
	}

	bill, err := h.fetchBill(ctx, req.BillUUID)
	if err != nil {
		return nil, err
	}

	if !bill.IsOpen() {
		return nil, utils.ErrBillClosed
	}

	if req.Amount.Currency != bill.Currency {
		return nil, utils.ErrCurrencyMismatch
	}

	if existingResp, err := h.checkIdempotency(ctx, req, bill); existingResp != nil || err != nil {
		return existingResp, err
	}

	lineItemUUID := uuid.New().String()
	workflowID := t.BillWorkflowIDPrefix + req.BillUUID

	if err := h.queryWorkflowState(ctx, workflowID, req.BillUUID); err != nil {
		return nil, err
	}

	signal := h.buildSignal(lineItemUUID, req)

	if err := h.signalWorkflow(ctx, workflowID, req.BillUUID, signal); err != nil {
		return nil, err
	}

	return h.buildPendingResponse(lineItemUUID, req), nil
}

func (h *AddLineItemHandler) fetchBill(ctx context.Context, billUUID string) (*entity.BillEntity, error) {
	bill, err := h.BillRepo.FetchByUUID(ctx, billUUID)
	if err != nil {
		if errors.Is(err, sqldb.ErrNoRows) {
			return nil, utils.ErrBillNotFoundAPI
		}
		return nil, utils.ErrInternal
	}
	return bill, nil
}

func (h *AddLineItemHandler) checkIdempotency(ctx context.Context, req *dto.AddLineItemRequest, bill *entity.BillEntity) (*dto.AddLineItemResponse, error) {
	existing, err := h.LineItemRepo.FetchByBillAndKey(ctx, req.BillUUID, req.IdempotencyKey)
	if err != nil {
		if errors.Is(err, sqldb.ErrNoRows) {
			return nil, nil
		}
		return nil, utils.ErrInternal
	}

	return &dto.AddLineItemResponse{
		UUID:    existing.UUID,
		FeeType: existing.FeeType,
		Amount: dto.Money{
			Amount:   existing.AmountCents,
			Currency: bill.Currency,
		},
		Status:    "persisted",
		CreatedAt: existing.CreatedAt.Format(time.RFC3339),
	}, nil
}

func (h *AddLineItemHandler) queryWorkflowState(ctx context.Context, workflowID, billUUID string) error {
	queryResp, err := h.TemporalClient.QueryWorkflow(ctx, workflowID, "", tbill.QueryGetBillState)
	if err != nil {
		var notFound *serviceerror.NotFound
		if errors.As(err, &notFound) {
			rlog.Error("data inconsistency: bill exists but workflow not found",
				"bill_uuid", billUUID,
				"workflow_id", workflowID)
			return utils.ErrWorkflowNotFound
		}

		rlog.Error("failed to query workflow state",
			"bill_uuid", billUUID,
			"error", err)
		return utils.ErrWorkflowQueryFailed
	}

	var billState tbill.BillStateQuery
	if err := queryResp.Get(&billState); err != nil {
		rlog.Error("failed to decode workflow state",
			"bill_uuid", billUUID,
			"error", err)
		return utils.ErrWorkflowQueryFailed
	}

	if billState.Status == "CLOSED" {
		rlog.Info("signaling workflow in closing state",
			"bill_uuid", billUUID,
			"workflow_status", billState.Status)
	}

	return nil
}

func (h *AddLineItemHandler) buildSignal(lineItemUUID string, req *dto.AddLineItemRequest) tbill.AddLineItemSignal {
	return tbill.AddLineItemSignal{
		UUID:           lineItemUUID,
		IdempotencyKey: req.IdempotencyKey,
		FeeType:        req.FeeType,
		Description:    req.Description,
		AmountCents:    req.Amount.Amount,
	}
}

func (h *AddLineItemHandler) signalWorkflow(ctx context.Context, workflowID, billUUID string, signal tbill.AddLineItemSignal) error {
	err := h.TemporalClient.SignalWorkflow(ctx, workflowID, "", tbill.SignalAddLineItem, signal)
	if err != nil {
		var notFound *serviceerror.NotFound
		if errors.As(err, &notFound) {
			rlog.Warn("workflow completed between query and signal",
				"bill_uuid", billUUID)
			return utils.ErrBillClosed
		}

		rlog.Error("failed to signal workflow",
			"bill_uuid", billUUID,
			"error", err)
		return utils.ErrWorkflowSignalFailed
	}
	return nil
}

func (h *AddLineItemHandler) buildPendingResponse(lineItemUUID string, req *dto.AddLineItemRequest) *dto.AddLineItemResponse {
	return &dto.AddLineItemResponse{
		UUID:    lineItemUUID,
		FeeType: req.FeeType,
		Amount: dto.Money{
			Amount:   req.Amount.Amount,
			Currency: req.Amount.Currency,
		},
		Status:    "pending",
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}
}

func validateAddLineItem(req *dto.AddLineItemRequest) []utils.ValidationError {
	var validationErrors []utils.ValidationError

	if req.BillUUID == "" {
		validationErrors = append(validationErrors, utils.ErrInvalidBillUUID)
	}
	if req.IdempotencyKey == "" {
		validationErrors = append(validationErrors, utils.ErrInvalidIdempotencyKey)
	}
	if req.FeeType == "" {
		validationErrors = append(validationErrors, utils.ErrInvalidFeeType)
	}
	if req.Amount.Amount <= 0 {
		validationErrors = append(validationErrors, utils.ErrInvalidAmount)
	}
	if req.Amount.Currency == "" {
		validationErrors = append(validationErrors, utils.ErrInvalidCurrency)
	}

	return validationErrors
}
