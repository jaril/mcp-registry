package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateServer(t *testing.T) {
	tests := []struct {
		name          string
		server        Server
		expectedError bool
		errorMessage  string
	}{
		{
			name: "valid server",
			server: Server{
				ID:          "test-1",
				Name:        "test-server",
				Description: "A test server",
				Version:     "1.0.0",
				Repository:  "https://github.com/test/test",
				Author:      "Test Author",
				Tags:        []string{"test"},
				IsActive:    true,
			},
			expectedError: false,
		},
		{
			name: "missing ID",
			server: Server{
				Name:        "test-server",
				Description: "A test server",
				Version:     "1.0.0",
			},
			expectedError: true,
			errorMessage:  "id: is required",
		},
		{
			name: "missing name",
			server: Server{
				ID:          "test-1",
				Description: "A test server",
				Version:     "1.0.0",
			},
			expectedError: true,
			errorMessage:  "name: is required",
		},
		{
			name: "missing version",
			server: Server{
				ID:          "test-1",
				Name:        "test-server",
				Description: "A test server",
			},
			expectedError: true,
			errorMessage:  "version: is required",
		},
		{
			name: "multiple validation errors",
			server: Server{
				Description: "A test server",
			},
			expectedError: true,
			errorMessage:  "validation failed:",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateServer(tt.server)

			if tt.expectedError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMessage)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidationError(t *testing.T) {
	t.Run("single validation error", func(t *testing.T) {
		err := ValidationError{
			Field:   "email",
			Message: "is required",
		}

		assert.Equal(t, "email: is required", err.Error())
	})
}

func TestValidationErrors(t *testing.T) {
	t.Run("multiple validation errors", func(t *testing.T) {
		errors := ValidationErrors{
			{Field: "id", Message: "is required"},
			{Field: "name", Message: "is required"},
			{Field: "version", Message: "is required"},
		}

		expectedMessage := "validation failed: id: is required, name: is required, version: is required"
		assert.Equal(t, expectedMessage, errors.Error())
	})

	t.Run("empty validation errors", func(t *testing.T) {
		errors := ValidationErrors{}
		assert.Equal(t, "no validation errors", errors.Error())
	})
}

// Benchmark for validation performance
func BenchmarkValidateServer(b *testing.B) {
	server := Server{
		ID:          "benchmark-test",
		Name:        "benchmark-server",
		Description: "A server for benchmarking",
		Version:     "1.0.0",
		Repository:  "https://github.com/test/benchmark",
		Author:      "Benchmark Author",
		Tags:        []string{"benchmark", "test"},
		IsActive:    true,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ValidateServer(server)
	}
}
