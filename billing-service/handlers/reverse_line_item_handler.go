package handlers

import (
	"context"
	"errors"
	"time"

	"encore.app/db/repository"
	"encore.app/dto"
	"encore.app/entity"
	t "encore.app/temporal"
	tbill "encore.app/temporal/bill"
	"encore.app/utils"

	"encore.dev/rlog"
	"encore.dev/storage/sqldb"
	"github.com/google/uuid"
	"go.temporal.io/api/serviceerror"
)

type ReverseLineItemHandler struct {
	BillRepo       repository.BillRepository
	LineItemRepo   repository.LineItemRepository
	TemporalClient t.WorkflowClient
}

func (h *ReverseLineItemHandler) Handle(ctx context.Context, req *dto.ReverseLineItemRequest) (*dto.ReverseLineItemResponse, error) {
	if validationErrors := validateReverseLineItem(req); len(validationErrors) != 0 {
		return nil, utils.ErrValidationFailedWithDetails(validationErrors)
	}

	bill, err := h.fetchBill(ctx, req.BillUUID)
	if err != nil {
		return nil, err
	}
	if !bill.IsOpen() {
		return nil, utils.ErrBillClosed
	}

	originalLineItem, err := h.fetchOriginalLineItem(ctx, req.LineItemUUID)
	if err != nil {
		return nil, err
	}

	if originalLineItem.BillUUID != req.BillUUID {
		return nil, utils.ErrLineItemNotFoundAPI
	}

	// check if original is a reversal (cannot reverse a reversal)
	if originalLineItem.FeeType == string(entity.FeeTypeReversal) {
		return nil, utils.ErrCannotReverseReversal
	}

	if err := h.checkAlreadyReversed(ctx, req.LineItemUUID); err != nil {
		return nil, err
	}

	if existingResp, err := h.checkIdempotency(ctx, req, bill); existingResp != nil || err != nil {
		return existingResp, err
	}

	reversalUUID := uuid.New().String()
	workflowID := t.BillWorkflowIDPrefix + req.BillUUID

	if err := h.queryWorkflowState(ctx, workflowID, req.BillUUID); err != nil {
		return nil, err
	}

	signal := h.buildReversalSignal(reversalUUID, req, originalLineItem)

	if err := h.signalWorkflow(ctx, workflowID, req.BillUUID, signal); err != nil {
		return nil, err
	}

	return h.buildPendingResponse(reversalUUID, req, originalLineItem, bill.Currency), nil
}

func (h *ReverseLineItemHandler) fetchBill(ctx context.Context, billUUID string) (*entity.BillEntity, error) {
	bill, err := h.BillRepo.FetchByUUID(ctx, billUUID)
	if err != nil {
		if errors.Is(err, sqldb.ErrNoRows) {
			return nil, utils.ErrBillNotFoundAPI
		}
		rlog.Error("error fetching bill", "bill_uuid", billUUID, "error", err)
		return nil, utils.ErrInternal
	}
	return bill, nil
}

func (h *ReverseLineItemHandler) fetchOriginalLineItem(ctx context.Context, lineItemUUID string) (*entity.LineItemEntity, error) {
	lineItem, err := h.LineItemRepo.FetchByUUID(ctx, lineItemUUID)
	if err != nil {
		if errors.Is(err, sqldb.ErrNoRows) {
			return nil, utils.ErrLineItemNotFoundAPI
		}
		rlog.Error("error fetching line item", "line_item_uuid", lineItemUUID, "error", err)
		return nil, utils.ErrInternal
	}
	return lineItem, nil
}

func (h *ReverseLineItemHandler) checkAlreadyReversed(ctx context.Context, originalUUID string) error {
	_, err := h.LineItemRepo.FetchReversalByOriginalUUID(ctx, originalUUID)
	if err != nil {
		if errors.Is(err, sqldb.ErrNoRows) {
			// Not reversed yet, this is expected
			return nil
		}
		rlog.Error("error checking if line item is reversed", "original_uuid", originalUUID, "error", err)
		return utils.ErrInternal
	}
	return utils.ErrAlreadyReversedAPI
}

func (h *ReverseLineItemHandler) checkIdempotency(ctx context.Context, req *dto.ReverseLineItemRequest, bill *entity.BillEntity) (*dto.ReverseLineItemResponse, error) {
	existing, err := h.LineItemRepo.FetchByBillAndKey(ctx, req.BillUUID, req.IdempotencyKey)
	if err != nil {
		if errors.Is(err, sqldb.ErrNoRows) {
			return nil, nil
		}
		rlog.Error("error checking idempotency", "bill_uuid", req.BillUUID, "idempotency_key", req.IdempotencyKey, "error", err)
		return nil, utils.ErrInternal
	}

	// Return existing reversal
	referenceUUID := ""
	if existing.ReferenceUUID != nil {
		referenceUUID = *existing.ReferenceUUID
	}

	return &dto.ReverseLineItemResponse{
		UUID:          existing.UUID,
		FeeType:       existing.FeeType,
		ReferenceUUID: referenceUUID,
		Amount: dto.Money{
			Amount:   existing.AmountCents,
			Currency: bill.Currency,
		},
		CreatedAt: existing.CreatedAt.Format(time.RFC3339),
	}, nil
}

func (h *ReverseLineItemHandler) queryWorkflowState(ctx context.Context, workflowID, billUUID string) error {
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

	return nil
}

func (h *ReverseLineItemHandler) buildReversalSignal(reversalUUID string, req *dto.ReverseLineItemRequest, original *entity.LineItemEntity) tbill.AddLineItemSignal {
	return tbill.AddLineItemSignal{
		UUID:           reversalUUID,
		IdempotencyKey: req.IdempotencyKey,
		FeeType:        string(entity.FeeTypeReversal),
		Description:    req.Reason,
		AmountCents:    -original.AmountCents, // Negative amount for reversal
		ReferenceUUID:  &req.LineItemUUID,     // Points to original line item
	}
}

func (h *ReverseLineItemHandler) signalWorkflow(ctx context.Context, workflowID, billUUID string, signal tbill.AddLineItemSignal) error {
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

func (h *ReverseLineItemHandler) buildPendingResponse(reversalUUID string, req *dto.ReverseLineItemRequest, original *entity.LineItemEntity, currency string) *dto.ReverseLineItemResponse {
	return &dto.ReverseLineItemResponse{
		UUID:          reversalUUID,
		FeeType:       string(entity.FeeTypeReversal),
		ReferenceUUID: req.LineItemUUID,
		Amount: dto.Money{
			Amount:   -original.AmountCents,
			Currency: currency,
		},
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}
}

func validateReverseLineItem(req *dto.ReverseLineItemRequest) []utils.ValidationError {
	var validationErrors []utils.ValidationError

	if req.BillUUID == "" {
		validationErrors = append(validationErrors, utils.ErrInvalidBillUUID)
	}
	if req.LineItemUUID == "" {
		validationErrors = append(validationErrors, utils.ErrInvalidLineItemUUID)
	}
	if req.IdempotencyKey == "" {
		validationErrors = append(validationErrors, utils.ErrInvalidIdempotencyKey)
	}

	return validationErrors
}
