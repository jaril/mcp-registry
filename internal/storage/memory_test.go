package storage

import (
	"fmt"
	"testing"
	"time"

	"registry/internal/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMemoryStore_GetAll(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(*MemoryStore)
		expected int
	}{
		{
			name:     "empty store",
			setup:    func(store *MemoryStore) {},
			expected: 0,
		},
		{
			name: "store with servers",
			setup: func(store *MemoryStore) {
				server1 := createTestServer("1", "server1")
				server2 := createTestServer("2", "server2")
				store.Create(server1)
				store.Create(server2)
			},
			expected: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := NewMemoryStore()
			tt.setup(store)

			servers, err := store.GetAll()

			require.NoError(t, err)
			assert.Len(t, servers, tt.expected)
		})
	}
}

func TestMemoryStore_GetByID(t *testing.T) {
	store := NewMemoryStore()
	testServer := createTestServer("test-1", "test-server")

	// Add server to store
	err := store.Create(testServer)
	require.NoError(t, err)

	tests := []struct {
		name        string
		id          string
		expectError bool
		errorType   error
	}{
		{
			name:        "existing server",
			id:          "test-1",
			expectError: false,
		},
		{
			name:        "non-existent server",
			id:          "non-existent",
			expectError: true,
			errorType:   ErrServerNotFound,
		},
		{
			name:        "empty ID",
			id:          "",
			expectError: true,
			errorType:   ErrInvalidID,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server, err := store.GetByID(tt.id)

			if tt.expectError {
				assert.Error(t, err)
				assert.ErrorIs(t, err, tt.errorType)
				assert.Nil(t, server)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, server)
				assert.Equal(t, tt.id, server.ID)
			}
		})
	}
}

func TestMemoryStore_Create(t *testing.T) {
	tests := []struct {
		name        string
		server      models.Server
		setup       func(*MemoryStore)
		expectError bool
		errorType   error
	}{
		{
			name:        "valid server",
			server:      createTestServer("test-1", "test-server"),
			setup:       func(store *MemoryStore) {},
			expectError: false,
		},
		{
			name:   "duplicate server",
			server: createTestServer("test-1", "test-server"),
			setup: func(store *MemoryStore) {
				existingServer := createTestServer("test-1", "existing-server")
				store.Create(existingServer)
			},
			expectError: true,
			errorType:   ErrServerExists,
		},
		{
			name: "invalid server - missing ID",
			server: models.Server{
				Name:        "test-server",
				Description: "A test server",
				Version:     "1.0.0",
			},
			setup:       func(store *MemoryStore) {},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := NewMemoryStore()
			tt.setup(store)

			err := store.Create(tt.server)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorType != nil {
					assert.ErrorIs(t, err, tt.errorType)
				}
			} else {
				assert.NoError(t, err)

				// Verify server was actually created
				createdServer, getErr := store.GetByID(tt.server.ID)
				assert.NoError(t, getErr)
				assert.Equal(t, tt.server.Name, createdServer.Name)
			}
		})
	}
}

func TestMemoryStore_Update(t *testing.T) {
	store := NewMemoryStore()
	originalServer := createTestServer("test-1", "original-server")

	// Create initial server
	err := store.Create(originalServer)
	require.NoError(t, err)

	tests := []struct {
		name        string
		server      models.Server
		expectError bool
		errorType   error
	}{
		{
			name: "valid update",
			server: models.Server{
				ID:          "test-1",
				Name:        "updated-server",
				Description: "An updated server",
				Version:     "2.0.0",
				Repository:  "https://github.com/test/updated",
				Author:      "Updated Author",
				Tags:        []string{"updated"},
				IsActive:    false,
			},
			expectError: false,
		},
		{
			name: "update non-existent server",
			server: models.Server{
				ID:          "non-existent",
				Name:        "non-existent-server",
				Description: "Does not exist",
				Version:     "1.0.0",
				Repository:  "https://github.com/test/nonexistent",
				Author:      "Test Author",
				Tags:        []string{"test"},
				IsActive:    true,
			},
			expectError: true,
			errorType:   ErrServerNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := store.Update(tt.server)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorType != nil {
					assert.ErrorIs(t, err, tt.errorType)
				}
			} else {
				assert.NoError(t, err)

				// Verify server was actually updated
				updatedServer, getErr := store.GetByID(tt.server.ID)
				assert.NoError(t, getErr)
				assert.Equal(t, tt.server.Name, updatedServer.Name)
				assert.Equal(t, tt.server.Version, updatedServer.Version)
				// Verify creation time was preserved
				assert.Equal(t, originalServer.CreatedAt, updatedServer.CreatedAt)
			}
		})
	}
}

