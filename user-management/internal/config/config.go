// Package config provides configuration loading from multiple sources using Viper.
package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// Config holds all application configuration
type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	Temporal TemporalConfig `mapstructure:"temporal"`
}

// TemporalConfig holds Temporal worker configuration
type TemporalConfig struct {
	Host      string `mapstructure:"host"`
	Namespace string `mapstructure:"namespace"`
	QueueName string `mapstructure:"queue_name"`
}

// ServerConfig holds Fiber server configuration
type ServerConfig struct {
	Host         string        `mapstructure:"host"`
	Port         int           `mapstructure:"port"`
	LogLevel     string        `mapstructure:"log_level"`
	ReadTimeout  time.Duration `mapstructure:"read_timeout"`
	WriteTimeout time.Duration `mapstructure:"write_timeout"`
	BodyLimit    int           `mapstructure:"body_limit"`
	Prefork      bool          `mapstructure:"prefork"`
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	Host            string        `mapstructure:"host"`
	Port            int           `mapstructure:"port"`
	User            string        `mapstructure:"user"`
	Password        string        `mapstructure:"password"`
	Database        string        `mapstructure:"database"`
	MaxOpenConns    int           `mapstructure:"max_open_conns"`
	MaxIdleConns    int           `mapstructure:"max_idle_conns"`
	ConnMaxLifetime time.Duration `mapstructure:"conn_max_lifetime"`
	RetryAttempts   int           `mapstructure:"retry_attempts"`
	RetryBackoff    time.Duration `mapstructure:"retry_backoff"`
}

// LoadConfig loads configuration from TOML file and environment variables.
//
// Configuration priority (highest to lowest):
//  1. Environment variables
//  2. config.toml file
//  3. Default values
//
// Environment variable naming convention:
//   - Nested config keys are flattened with underscores
//   - All uppercase
//   - Examples:
//     server.host              → SERVER_HOST
//     server.port              → SERVER_PORT
//     server.log_level         → SERVER_LOG_LEVEL
//     database.host            → DATABASE_HOST
//     database.max_open_conns  → DATABASE_MAX_OPEN_CONNS
func LoadConfig() (*Config, error) {
	v := viper.New()

	// Set defaults to ensure all keys exist for env var binding
	setDefaults(v)

	// Configure config file
	v.SetConfigName("config")
	v.SetConfigType("toml")
	v.AddConfigPath(".")
	v.AddConfigPath("/etc/user-management/")
	v.AddConfigPath("$HOME/.user-management")

	// Read config file (optional - don't fail if not found)
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			// Config file was found but another error occurred
			return nil, fmt.Errorf("read config file: %w", err)
		}
		// Config file not found; will use defaults + environment variables
	}

	// Enable automatic environment variable override
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Unmarshal into struct
	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}

	return &cfg, nil
}

// setDefaults sets default configuration values
func setDefaults(v *viper.Viper) {
	// Server defaults
	v.SetDefault("server.host", "0.0.0.0")
	v.SetDefault("server.port", 8080)
	v.SetDefault("server.log_level", "info")
	v.SetDefault("server.read_timeout", "5s")
	v.SetDefault("server.write_timeout", "10s")
	v.SetDefault("server.body_limit", 4194304) // 4MB
	v.SetDefault("server.prefork", false)

	// Database defaults
	v.SetDefault("database.host", "localhost")
	v.SetDefault("database.port", 3306)
	v.SetDefault("database.user", "root")
	v.SetDefault("database.password", "")
	v.SetDefault("database.database", "user-management")
	v.SetDefault("database.max_open_conns", 25)
	v.SetDefault("database.max_idle_conns", 10)
	v.SetDefault("database.conn_max_lifetime", "5m")
	v.SetDefault("database.retry_attempts", 3)
	v.SetDefault("database.retry_backoff", "2s")

	// Temporal defaults
	v.SetDefault("temporal.host", "localhost:7233")
	v.SetDefault("temporal.namespace", "default")
	v.SetDefault("temporal.queue_name", "user-management-queue")
}

// Load reads configuration from a TOML file (backward compatibility).
func Load(path string) (*Config, error) {
	v := viper.New()

	// Read specific config file
	v.SetConfigFile(path)
	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("read config file: %w", err)
	}

	// Unmarshal into struct
	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}

	return &cfg, nil
}
