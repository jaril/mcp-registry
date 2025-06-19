package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"registry/internal/models"
	"registry/internal/storage"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandler_HealthHandler(t *testing.T) {
	// Create handler with mock store
	mockStore := storage.NewMockStore()
	handler := NewHandler(mockStore)

	tests := []struct {
		name           string
		method         string
		expectedStatus int
		expectedBody   map[string]string
	}{
		{
			name:           "successful health check",
			method:         http.MethodGet,
			expectedStatus: http.StatusOK,
			expectedBody:   map[string]string{"status": "ok"},
		},
		{
			name:           "method not allowed",
			method:         http.MethodPost,
			expectedStatus: http.StatusMethodNotAllowed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, "/health", nil)
			w := httptest.NewRecorder()

			handler.HealthHandler(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedBody != nil {
				var response map[string]string
				err := json.NewDecoder(w.Body).Decode(&response)
				require.NoError(t, err)
				assert.Equal(t, tt.expectedBody, response)
			}
		})
	}
}

func TestHandler_ServersHandler_GET(t *testing.T) {
	// Setup test data
	testServers := []models.Server{
		{
			ID:          "1",
			Name:        "test-server-1",
			Description: "First test server",
			Version:     "1.0.0",
		},
		{
			ID:          "2",
			Name:        "test-server-2",
			Description: "Second test server",
			Version:     "2.0.0",
		},
	}

	tests := []struct {
		name           string
		setupMock      func(*storage.MockStore)
		expectedStatus int
		expectedCount  int
	}{
		{
			name: "successful get all servers",
			setupMock: func(mock *storage.MockStore) {
				mock.ExpectGetAll(testServers, nil)
			},
			expectedStatus: http.StatusOK,
			expectedCount:  2,
		},
		{
			name: "storage error",
			setupMock: func(mock *storage.MockStore) {
				mock.ExpectGetAll(nil, assert.AnError)
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStore := storage.NewMockStore()
			tt.setupMock(mockStore)

			handler := NewHandler(mockStore)
			req := httptest.NewRequest(http.MethodGet, "/servers", nil)
			w := httptest.NewRecorder()

			handler.ServersHandler(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedStatus == http.StatusOK {
				var response map[string]interface{}
				err := json.NewDecoder(w.Body).Decode(&response)
				require.NoError(t, err)

				servers, ok := response["servers"].([]interface{})
				require.True(t, ok)
				assert.Len(t, servers, tt.expectedCount)
			}

			mockStore.AssertExpectations(t)
		})
	}
}

func TestHandler_ServersHandler_POST(t *testing.T) {
	validServer := models.Server{
		ID:          "test-create",
		Name:        "new-server",
		Description: "A newly created server",
		Version:     "1.0.0",
		Repository:  "https://github.com/test/new-server",
		Author:      "Test Author",
		Tags:        []string{"test"},
		IsActive:    true,
	}

	tests := []struct {
		name           string
		requestBody    interface{}
		setupMock      func(*storage.MockStore)
		expectedStatus int
	}{
		{
			name:        "successful create",
			requestBody: validServer,
			setupMock: func(mock *storage.MockStore) {
				mock.ExpectCreate(validServer, nil)
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "invalid JSON",
			requestBody:    "invalid json",
			setupMock:      func(mock *storage.MockStore) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "validation error",
			requestBody: models.Server{
				Name: "invalid-server", // Missing required fields
			},
			setupMock: func(mock *storage.MockStore) {
				validationErr := models.ValidationErrors{
					{Field: "id", Message: "is required"},
					{Field: "version", Message: "is required"},
				}
				mock.ExpectCreate(mock.MatchedBy(func(s models.Server) bool {
					return s.Name == "invalid-server"
				}), validationErr)
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:        "server already exists",
			requestBody: validServer,
			setupMock: func(mock *storage.MockStore) {
				mock.ExpectCreate(validServer, storage.ErrServerExists)
			},
			expectedStatus: http.StatusConflict,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStore := storage.NewMockStore()
			tt.setupMock(mockStore)

			handler := NewHandler(mockStore)

			// Create request body
			var body []byte
			var err error
			if str, ok := tt.requestBody.(string); ok {
				body = []byte(str)
			} else {
				body, err = json.Marshal(tt.requestBody)
				require.NoError(t, err)
			}

			req := httptest.NewRequest(http.MethodPost, "/servers", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			handler.ServersHandler(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			mockStore.AssertExpectations(t)
		})
	}
}

func TestHandler_ServerDetailHandler(t *testing.T) {
	testServer := &models.Server{
		ID:          "test-1",
		Name:        "test-server",
		Description: "A test server",
		Version:     "1.0.0",
	}

	tests := []struct {
		name           string
		method         string
		serverID       string
		setupMock      func(*storage.MockStore)
		expectedStatus int
	}{
		{
			name:     "successful get server",
			method:   http.MethodGet,
			serverID: "test-1",
			setupMock: func(mock *storage.MockStore) {
				mock.ExpectGetByID("test-1", testServer, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:     "server not found",
			method:   http.MethodGet,
			serverID: "non-existent",
			setupMock: func(mock *storage.MockStore) {
				mock.ExpectGetByID("non-existent", nil, storage.ErrServerNotFound)
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:     "empty server ID",
			method:   http.MethodGet,
			serverID: "",
			setupMock: func(mock *storage.MockStore) {
				// No mock setup needed - validation happens before storage call
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "method not allowed",
			method:         http.MethodPost,
			serverID:       "test-1",
			setupMock:      func(mock *storage.MockStore) {},
			expectedStatus: http.StatusMethodNotAllowed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStore := storage.NewMockStore()
			tt.setupMock(mockStore)

			handler := NewHandler(mockStore)

			url := "/servers/"
			if tt.serverID != "" {
				url += tt.serverID
			}

			req := httptest.NewRequest(tt.method, url, nil)
			w := httptest.NewRecorder()

			handler.ServerDetailHandler(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedStatus == http.StatusOK {
				var response models.Server
				err := json.NewDecoder(w.Body).Decode(&response)
				require.NoError(t, err)
				assert.Equal(t, testServer.ID, response.ID)
				assert.Equal(t, testServer.Name, response.Name)
			}

			mockStore.AssertExpectations(t)
		})
	}
}

func TestHandler_CountHandler(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		setupMock      func(*storage.MockStore)
		expectedStatus int
		expectedTotal  int
		expectedActive int
	}{
		{
			name:   "successful count",
			method: http.MethodGet,
			setupMock: func(mock *storage.MockStore) {
				mock.ExpectCount(5, 3, nil)
			},
			expectedStatus: http.StatusOK,
			expectedTotal:  5,
			expectedActive: 3,
		},
		{
			name:   "storage error",
			method: http.MethodGet,
			setupMock: func(mock *storage.MockStore) {
				mock.ExpectCount(0, 0, assert.AnError)
			},
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:           "method not allowed",
			method:         http.MethodPost,
			setupMock:      func(mock *storage.MockStore) {},
			expectedStatus: http.StatusMethodNotAllowed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStore := storage.NewMockStore()
			tt.setupMock(mockStore)

			handler := NewHandler(mockStore)
			req := httptest.NewRequest(tt.method, "/servers/count", nil)
			w := httptest.NewRecorder()

			handler.CountHandler(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedStatus == http.StatusOK {
				var response map[string]int
				err := json.NewDecoder(w.Body).Decode(&response)
				require.NoError(t, err)
				assert.Equal(t, tt.expectedTotal, response["total"])
				assert.Equal(t, tt.expectedActive, response["active"])
			}

			mockStore.AssertExpectations(t)
		})
	}
}

func TestHandler_SearchHandler(t *testing.T) {
	testServers := []models.Server{
		{
			ID:   "1",
			Name: "filesystem-server",
		},
		{
			ID:   "2",
			Name: "file-processor",
		},
	}

	tests := []struct {
		name           string
		method         string
		query          string
		setupMock      func(*storage.MockStore)
		expectedStatus int
		expectedCount  int
	}{
		{
			name:   "successful search",
			method: http.MethodGet,
			query:  "file",
			setupMock: func(mock *storage.MockStore) {
				mock.ExpectSearch("file", testServers, nil)
			},
			expectedStatus: http.StatusOK,
			expectedCount:  2,
		},
		{
			name:           "missing search term",
			method:         http.MethodGet,
			query:          "",
			setupMock:      func(mock *storage.MockStore) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:   "storage error",
			method: http.MethodGet,
			query:  "test",
			setupMock: func(mock *storage.MockStore) {
				mock.ExpectSearch("test", nil, assert.AnError)
			},
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:           "method not allowed",
			method:         http.MethodPost,
			query:          "test",
			setupMock:      func(mock *storage.MockStore) {},
			expectedStatus: http.StatusMethodNotAllowed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStore := storage.NewMockStore()
			tt.setupMock(mockStore)

			handler := NewHandler(mockStore)

			url := "/servers/search"
			if tt.query != "" {
				url += "?name=" + tt.query
			}

			req := httptest.NewRequest(tt.method, url, nil)
			w := httptest.NewRecorder()

			handler.SearchHandler(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedStatus == http.StatusOK {
				var response map[string]interface{}
				err := json.NewDecoder(w.Body).Decode(&response)
				require.NoError(t, err)

				servers, ok := response["servers"].([]interface{})
				require.True(t, ok)
				assert.Len(t, servers, tt.expectedCount)

				assert.Equal(t, tt.query, response["search_term"])
				assert.Equal(t, float64(tt.expectedCount), response["count"])
			}

			mockStore.AssertExpectations(t)
		})
	}
}

// Integration test that tests the full HTTP stack
func TestHandler_Integration(t *testing.T) {
	// Use real memory store for integration test
	store := storage.NewMemoryStore()
	handler := NewHandler(store)

	// Test complete flow: create -> get -> list -> search
	t.Run("complete server lifecycle", func(t *testing.T) {
		// 1. Create a server
		server := models.Server{
			ID:          "integration-test",
			Name:        "integration-server",
			Description: "Server for integration testing",
			Version:     "1.0.0",
			Repository:  "https://github.com/test/integration",
			Author:      "Test Author",
			Tags:        []string{"integration", "test"},
			IsActive:    true,
		}

		// POST /servers
		serverJSON, err := json.Marshal(server)
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodPost, "/servers", bytes.NewBuffer(serverJSON))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.ServersHandler(w, req)
		assert.Equal(t, http.StatusCreated, w.Code)

		// 2. Get the server by ID
		req = httptest.NewRequest(http.MethodGet, "/servers/integration-test", nil)
		w = httptest.NewRecorder()

		handler.ServerDetailHandler(w, req)
		assert.Equal(t, http.StatusOK, w.Code)

		var retrievedServer models.Server
		err = json.NewDecoder(w.Body).Decode(&retrievedServer)
		require.NoError(t, err)
		assert.Equal(t, server.Name, retrievedServer.Name)

		// 3. List all servers (should include our new one)
		req = httptest.NewRequest(http.MethodGet, "/servers", nil)
		w = httptest.NewRecorder()

		handler.ServersHandler(w, req)
		assert.Equal(t, http.StatusOK, w.Code)

		var listResponse map[string]interface{}
		err = json.NewDecoder(w.Body).Decode(&listResponse)
		require.NoError(t, err)

		servers, ok := listResponse["servers"].([]interface{})
		require.True(t, ok)
		assert.GreaterOrEqual(t, len(servers), 1)

		// 4. Search for the server
		req = httptest.NewRequest(http.MethodGet, "/servers/search?name=integration", nil)
		w = httptest.NewRecorder()

		handler.SearchHandler(w, req)
		assert.Equal(t, http.StatusOK, w.Code)

		var searchResponse map[string]interface{}
		err = json.NewDecoder(w.Body).Decode(&searchResponse)
		require.NoError(t, err)

		searchResults, ok := searchResponse["servers"].([]interface{})
		require.True(t, ok)
		assert.Len(t, searchResults, 1)

		// 5. Get count (should include our server)
		req = httptest.NewRequest(http.MethodGet, "/servers/count", nil)
		w = httptest.NewRecorder()

		handler.CountHandler(w, req)
		assert.Equal(t, http.StatusOK, w.Code)

		var countResponse map[string]interface{}
		err = json.NewDecoder(w.Body).Decode(&countResponse)
		require.NoError(t, err)

		total, ok := countResponse["total"].(float64)
		require.True(t, ok)
		assert.GreaterOrEqual(t, int(total), 1)

		active, ok := countResponse["active"].(float64)
		require.True(t, ok)
		assert.GreaterOrEqual(t, int(active), 1)
	})
}

// Benchmark tests for handlers
func BenchmarkHandler_GetServers(b *testing.B) {
	// Setup
	store := storage.NewMemoryStore()
	handler := NewHandler(store)

	// Add some test data
	for i := 0; i < 100; i++ {
		server := models.Server{
			ID:          fmt.Sprintf("bench-%d", i),
			Name:        fmt.Sprintf("bench-server-%d", i),
			Description: "Benchmark server",
			Version:     "1.0.0",
			Repository:  fmt.Sprintf("https://github.com/test/bench-%d", i),
			Author:      "Bench Author",
			Tags:        []string{"benchmark"},
			IsActive:    true,
		}
		store.Create(server)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodGet, "/servers", nil)
		w := httptest.NewRecorder()
		handler.ServersHandler(w, req)
	}
}
