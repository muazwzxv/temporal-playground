package billing

import (
	"context"
	"fmt"
	"log"

	"encore.app/db/repository"
	t "encore.app/temporal"

	tworker "go.temporal.io/sdk/worker"
)

//encore:service
type Service struct {
	cfg            *Config
	temporalClient t.WorkflowClient
	temporalWorker tworker.Worker

	// Repositories
	billRepo     repository.BillRepository
	lineItemRepo repository.LineItemRepository
	customerRepo repository.CustomerRepository
}

// encore automatically triggers initService as part
// of their application lifecycle
func initService() (*Service, error) {
	tc, err := t.NewClient(
		cfg.TemporalHost(),
		cfg.TemporalPort(),
		cfg.TemporalNamespace(),
	)
	if err != nil {
		return nil, fmt.Errorf("init temporal client: %w", err)
	}

	// Initialize repositories
	billRepo := &repository.BillRepo{DB: db}
	lineItemRepo := &repository.LineItemRepo{DB: db}
	customerRepo := &repository.CustomerRepo{DB: db}

	w := t.NewWorker(tc, billRepo, lineItemRepo)

	go func() {
		if err := w.Run(tworker.InterruptCh()); err != nil {
			log.Printf("temporal worker error: %v", err)
		}
	}()

	return &Service{
		cfg:            cfg,
		temporalClient: tc,
		temporalWorker: w,
		billRepo:       billRepo,
		lineItemRepo:   lineItemRepo,
		customerRepo:   customerRepo,
	}, nil
}

// encore automatically triggers shutdown as part of their
// graceful shutdown abstraction
func (s *Service) Shutdown(force context.Context) {
	s.temporalWorker.Stop()
}
