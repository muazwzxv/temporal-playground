package redis

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2/log"
	"github.com/redis/go-redis/v9"
)

var ErrConfigInvalid = errors.New("invalid redis configuration")

type Client struct {
	*redis.Client
	cfg *Config
}

type Config struct {
	Host         string
	Port         int
	Password     string
	DB           int
	MaxRetries   int
	PoolSize     int
	MinIdleConns int
	DialTimeout  time.Duration
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

func NewClient(ctx context.Context, cfg *Config) (*Client, error) {
	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	log.Debugw("redis attempting connection",
		"host", cfg.Host,
		"port", cfg.Port,
		"db", cfg.DB)

	client := redis.NewClient(&redis.Options{
		Addr:         addr,
		Password:     cfg.Password,
		DB:           cfg.DB,
		MaxRetries:   cfg.MaxRetries,
		PoolSize:     cfg.PoolSize,
		MinIdleConns: cfg.MinIdleConns,
		DialTimeout:  cfg.DialTimeout,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
	})

	// Verify connection
	if err := client.Ping(ctx).Err(); err != nil {
		client.Close()
		return nil, fmt.Errorf("redis ping: %w", err)
	}

	log.Info("redis connection established successfully")

	return &Client{
		Client: client,
		cfg:    cfg,
	}, nil
}

// Get retrieves a value by key
func (c *Client) Get(ctx context.Context, key string) (string, error) {
	val, err := c.Client.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return "", nil
		}
		return "", fmt.Errorf("redis get: %w", err)
	}
	return val, nil
}

// Set stores a key-value pair with optional expiration
func (c *Client) Set(ctx context.Context, key string, value any, expiration time.Duration) error {
	if err := c.Client.Set(ctx, key, value, expiration).Err(); err != nil {
		return fmt.Errorf("redis set: %w", err)
	}
	return nil
}

func (c *Client) Del(ctx context.Context, keys ...string) error {
	if err := c.Client.Del(ctx, keys...).Err(); err != nil {
		return fmt.Errorf("redis del: %w", err)
	}
	return nil
}

func (c *Client) Exists(ctx context.Context, keys ...string) (int64, error) {
	count, err := c.Client.Exists(ctx, keys...).Result()
	if err != nil {
		return 0, fmt.Errorf("redis exists: %w", err)
	}
	return count, nil
}

func (c *Client) Shutdown() error {
	log.Info("redis shutting down connection")
	return c.Client.Close()
}

func (c *Client) HealthCheck() error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := c.Client.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("redis health check failed: %w", err)
	}

	return nil
}
