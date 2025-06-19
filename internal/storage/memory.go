package storage

import (
	"errors"
	"strings"
	"sync"
	"time"

	"registry/internal/models"
)

var (
	ErrServerNotFound = errors.New("server not found")
	ErrServerExists   = errors.New("server already exists")
	ErrInvalidID      = errors.New("invalid server ID")
)

// MemoryStore implements the ServerStore interface using in-memory storage
// This implementation is thread-safe using a read-write mutex
type MemoryStore struct {
	servers map[string]models.Server
	mu      sync.RWMutex // Allows multiple readers OR one writer
}

// NewMemoryStore creates a new in-memory storage instance
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		servers: make(map[string]models.Server),
	}
}

// GetAll returns all servers in the registry
func (m *MemoryStore) GetAll() ([]models.Server, error) {
	m.mu.RLock()         // Acquire read lock
	defer m.mu.RUnlock() // Release when function returns

	// Convert map to slice
	servers := make([]models.Server, 0, len(m.servers))
	for _, server := range m.servers {
		servers = append(servers, server)
	}

	return servers, nil
}

// GetByID returns a specific server by its ID
func (m *MemoryStore) GetByID(id string) (*models.Server, error) {
	if id == "" {
		return nil, ErrInvalidID
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	server, exists := m.servers[id]
	if !exists {
		return nil, ErrServerNotFound
	}

	// Return a copy to prevent external modification
	serverCopy := server
	return &serverCopy, nil
}

// Create adds a new server to the registry
func (m *MemoryStore) Create(server models.Server) error {
	// Validate the server first
	if err := models.ValidateServer(server); err != nil {
		return err
	}

	m.mu.Lock() // Acquire write lock (exclusive)
	defer m.mu.Unlock()

	// Check if server already exists
	if _, exists := m.servers[server.ID]; exists {
		return ErrServerExists
	}

	// Set creation time if not provided
	if server.CreatedAt == "" {
		server.CreatedAt = time.Now().Format(time.RFC3339)
	}

	// Store the server
	m.servers[server.ID] = server
	return nil
}

// Update modifies an existing server
func (m *MemoryStore) Update(server models.Server) error {
	// Validate the server first
	if err := models.ValidateServer(server); err != nil {
		return err
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	// Check if server exists
	if _, exists := m.servers[server.ID]; !exists {
		return ErrServerNotFound
	}

	// Update the server (preserve original creation time)
	existing := m.servers[server.ID]
	server.CreatedAt = existing.CreatedAt
	m.servers[server.ID] = server

	return nil
}

// Delete removes a server from the registry
func (m *MemoryStore) Delete(id string) error {
	if id == "" {
		return ErrInvalidID
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	// Check if server exists
	if _, exists := m.servers[id]; !exists {
		return ErrServerNotFound
	}

	delete(m.servers, id)
	return nil
}

func (m *MemoryStore) Search(nameQuery string) ([]models.Server, error) {
	if nameQuery == "" {
		return []models.Server{}, nil
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	var results []models.Server
	queryLower := strings.ToLower(nameQuery)

	for _, server := range m.servers {
		if strings.Contains(strings.ToLower(server.Name), queryLower) {
			results = append(results, server)
		}
	}

	return results, nil
}

// Count returns the total number of servers and active servers
func (m *MemoryStore) Count() (total int, active int, err error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	total = len(m.servers)
	for _, server := range m.servers {
		if server.IsActive {
			active++
		}
	}

	return total, active, nil
}

// InitWithSampleData populates the store with sample data for testing
func (m *MemoryStore) InitWithSampleData() error {
	sampleServers := []models.Server{
		{
			ID:          "1",
			Name:        "filesystem-server",
			Description: "A server for accessing local filesystem",
			Version:     "1.0.0",
			Repository:  "https://github.com/example/filesystem-server",
			Author:      "Jane Doe",
			Tags:        []string{"filesystem", "local", "files"},
			IsActive:    true,
			CreatedAt:   time.Now().Format(time.RFC3339),
		},
		{
			ID:          "2",
			Name:        "web-scraper-server",
			Description: "A server for web scraping operations",
			Version:     "2.1.0",
			Repository:  "https://github.com/example/web-scraper",
			Author:      "John Smith",
			Tags:        []string{"web", "scraping", "http"},
			IsActive:    true,
			CreatedAt:   time.Now().Format(time.RFC3339),
		},
		{
			ID:          "3",
			Name:        "database-server",
			Description: "A server for database operations",
			Version:     "1.5.2",
			Repository:  "https://github.com/example/database-server",
			Author:      "Alice Johnson",
			Tags:        []string{"database", "sql", "storage"},
			IsActive:    false,
			CreatedAt:   time.Now().Format(time.RFC3339),
		},
	}

	for _, server := range sampleServers {
		if err := m.Create(server); err != nil {
			return err
		}
	}

	return nil
}
