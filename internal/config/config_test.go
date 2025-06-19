package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name        string
		config      Config
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid config",
			config: Config{
				Port:            "8080",
				Host:            "localhost",
				Environment:     "dev",
				LogLevel:        "info",
				StorageType:     "memory",
				DatabaseDriver:  "sqlite3",
				MaxOpenConns:    25,
				MaxIdleConns:    25,
				ConnMaxLifetime: 5,
			},
			expectError: false,
		},
		{
			name: "invalid environment",
			config: Config{
				Port:        "8080",
				Environment: "invalid",
				LogLevel:    "info",
				StorageType: "memory",
			},
			expectError: true,
			errorMsg:    "invalid environment",
		},
		{
			name: "invalid log level",
			config: Config{
				Port:        "8080",
				Environment: "dev",
				LogLevel:    "invalid",
				StorageType: "memory",
			},
			expectError: true,
			errorMsg:    "invalid log level",
		},
		{
			name: "invalid storage type",
			config: Config{
				Port:        "8080",
				Environment: "dev",
				LogLevel:    "info",
				StorageType: "invalid",
			},
			expectError: true,
			errorMsg:    "invalid storage type",
		},
		{
			name: "invalid port",
			config: Config{
				Port:        "not-a-number",
				Environment: "dev",
				LogLevel:    "info",
				StorageType: "memory",
			},
			expectError: true,
			errorMsg:    "invalid port",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestConfig_IsDevelopment(t *testing.T) {
	tests := []struct {
		name        string
		environment string
		expected    bool
	}{
		{"dev", "dev", true},
		{"development", "development", true},
		{"staging", "staging", false},
		{"production", "production", false},
		{"prod", "prod", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := Config{Environment: tt.environment}
			assert.Equal(t, tt.expected, config.IsDevelopment())
		})
	}
}

func TestConfig_IsProduction(t *testing.T) {
	tests := []struct {
		name        string
		environment string
		expected    bool
	}{
		{"dev", "dev", false},
		{"development", "development", false},
		{"staging", "staging", false},
		{"production", "production", true},
		{"prod", "prod", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := Config{Environment: tt.environment}
			assert.Equal(t, tt.expected, config.IsProduction())
		})
	}
}

func TestLoad_EnvironmentVariables(t *testing.T) {
	// Save original environment
	originalEnv := map[string]string{
		"MCP_PORT":        os.Getenv("MCP_PORT"),
		"MCP_ENVIRONMENT": os.Getenv("MCP_ENVIRONMENT"),
		"MCP_LOG_LEVEL":   os.Getenv("MCP_LOG_LEVEL"),
	}

	// Clean up after test
	defer func() {
		for key, value := range originalEnv {
			if value == "" {
				os.Unsetenv(key)
			} else {
				os.Setenv(key, value)
			}
		}
	}()

	// Set test environment variables
	os.Setenv("MCP_PORT", "9999")
	os.Setenv("MCP_ENVIRONMENT", "test")
	os.Setenv("MCP_LOG_LEVEL", "debug")

	// Reset command line args to avoid interference
	oldArgs := os.Args
	os.Args = []string{"test"}
	defer func() { os.Args = oldArgs }()

	config, err := Load()
	require.NoError(t, err)

	assert.Equal(t, "9999", config.Port)
	assert.Equal(t, "test", config.Environment)
	assert.Equal(t, "debug", config.LogLevel)
}

func TestGetEnvInt(t *testing.T) {
	tests := []struct {
		name         string
		envValue     string
		defaultValue int
		expected     int
	}{
		{"valid integer", "42", 10, 42},
		{"invalid integer", "not-a-number", 10, 10},
		{"empty value", "", 10, 10},
		{"zero value", "0", 10, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variable
			if tt.envValue != "" {
				os.Setenv("TEST_INT_VAR", tt.envValue)
				defer os.Unsetenv("TEST_INT_VAR")
			}

			result := getEnvInt("TEST_INT_VAR", tt.defaultValue)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetEnvBool(t *testing.T) {
	tests := []struct {
		name         string
		envValue     string
		defaultValue bool
		expected     bool
	}{
		{"true", "true", false, true},
		{"1", "1", false, true},
		{"yes", "yes", false, true},
		{"on", "on", false, true},
		{"false", "false", true, false},
		{"0", "0", true, false},
		{"no", "no", true, false},
		{"off", "off", true, false},
		{"invalid", "invalid", true, true},
		{"empty", "", true, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variable
			if tt.envValue != "" {
				os.Setenv("TEST_BOOL_VAR", tt.envValue)
				defer os.Unsetenv("TEST_BOOL_VAR")
			}

			result := getEnvBool("TEST_BOOL_VAR", tt.defaultValue)
			assert.Equal(t, tt.expected, result)
		})
	}
}
