package database

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gofiber/fiber/v2/log"
	"github.com/jmoiron/sqlx"
)

var ErrConfigInvalid = errors.New("invalid database configuration")

type Database struct {
	*sqlx.DB
	cfg *DBConfig
}

type DBConfig struct {
	Host            string
	Port            int
	User            string
	Password        string
	Database        string
	Params          map[string]string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	TLSConfigName   string
	LogQueries      bool
	RetryAttempts   int
	RetryBackoff    time.Duration
}

func NewDatabase(ctx context.Context, cfg *DBConfig) (*Database, error) {
	if err := validateConfig(cfg); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrConfigInvalid, err)
	}

	dsn, err := buildDSN(cfg)
	if err != nil {
		return nil, fmt.Errorf("build DSN: %w", err)
	}

	log.Debugw("database attempting connection",
		"user", cfg.User,
		"host", cfg.Host,
		"port", cfg.Port,
		"database", cfg.Database)

	db, err := sqlx.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("database open: %w", err)
	}

	if cfg.MaxOpenConns > 0 {
		db.SetMaxOpenConns(cfg.MaxOpenConns)
	}
	if cfg.MaxIdleConns > 0 {
		db.SetMaxIdleConns(cfg.MaxIdleConns)
	}
	if cfg.ConnMaxLifetime > 0 {
		db.SetConnMaxLifetime(cfg.ConnMaxLifetime)
	}

	if err := pingWithRetry(ctx, db, cfg); err != nil {
		db.Close()
		return nil, fmt.Errorf("database ping: %w", err)
	}

	log.Info("database connection established successfully")

	return &Database{
		DB:  db,
		cfg: cfg,
	}, nil
}

func (d *Database) Ping(ctx context.Context) error {
	if d.DB != nil {
		return d.DB.PingContext(ctx)
	}
	return errors.New("database connection is nil")
}

func (d *Database) Close() error {
	if d.DB != nil {
		return d.DB.Close()
	}
	return nil
}

// Shutdown gracefully closes the database connection
// Implements do.Shutdowner interface for dependency injection lifecycle management
func (d *Database) Shutdown() error {
	log.Info("database shutting down connection")
	return d.Close()
}

// HealthCheck verifies database connectivity
// Implements do.Healthchecker interface for dependency injection health checks
func (d *Database) HealthCheck() error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := d.Ping(ctx); err != nil {
		return fmt.Errorf("database health check failed: %w", err)
	}

	return nil
}

func validateConfig(cfg *DBConfig) error {
	if cfg == nil {
		return errors.New("config is nil")
	}
	if cfg.Host == "" {
		return errors.New("host is required")
	}
	if cfg.User == "" {
		return errors.New("user is required")
	}
	if cfg.Database == "" {
		return errors.New("database is required")
	}
	if cfg.Port <= 0 || cfg.Port > 65535 {
		return errors.New("port must be between 1 and 65535")
	}
	if cfg.MaxOpenConns < 0 {
		return errors.New("max open connections cannot be negative")
	}
	if cfg.MaxIdleConns < 0 {
		return errors.New("max idle connections cannot be negative")
	}
	if cfg.ConnMaxLifetime < 0 {
		return errors.New("connection max lifetime cannot be negative")
	}
	if cfg.RetryAttempts < 0 {
		return errors.New("retry attempts cannot be negative")
	}
	if cfg.RetryBackoff < 0 {
		return errors.New("retry backoff cannot be negative")
	}
	return nil
}

func buildDSN(cfg *DBConfig) (string, error) {
	params := url.Values{}
	params.Add("parseTime", "true")

	for k, v := range cfg.Params {
		params.Set(k, v)
	}

	if cfg.TLSConfigName != "" {
		params.Set("tls", cfg.TLSConfigName)
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?%s",
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.Database,
		params.Encode(),
	)

	return dsn, nil
}

func pingWithRetry(ctx context.Context, db *sqlx.DB, cfg *DBConfig) error {
	attempts := 1
	if cfg.RetryAttempts > 0 {
		attempts = cfg.RetryAttempts
	}

	backoff := time.Second
	if cfg.RetryBackoff > 0 {
		backoff = cfg.RetryBackoff
	}

	var lastErr error
	for i := 0; i < attempts; i++ {
		if i > 0 {
			log.Debugw("database ping attempt",
				"attempt", i+1,
				"max_attempts", attempts,
				"backoff", backoff)
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(backoff):
			}
		}

		if err := db.PingContext(ctx); err != nil {
			lastErr = err
			log.Warnw("database ping failed",
				"attempt", i+1,
				"max_attempts", attempts,
				"error", err)
			continue
		}

		return nil
	}

	return fmt.Errorf("failed after %d attempts: %w", attempts, lastErr)
}
