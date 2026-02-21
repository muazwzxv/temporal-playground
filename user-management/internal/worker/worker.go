package worker

import (
	"context"
	"fmt"

	"github.com/gofiber/fiber/v2/log"
	"github.com/muazwzxv/user-management/internal/config"
	"github.com/muazwzxv/user-management/internal/worker/createuser"
	"github.com/samber/do/v2"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
)

type Worker struct {
	temporalClient client.Client
	worker         worker.Worker
	config         *config.TemporalConfig
	registrars     []WorkflowRegistrar
}

func InjectWorkflow(i do.Injector) {
	do.Provide(i, createuser.NewUserWorkflowRegistrar)

	do.Provide(i, NewWorker)
}

func NewWorker(i do.Injector) (*Worker, error) {
	cfg := do.MustInvoke[*config.Config](i)

	c, err := client.Dial(client.Options{
		HostPort:  cfg.Temporal.Host,
		Namespace: cfg.Temporal.Namespace,
	})
	if err != nil {
		return nil, fmt.Errorf("create temporal client: %w", err)
	}

	w := worker.New(c, cfg.Temporal.QueueName, worker.Options{})

	registrars := []WorkflowRegistrar{
		do.MustInvoke[*createuser.UserWorkflowRegistrar](i),
		// Add future workflow registrars here
	}

	log.Infow("temporal worker initialized",
		"host", cfg.Temporal.Host,
		"namespace", cfg.Temporal.Namespace,
		"queue", cfg.Temporal.QueueName,
		"registrars_count", len(registrars))

	return &Worker{
		temporalClient: c,
		worker:         w,
		config:         &cfg.Temporal,
		registrars:     registrars,
	}, nil
}

func (w *Worker) Client() client.Client {
	return w.temporalClient
}

func (w *Worker) Start(ctx context.Context) error {
	for _, r := range w.registrars {
		r.Register(w.worker)
	}

	log.Infow("starting temporal worker",
		"host", w.config.Host,
		"namespace", w.config.Namespace,
		"queue", w.config.QueueName,
		"registrars_count", len(w.registrars))

	errChan := make(chan error, 1)
	go func() {
		errChan <- w.worker.Run(worker.InterruptCh())
	}()

	select {
	case <-ctx.Done():
		w.Stop()
		return nil
	case err := <-errChan:
		if err != nil {
			return fmt.Errorf("worker error: %w", err)
		}
		return nil
	}
}

func (w *Worker) Stop() {
	log.Info("shutting down temporal worker")
	w.worker.Stop()
	w.temporalClient.Close()
}

func (w *Worker) Shutdown() error {
	w.Stop()
	return nil
}
