package config

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
)

// Config holds all configuration for the application
type Config struct {
	// Server configuration
	Port    string `json:"port"`
	Host    string `json:"host"`
	Address string `json:"address"` // Computed from Host:Port

	// Application configuration
	Environment string `json:"environment"` // dev, staging, production
	LogLevel    string `json:"log_level"`   // debug, info, warn, error
	Version     string `json:"version"`

	// Storage configuration
	StorageType string `json:"storage_type"` // memory, file, database
	DataPath    string `json:"data_path"`    // For file storage

	// Feature flags
	EnableMetrics bool `json:"enable_metrics"`
	EnableCORS    bool `json:"enable_cors"`

	// Database configuration
	DatabaseURL     string `json:"database_url"`      // Connection string
	DatabaseDriver  string `json:"database_driver"`   // sqlite3, postgres, mysql
	DatabaseName    string `json:"database_name"`     // Database name
	MaxOpenConns    int    `json:"max_open_conns"`    // Connection pool size
	MaxIdleConns    int    `json:"max_idle_conns"`    // Idle connections
	ConnMaxLifetime int    `json:"conn_max_lifetime"` // Connection lifetime (minutes
}

// Load reads configuration from environment variables, command line flags, and defaults
func Load() (*Config, error) {
	cfg := &Config{}

	// Define command line flags
	var (
		port           = flag.String("port", "", "Server port (default: 8080)")
		host           = flag.String("host", "", "Server host (default: localhost)")
		environment    = flag.String("env", "", "Environment: dev, test, staging, production (default: dev)")
		logLevel       = flag.String("log-level", "", "Log level: debug, info, warn, error (default: info)")
		storageType    = flag.String("storage", "", "Storage type: memory, sqlite, database (default: memory)")
		version        = flag.Bool("version", false, "Show version and exit")
		help           = flag.Bool("help", false, "Show help and exit")
		databaseURL    = flag.String("db-url", "", "Database URL (default: ./data/registry.db)")
		databaseDriver = flag.String("db-driver", "", "Database driver: sqlite3, postgres (default: sqlite3)")
	)

	flag.Parse()

	// Show help if requested
	if *help {
		fmt.Println("MCP Registry Server")
		fmt.Println("Configuration options:")
		flag.PrintDefaults()
		fmt.Println("\nEnvironment variables:")
		fmt.Println("  MCP_PORT          Server port")
		fmt.Println("  MCP_HOST          Server host")
		fmt.Println("  MCP_ENVIRONMENT   Environment (dev/test/staging/production)")
		fmt.Println("  MCP_LOG_LEVEL     Log level (debug/info/warn/error)")
		fmt.Println("  MCP_STORAGE_TYPE  Storage type (memory/sqlite/database)")
		os.Exit(0)
	}

	// Show version if requested
	if *version {
		fmt.Printf("MCP Registry v%s\n", getEnvOr("MCP_VERSION", "dev"))
		os.Exit(0)
	}

	// Load configuration with precedence: flags > env vars > defaults
	cfg.Port = getConfigValue(*port, "MCP_PORT", "8080")
	cfg.Host = getConfigValue(*host, "MCP_HOST", "localhost")
	cfg.Environment = getConfigValue(*environment, "MCP_ENVIRONMENT", "dev")
	cfg.LogLevel = getConfigValue(*logLevel, "MCP_LOG_LEVEL", "info")
	cfg.StorageType = getConfigValue(*storageType, "MCP_STORAGE_TYPE", "sqlite")
	cfg.Version = getEnvOr("MCP_VERSION", "dev")
	cfg.DataPath = getEnvOr("MCP_DATA_PATH", "./data")
	cfg.DatabaseURL = getConfigValue(*databaseURL, "MCP_DATABASE_URL", "./data/registry.db")
	cfg.DatabaseDriver = getConfigValue(*databaseDriver, "MCP_DATABASE_DRIVER", "sqlite3")
	cfg.DatabaseName = getEnvOr("MCP_DATABASE_NAME", "mcp_registry")
	cfg.MaxOpenConns = getEnvInt("MCP_MAX_OPEN_CONNS", 25)
	cfg.MaxIdleConns = getEnvInt("MCP_MAX_IDLE_CONNS", 25)
	cfg.ConnMaxLifetime = getEnvInt("MCP_CONN_MAX_LIFETIME", 5)

	// Feature flags
	cfg.EnableMetrics = getEnvBool("MCP_ENABLE_METRICS", false)
	cfg.EnableCORS = getEnvBool("MCP_ENABLE_CORS", true)

	// Compute derived values
	cfg.Address = cfg.Host + ":" + cfg.Port

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return cfg, nil
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	// Validate environment
	switch c.Environment {
	case "dev", "development", "test", "staging", "prod", "production":
		// Valid environments
	default:
		return fmt.Errorf("invalid environment: %s (must be dev/test/staging/production)", c.Environment)
	}

	// Validate log level
	switch c.LogLevel {
	case "debug", "info", "warn", "error":
		// Valid log levels
	default:
		return fmt.Errorf("invalid log level: %s (must be debug/info/warn/error)", c.LogLevel)
	}

	// Validate storage type
	switch c.StorageType {
	case "memory", "file", "sqlite", "database":
		// Valid storage types (database will be added later)
	default:
		return fmt.Errorf("invalid storage type: %s (must be memory/file)", c.StorageType)
	}

	// Validate port is numeric
	if _, err := strconv.Atoi(c.Port); err != nil {
		return fmt.Errorf("invalid port: %s (must be numeric)", c.Port)
	}

	// Validate database driver
	switch c.DatabaseDriver {
	case "sqlite3", "postgres", "mysql":
		// Valid drivers
	default:
		return fmt.Errorf("invalid database driver: %s (must be sqlite3/postgres/mysql)", c.DatabaseDriver)
	}

	// Validate connection pool settings
	if c.MaxOpenConns <= 0 {
		return fmt.Errorf("max_open_conns must be positive, got: %d", c.MaxOpenConns)
	}

	if c.MaxIdleConns <= 0 {
		return fmt.Errorf("max_idle_conns must be positive, got: %d", c.MaxIdleConns)
	}

	return nil
}

// IsDevelopment returns true if running in development mode
func (c *Config) IsDevelopment() bool {
	return c.Environment == "dev" || c.Environment == "development" || c.Environment == "test"
}

// IsProduction returns true if running in production mode
func (c *Config) IsProduction() bool {
	return c.Environment == "prod" || c.Environment == "production"
}

// LogConfig prints the current configuration (excluding sensitive data)
func (c *Config) LogConfig() {
	log.Printf("Configuration loaded:")
	log.Printf("  Environment: %s", c.Environment)
	log.Printf("  Address: %s", c.Address)
	log.Printf("  Log Level: %s", c.LogLevel)
	log.Printf("  Storage Type: %s", c.StorageType)
	log.Printf("  Enable CORS: %v", c.EnableCORS)
	log.Printf("  Enable Metrics: %v", c.EnableMetrics)
}

// getConfigValue returns the first non-empty value from flag, env var, or default
func getConfigValue(flagValue, envKey, defaultValue string) string {
	if flagValue != "" {
		return flagValue
	}
	if envValue := os.Getenv(envKey); envValue != "" {
		return envValue
	}
	return defaultValue
}

// getEnvOr returns environment variable value or default
func getEnvOr(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvBool returns environment variable as boolean or default
func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		switch value {
		case "true", "1", "yes", "on":
			return true
		case "false", "0", "no", "off":
			return false
		}
	}
	return defaultValue
}

// Add helper function for integer environment variables:
func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}