func TestMemoryStore_Delete(t *testing.T) {
	store := NewMemoryStore()
	testServer := createTestServer("test-1", "test-server")

	// Add server to store
	err := store.Create(testServer)
	require.NoError(t, err)

	tests := []struct {
		name        string
		id          string
		expectError bool
		errorType   error
	}{
		{
			name:        "delete existing server",
			id:          "test-1",
			expectError: false,
		},
		{
			name:        "delete non-existent server",
			id:          "non-existent",
			expectError: true,
			errorType:   ErrServerNotFound,
		},
		{
			name:        "delete with empty ID",
			id:          "",
			expectError: true,
			errorType:   ErrInvalidID,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := store.Delete(tt.id)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorType != nil {
					assert.ErrorIs(t, err, tt.errorType)
				}
			} else {
				assert.NoError(t, err)

				// Verify server was actually deleted
				_, getErr := store.GetByID(tt.id)
				assert.ErrorIs(t, getErr, ErrServerNotFound)
			}
		})
	}
}

func TestMemoryStore_Search(t *testing.T) {
	store := NewMemoryStore()

	// Add test servers
	servers := []models.Server{
		createTestServer("1", "filesystem-server"),
		createTestServer("2", "web-server"),
		createTestServer("3", "database-server"),
		createTestServer("4", "file-processor"),
	}

	for _, server := range servers {
		err := store.Create(server)
		require.NoError(t, err)
	}

	tests := []struct {
		name          string
		query         string
		expectedCount int
		expectedNames []string
	}{
		{
			name:          "search for 'server'",
			query:         "server",
			expectedCount: 3,
			expectedNames: []string{"filesystem-server", "web-server", "database-server"},
		},
		{
			name:          "search for 'file'",
			query:         "file",
			expectedCount: 2,
			expectedNames: []string{"filesystem-server", "file-processor"},
		},
		{
			name:          "search for 'web'",
			query:         "web",
			expectedCount: 1,
			expectedNames: []string{"web-server"},
		},
		{
			name:          "search for non-existent",
			query:         "nonexistent",
			expectedCount: 0,
			expectedNames: []string{},
		},
		{
			name:          "empty search",
			query:         "",
			expectedCount: 0,
			expectedNames: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results, err := store.Search(tt.query)

			assert.NoError(t, err)
			assert.Len(t, results, tt.expectedCount)

			resultNames := make([]string, len(results))
			for i, server := range results {
				resultNames[i] = server.Name
			}

			for _, expectedName := range tt.expectedNames {
				assert.Contains(t, resultNames, expectedName)
			}
		})
	}
}

func TestMemoryStore_Count(t *testing.T) {
	store := NewMemoryStore()

	// Test empty store
	total, active, err := store.Count()
	assert.NoError(t, err)
	assert.Equal(t, 0, total)
	assert.Equal(t, 0, active)

	// Add servers with different active states
	activeServer := createTestServer("1", "active-server")
	activeServer.IsActive = true

	inactiveServer := createTestServer("2", "inactive-server")
	inactiveServer.IsActive = false

	err = store.Create(activeServer)
	require.NoError(t, err)

	err = store.Create(inactiveServer)
	require.NoError(t, err)

	// Test with servers
	total, active, err = store.Count()
	assert.NoError(t, err)
	assert.Equal(t, 2, total)
	assert.Equal(t, 1, active)
}

func TestMemoryStore_Concurrency(t *testing.T) {
	store := NewMemoryStore()
	numGoroutines := 10
	serversPerGoroutine := 10

	// Test concurrent writes
	done := make(chan bool, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(routineID int) {
			for j := 0; j < serversPerGoroutine; j++ {
				serverID := fmt.Sprintf("server-%d-%d", routineID, j)
				server := createTestServer(serverID, fmt.Sprintf("test-server-%d-%d", routineID, j))
				store.Create(server)
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < numGoroutines; i++ {
		<-done
	}

	// Verify all servers were created
	servers, err := store.GetAll()
	assert.NoError(t, err)
	assert.Len(t, servers, numGoroutines*serversPerGoroutine)
}

// Helper function to create test servers
func createTestServer(id, name string) models.Server {
	return models.Server{
		ID:          id,
		Name:        name,
		Description: "A test server",
		Version:     "1.0.0",
		Repository:  "https://github.com/test/" + name,
		Author:      "Test Author",
		Tags:        []string{"test"},
		IsActive:    true,
		CreatedAt:   time.Now().Format(time.RFC3339),
	}
}

// Benchmark tests
func BenchmarkMemoryStore_Create(b *testing.B) {
	store := NewMemoryStore()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		server := createTestServer(fmt.Sprintf("bench-%d", i), fmt.Sprintf("bench-server-%d", i))
		store.Create(server)
	}
}

func BenchmarkMemoryStore_GetAll(b *testing.B) {
	store := NewMemoryStore()

	// Setup: add some servers
	for i := 0; i < 1000; i++ {
		server := createTestServer(fmt.Sprintf("bench-%d", i), fmt.Sprintf("bench-server-%d", i))
		store.Create(server)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		store.GetAll()
	}
}
