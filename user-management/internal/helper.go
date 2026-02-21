package app

import (
	"context"
	"fmt"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/muazwzxv/user-management/internal/config"
	"github.com/muazwzxv/user-management/internal/database"
	"github.com/muazwzxv/user-management/internal/database/store"
	"github.com/muazwzxv/user-management/internal/handler"
	"github.com/muazwzxv/user-management/internal/redis"
	"github.com/samber/do/v2"
)

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
