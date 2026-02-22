package app

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"github.com/muazwzxv/user-management/internal/config"
	"github.com/muazwzxv/user-management/internal/handler"
	"github.com/muazwzxv/user-management/internal/repository"
	"github.com/muazwzxv/user-management/internal/service"
	"github.com/muazwzxv/user-management/internal/worker"
	"github.com/samber/do/v2"
	"golang.org/x/sync/errgroup"
)

// Application represents the main application with dependency injection
type Application struct {
	injector do.Injector
	app      *fiber.App
	config   *config.Config
}

// NewApplication creates a new application with all dependencies wired via DI container
func NewApplication(cfg *config.Config) (*Application, error) {
	if err := setLogLevel(cfg.Server.LogLevel); err != nil {
		return nil, fmt.Errorf("invalid log level: %w", err)
	}

	log.Infow("application config loaded",
		"server_host", cfg.Server.Host,
		"server_port", cfg.Server.Port,
		"database_host", cfg.Database.Host,
		"database_port", cfg.Database.Port,
		"log_level", cfg.Server.LogLevel)

	injector := do.New()

	do.ProvideValue(injector, cfg)

	// Provide infrastructure components
	do.Provide(injector, NewFiberApp)
	do.Provide(injector, NewDatabase)
	do.Provide(injector, NewQueries)
	do.Provide(injector, NewRedis)

	app := do.MustInvoke[*fiber.App](injector)

	repository.InjectRepository(injector)
	service.InjectServices(injector)
	worker.InjectWorkflow(injector)
	handler.InjectHandler(injector, app)

	log.Infow("application initialized successfully",
		"di_enabled", true,
		"providers_count", 9)

	return &Application{
		injector: injector,
		app:      app,
		config:   cfg,
	}, nil
}

type RunMode string

const (
	RunModeHTTP   RunMode = "http"
	RunModeWorker RunMode = "worker"
	RunModeBoth   RunMode = "both"
)

// StartHTTP starts the HTTP server with graceful shutdown support
func (a *Application) StartHTTP(ctx context.Context) error {
	addr := fmt.Sprintf("%s:%d", a.config.Server.Host, a.config.Server.Port)
	log.Infow("starting http server",
		"address", addr,
		"host", a.config.Server.Host,
		"port", a.config.Server.Port)

	// Start server in goroutine
	errChan := make(chan error, 1)
	go func() {
		if err := a.app.Listen(addr); err != nil {
			errChan <- err
		}
	}()

	// Wait for context cancellation or server error
	select {
	case <-ctx.Done():
		log.Info("shutting down http server")
		if err := a.app.Shutdown(); err != nil {
			log.Errorw("http server shutdown error", "error", err)
		}
		// Shutdown DI container
		if err := a.injector.Shutdown(); err != nil {
			log.Errorw("di container shutdown error", "error", err)
		}
		return nil
	case err := <-errChan:
		return err
	}
}

// Start runs the application based on the specified mode with graceful shutdown
func (a *Application) Start(ctx context.Context, mode RunMode) error {
	// Create a cancellable context for coordinated shutdown
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Setup signal handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	g, ctx := errgroup.WithContext(ctx)

	switch mode {
	case RunModeHTTP:
		g.Go(func() error {
			return a.StartHTTP(ctx)
		})
	case RunModeWorker:
		g.Go(func() error {
			w := do.MustInvoke[*worker.Worker](a.injector)
			return w.Start(ctx)
		})
	case RunModeBoth:
		g.Go(func() error {
			return a.StartHTTP(ctx)
		})
		g.Go(func() error {
			w := do.MustInvoke[*worker.Worker](a.injector)
			return w.Start(ctx)
		})
	default:
		return fmt.Errorf("unknown run mode: %s", mode)
	}

	// Wait for shutdown signal in a separate goroutine
	g.Go(func() error {
		select {
		case sig := <-sigChan:
			log.Infow("received shutdown signal", "signal", sig)
			cancel()
			return nil
		case <-ctx.Done():
			return nil
		}
	})

	if err := g.Wait(); err != nil {
		return fmt.Errorf("application error: %w", err)
	}

	log.Info("application shutdown complete")
	return nil
}

// Init initializes the application with config loaded from default locations
func Init() (*Application, error) {
	cfg, err := config.LoadConfig()
	if err != nil {
		return nil, fmt.Errorf("load config: %w", err)
	}

	return NewApplication(cfg)
}

// InitFromFile initializes the application with config loaded from a specific file
func InitFromFile(configPath string) (*Application, error) {
	cfg, err := config.Load(configPath)
	if err != nil {
		return nil, fmt.Errorf("load config: %w", err)
	}

	return NewApplication(cfg)
}
