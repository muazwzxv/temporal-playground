package app

import (
	"context"
	"fmt"
	"os"
	"strings"
	"syscall"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/samber/do/v2"
	"github.com/muazwzxv/user-management/internal/config"
	"github.com/muazwzxv/user-management/internal/database"
	"github.com/muazwzxv/user-management/internal/database/store"
	"github.com/muazwzxv/user-management/internal/handler"
	healthHandler "github.com/muazwzxv/user-management/internal/handler/health"
	userHandler "github.com/muazwzxv/user-management/internal/handler/user"
	"github.com/muazwzxv/user-management/internal/repository"
	service "github.com/muazwzxv/user-management/internal/service/user"
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

	// Provide repositories
	do.Provide(injector, repository.NewUserRepository)

	// Provide services
	do.Provide(injector, service.NewUserService)

	// Provide handlers
	do.Provide(injector, healthHandler.NewHealthHandler)
	do.Provide(injector, userHandler.NewUserHandler)

	// Invoke fiber app to initialize it and register routes
	app := do.MustInvoke[*fiber.App](injector)

	// Register all routes
	RegisterRoutes(app, injector)

	log.Infow("application initialized successfully",
		"di_enabled", true,
		"providers_count", 8)

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

// RegisterRoutes registers all HTTP routes by invoking handlers from DI container
func RegisterRoutes(app *fiber.App, injector do.Injector) {
	// Invoke handlers from DI container and register their routes
	do.MustInvoke[*healthHandler.HealthHandler](injector).RegisterRoutes(app)
	do.MustInvoke[*userHandler.UserHandler](injector).RegisterRoutes(app)
}

// Start starts the HTTP server and handles graceful shutdown
func (a *Application) Start() error {
	addr := fmt.Sprintf("%s:%d", a.config.Server.Host, a.config.Server.Port)
	log.Infow("application starting server",
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

	// Handle graceful shutdown with DI container
	go func() {
		// ShutdownOnSignals blocks until signal is received
		signal, _ := a.injector.RootScope().ShutdownOnSignals(syscall.SIGTERM, os.Interrupt)
		log.Infow("application received shutdown signal",
			"signal", signal)
		
		// Shutdown Fiber server
		if err := a.app.Shutdown(); err != nil {
			log.Errorw("application shutdown error",
				"error", err)
		}
		
		log.Info("application shutdown complete")
		errChan <- nil
	}()

	return <-errChan
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
