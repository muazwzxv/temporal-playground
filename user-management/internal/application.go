package app

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/muazwzxv/user-management/internal/config"
	"github.com/muazwzxv/user-management/internal/database"
	"github.com/muazwzxv/user-management/internal/database/store"
	"github.com/muazwzxv/user-management/internal/handler"
	healthHandler "github.com/muazwzxv/user-management/internal/handler/health"
	userHandler "github.com/muazwzxv/user-management/internal/handler/user"
	"github.com/muazwzxv/user-management/internal/redis"
	"github.com/muazwzxv/user-management/internal/repository"
	service "github.com/muazwzxv/user-management/internal/service/user"
	"github.com/muazwzxv/user-management/internal/worker"
	userWorkflow "github.com/muazwzxv/user-management/internal/worker/user"
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
	// Set log level from config
	if err := setLogLevel(cfg.Server.LogLevel); err != nil {
		return nil, fmt.Errorf("invalid log level: %w", err)
	}

	log.Infow("application config loaded",
		"server_host", cfg.Server.Host,
		"server_port", cfg.Server.Port,
		"database_host", cfg.Database.Host,
		"database_port", cfg.Database.Port,
		"log_level", cfg.Server.LogLevel)

	// Create DI container
	injector := do.New()

	// Provide configuration as a value (not lazy-loaded)
	do.ProvideValue(injector, cfg)

	// Provide infrastructure components
	do.Provide(injector, NewDatabase)
	do.Provide(injector, NewFiberApp)
	do.Provide(injector, NewQueries)
	do.Provide(injector, NewRedis)
	do.Provide(injector, worker.NewWorker)
	do.Provide(injector, userWorkflow.NewUserWorkflowRegistrar)

	// Provide repositories
	do.Provide(injector, repository.NewUserRepository)

	// Provide services
	do.Provide(injector, service.NewUserService)

	// Provide handlers
	do.Provide(injector, healthHandler.NewHealthHandler)
	do.Provide(injector, userHandler.NewUserHandler)

	app := do.MustInvoke[*fiber.App](injector)

	do.MustInvoke[*redis.Client](injector)

	// Register all routes
	RegisterRoutes(app, injector)

	log.Infow("application initialized successfully",
		"di_enabled", true,
		"providers_count", 9)

	return &Application{
		injector: injector,
		app:      app,
		config:   cfg,
	}, nil
}

// setLogLevel configures the global log level from config string
func setLogLevel(level string) error {
	switch strings.ToLower(level) {
	case "trace":
		log.SetLevel(log.LevelTrace)
	case "debug":
		log.SetLevel(log.LevelDebug)
	case "info":
		log.SetLevel(log.LevelInfo)
	case "warn", "warning":
		log.SetLevel(log.LevelWarn)
	case "error":
		log.SetLevel(log.LevelError)
	case "fatal":
		log.SetLevel(log.LevelFatal)
	case "panic":
		log.SetLevel(log.LevelPanic)
	default:
		return fmt.Errorf("unknown log level: %s (valid: trace, debug, info, warn, error, fatal, panic)", level)
	}
	return nil
}

// NewDatabase creates a new database connection from config
func NewDatabase(i do.Injector) (*database.Database, error) {
	cfg := do.MustInvoke[*config.Config](i)

	dbCfg := &database.DBConfig{
		Host:            cfg.Database.Host,
		Port:            cfg.Database.Port,
		User:            cfg.Database.User,
		Password:        cfg.Database.Password,
		Database:        cfg.Database.Database,
		MaxOpenConns:    cfg.Database.MaxOpenConns,
		MaxIdleConns:    cfg.Database.MaxIdleConns,
		ConnMaxLifetime: cfg.Database.ConnMaxLifetime,
		RetryAttempts:   cfg.Database.RetryAttempts,
		RetryBackoff:    cfg.Database.RetryBackoff,
	}

	return database.NewDatabase(context.Background(), dbCfg)
}

// NewFiberApp creates a new Fiber app from config with middleware
func NewFiberApp(i do.Injector) (*fiber.App, error) {
	cfg := do.MustInvoke[*config.Config](i)

	app := fiber.New(fiber.Config{
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		BodyLimit:    cfg.Server.BodyLimit,
		Prefork:      cfg.Server.Prefork,
	})

	// Register global middleware
	app.Use(logger.New())
	app.Use(recover.New())
	app.Use(handler.RequestLoggerMiddleware())
	app.Use(handler.ErrorHandlerMiddleware())
	app.Use(handler.CORSMiddleware())

	return app, nil
}

// NewQueries creates sqlc queries instance
func NewQueries(i do.Injector) (*store.Queries, error) {
	return store.New(), nil
}

// NewRedis creates a new Redis client from config
func NewRedis(i do.Injector) (*redis.Client, error) {
	cfg := do.MustInvoke[*config.Config](i)

	redisCfg := &redis.Config{
		Host:         cfg.Redis.Host,
		Port:         cfg.Redis.Port,
		Password:     cfg.Redis.Password,
		DB:           cfg.Redis.DB,
		MaxRetries:   cfg.Redis.MaxRetries,
		PoolSize:     cfg.Redis.PoolSize,
		MinIdleConns: cfg.Redis.MinIdleConns,
		DialTimeout:  cfg.Redis.DialTimeout,
		ReadTimeout:  cfg.Redis.ReadTimeout,
		WriteTimeout: cfg.Redis.WriteTimeout,
	}

	return redis.NewClient(context.Background(), redisCfg)
}

// RegisterRoutes registers all HTTP routes by invoking handlers from DI container
func RegisterRoutes(app *fiber.App, injector do.Injector) {
	// Invoke handlers from DI container and register their routes
	do.MustInvoke[*healthHandler.HealthHandler](injector).RegisterRoutes(app)
	do.MustInvoke[*userHandler.UserHandler](injector).RegisterRoutes(app)
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
			registrar := do.MustInvoke[*userWorkflow.UserWorkflowRegistrar](a.injector)
			w.RegisterWorkflows(registrar)
			return w.Start(ctx)
		})
	case RunModeBoth:
		g.Go(func() error {
			return a.StartHTTP(ctx)
		})
		g.Go(func() error {
			w := do.MustInvoke[*worker.Worker](a.injector)
			registrar := do.MustInvoke[*userWorkflow.UserWorkflowRegistrar](a.injector)
			w.RegisterWorkflows(registrar)
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
