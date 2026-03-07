package billing

import (
	"context"

	"encore.app/dto"
	"encore.app/handlers"
)

// Customer endpoints

//encore:api public method=POST path=/v1/customer/create
func (s *Service) CreateCustomer(ctx context.Context, req *dto.CreateCustomerRequest) (*dto.CreateCustomerResponse, error) {
	h := handlers.CreateCustomerHandler{
		CustomerRepo: s.customerRepo,
	}
	return h.Handle(ctx, req)
}

//encore:api public method=POST path=/v1/customer/get
func (s *Service) GetCustomer(ctx context.Context, req *dto.GetCustomerRequest) (*dto.GetCustomerResponse, error) {
	h := handlers.GetCustomerHandler{
		CustomerRepo: s.customerRepo,
	}
	return h.Handle(ctx, req)
}

// Billing endpoints

//encore:api public method=POST path=/v1/bill/create
func (s *Service) CreateBill(ctx context.Context, req *dto.CreateBillRequest) (*dto.CreateBillResponse, error) {
	h := handlers.CreateBillHandler{
		BillRepo:       s.billRepo,
		CustomerRepo:   s.customerRepo,
		TemporalClient: s.temporalClient,
	}
	return h.Handle(ctx, req)
}

//encore:api public method=POST path=/v1/bill/add-line-item
func (s *Service) AddLineItem(ctx context.Context, req *dto.AddLineItemRequest) (*dto.AddLineItemResponse, error) {
	h := handlers.AddLineItemHandler{
		BillRepo:       s.billRepo,
		LineItemRepo:   s.lineItemRepo,
		TemporalClient: s.temporalClient,
	}
	return h.Handle(ctx, req)
}

//encore:api public method=POST path=/v1/bill/get
func (s *Service) GetBill(ctx context.Context, req *dto.GetBillRequest) (*dto.GetBillResponse, error) {
	h := handlers.GetBillHandler{
		BillRepo: s.billRepo,
	}
	return h.Handle(ctx, req)
}

//encore:api public method=POST path=/v1/bill/close
func (s *Service) CloseBill(ctx context.Context, req *dto.CloseBillRequest) (*dto.CloseBillResponse, error) {
	h := handlers.CloseBillHandler{
		BillRepo:       s.billRepo,
		TemporalClient: s.temporalClient,
	}
	return h.Handle(ctx, req)
}

//encore:api public method=POST path=/v1/bill/list
func (s *Service) ListBills(ctx context.Context, req *dto.ListBillsRequest) (*dto.ListBillsResponse, error) {
	h := handlers.ListBillsHandler{
		BillRepo: s.billRepo,
	}
	return h.Handle(ctx, req)
}

//encore:api public method=POST path=/v1/bill/list-line-items
func (s *Service) ListLineItems(ctx context.Context, req *dto.ListLineItemsRequest) (*dto.ListLineItemsResponse, error) {
	h := handlers.ListLineItemsHandler{
		BillRepo:     s.billRepo,
		LineItemRepo: s.lineItemRepo,
	}
	return h.Handle(ctx, req)
}

//encore:api public method=POST path=/v1/bill/reverse-line-item
func (s *Service) ReverseLineItem(ctx context.Context, req *dto.ReverseLineItemRequest) (*dto.ReverseLineItemResponse, error) {
	h := handlers.ReverseLineItemHandler{
		BillRepo:       s.billRepo,
		LineItemRepo:   s.lineItemRepo,
		TemporalClient: s.temporalClient,
	}
	return h.Handle(ctx, req)
}
