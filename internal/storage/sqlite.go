// Package storage provides storage implementations for the MCP registry
package storage

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"registry/internal/models"

	_ "github.com/mattn/go-sqlite3" // SQLite driver
)

// SQLiteStore implements ServerStore using SQLite database
type SQLiteStore struct {
	db *sql.DB
}

// NewSQLiteStore creates a new SQLite storage instance
func NewSQLiteStore(databaseURL string, maxOpenConns, maxIdleConns int, connMaxLifetime time.Duration) (*SQLiteStore, error) {
	// Ensure data directory exists
	if err := ensureDataDir(databaseURL); err != nil {
		return nil, fmt.Errorf("failed to create data directory: %w", err)
	}

	// Open database connection
	db, err := sql.Open("sqlite3", databaseURL+"?_foreign_keys=on&_journal_mode=WAL")
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(maxOpenConns)
	db.SetMaxIdleConns(maxIdleConns)
	db.SetConnMaxLifetime(connMaxLifetime)

	// Test connection
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	store := &SQLiteStore{db: db}

	// Run migrations
	if err := store.runMigrations(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	return store, nil
}

// ensureDataDir creates the data directory if it doesn't exist
func ensureDataDir(databaseURL string) error {
	dir := filepath.Dir(databaseURL)
	if dir == "." {
		return nil // Current directory
	}
	return os.MkdirAll(dir, 0755)
}

// runMigrations applies database schema migrations
func (s *SQLiteStore) runMigrations() error {
	// Read migration file
	migrationSQL, err := os.ReadFile("internal/storage/migrations.sql")
	if err != nil {
		return fmt.Errorf("failed to read migrations.sql: %w", err)
	}

	// Execute migration
	_, err = s.db.Exec(string(migrationSQL))
	if err != nil {
		return fmt.Errorf("failed to execute migrations: %w", err)
	}

	return nil
}

// GetAll returns all servers from the database
func (s *SQLiteStore) GetAll() ([]models.Server, error) {
	query := `
		SELECT id, name, description, version, repository, author, tags, is_active, created_at
		FROM servers 
		ORDER BY created_at DESC
	`

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query servers: %w", err)
	}
	defer rows.Close()

	var servers []models.Server
	for rows.Next() {
		server, err := s.scanServer(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan server: %w", err)
		}
		servers = append(servers, *server)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return servers, nil
}

// GetByID returns a specific server by ID
func (s *SQLiteStore) GetByID(id string) (*models.Server, error) {
	if id == "" {
		return nil, ErrInvalidID
	}

	query := `
		SELECT id, name, description, version, repository, author, tags, is_active, created_at
		FROM servers 
		WHERE id = ?
	`

	row := s.db.QueryRow(query, id)
	server, err := s.scanServer(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrServerNotFound
		}
		return nil, fmt.Errorf("failed to get server by ID: %w", err)
	}

	return server, nil
}

// Create adds a new server to the database
func (s *SQLiteStore) Create(server models.Server) error {
	// Validate server
	if err := models.ValidateServer(server); err != nil {
		return err
	}

	// Check if server already exists
	if _, err := s.GetByID(server.ID); err == nil {
		return ErrServerExists
	} else if !errors.Is(err, ErrServerNotFound) {
		return fmt.Errorf("failed to check existing server: %w", err)
	}

	// Convert tags to JSON
	tagsJSON, err := json.Marshal(server.Tags)
	if err != nil {
		return fmt.Errorf("failed to marshal tags: %w", err)
	}

	// Set creation time if not provided
	if server.CreatedAt == "" {
		server.CreatedAt = time.Now().Format(time.RFC3339)
	}

	// Insert server
	query := `
		INSERT INTO servers (id, name, description, version, repository, author, tags, is_active, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err = s.db.Exec(query,
		server.ID, server.Name, server.Description, server.Version,
		server.Repository, server.Author, string(tagsJSON), server.IsActive, server.CreatedAt)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			return ErrServerExists
		}
		return fmt.Errorf("failed to insert server: %w", err)
	}

	return nil
}

// Update modifies an existing server
func (s *SQLiteStore) Update(server models.Server) error {
	// Validate server
	if err := models.ValidateServer(server); err != nil {
		return err
	}

	// Check if server exists
	existing, err := s.GetByID(server.ID)
	if err != nil {
		return err
	}

	// Convert tags to JSON
	tagsJSON, err := json.Marshal(server.Tags)
	if err != nil {
		return fmt.Errorf("failed to marshal tags: %w", err)
	}

	// Preserve original creation time
	server.CreatedAt = existing.CreatedAt

	// Update server
	query := `
		UPDATE servers 
		SET name = ?, description = ?, version = ?, repository = ?, author = ?, 
			tags = ?, is_active = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`

	result, err := s.db.Exec(query,
		server.Name, server.Description, server.Version, server.Repository,
		server.Author, string(tagsJSON), server.IsActive, server.ID)
	if err != nil {
		return fmt.Errorf("failed to update server: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return ErrServerNotFound
	}

	return nil
}

// Delete removes a server from the database
func (s *SQLiteStore) Delete(id string) error {
	if id == "" {
		return ErrInvalidID
	}

	query := `DELETE FROM servers WHERE id = ?`

	result, err := s.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete server: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return ErrServerNotFound
	}

	return nil
}

// Search finds servers by name
func (s *SQLiteStore) Search(nameQuery string) ([]models.Server, error) {
	if nameQuery == "" {
		return []models.Server{}, nil
	}

	query := `
		SELECT id, name, description, version, repository, author, tags, is_active, created_at
		FROM servers 
		WHERE name LIKE ? 
		ORDER BY created_at DESC
	`

	// Use SQL LIKE with wildcards for case-insensitive search
	searchPattern := "%" + strings.ToLower(nameQuery) + "%"

	rows, err := s.db.Query(query, searchPattern)
	if err != nil {
		return nil, fmt.Errorf("failed to search servers: %w", err)
	}
	defer rows.Close()

	var servers []models.Server
	for rows.Next() {
		server, err := s.scanServer(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan server: %w", err)
		}
		servers = append(servers, *server)
	}

	return servers, nil
}

// Count returns server statistics
func (s *SQLiteStore) Count() (total int, active int, err error) {
	query := `
		SELECT 
			COUNT(*) as total,
			COALESCE(SUM(CASE WHEN is_active = 1 THEN 1 ELSE 0 END), 0) as active
		FROM servers
	`

	row := s.db.QueryRow(query)
	err = row.Scan(&total, &active)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to get server count: %w", err)
	}

	return total, active, nil
}

// Close closes the database connection
func (s *SQLiteStore) Close() error {
	if s.db != nil {
		return s.db.Close()
	}
	return nil
}

// Helper methods

// scanServer scans a database row into a Server struct
func (s *SQLiteStore) scanServer(row interface{ Scan(...interface{}) error }) (*models.Server, error) {
	var server models.Server
	var tagsJSON string

	err := row.Scan(
		&server.ID, &server.Name, &server.Description, &server.Version,
		&server.Repository, &server.Author, &tagsJSON, &server.IsActive, &server.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	// Parse tags JSON
	if tagsJSON != "" {
		if err := json.Unmarshal([]byte(tagsJSON), &server.Tags); err != nil {
			return nil, fmt.Errorf("failed to unmarshal tags: %w", err)
		}
	}

	return &server, nil
}

func (s *SQLiteStore) InitWithSampleData() error {
	// Check if we already have data
	count, _, err := s.Count()
	if err != nil {
		return fmt.Errorf("failed to check existing data: %w", err)
	}

	if count > 0 {
		// Data already exists, don't overwrite
		return nil
	}

	return s.loadFromJSONFile("data/seed_2025_05_16.json")
}

// loadFromJSONFile loads servers from a JSON file and converts them to the current Server model
func (s *SQLiteStore) loadFromJSONFile(filePath string) error {
	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		fmt.Printf("⚠️  Seed file not found: %s, using fallback sample data\n", filePath)
		return s.loadFallbackSampleData()
	}

	// Read JSON file
	data, err := os.ReadFile(filePath)
	if err != nil {
		fmt.Printf("⚠️  Failed to read seed file: %v, using fallback sample data\n", err)
		return s.loadFallbackSampleData()
	}

	// Define temporary struct matching your JSON structure
	type JSONServer struct {
		ID          string `json:"id"`
		Name        string `json:"name"`
		Description string `json:"description"`
		Repository  struct {
			URL    string `json:"url"`
			Source string `json:"source"`
			ID     string `json:"id"`
		} `json:"repository"`
		VersionDetail struct {
			Version     string `json:"version"`
			ReleaseDate string `json:"release_date"`
			IsLatest    bool   `json:"is_latest"`
		} `json:"version_detail"`
		Packages []struct {
			RegistryName string `json:"registry_name"`
			Name         string `json:"name"`
			Version      string `json:"version"`
		} `json:"packages,omitempty"`
	}

	// Parse JSON
	var jsonServers []JSONServer
	if err := json.Unmarshal(data, &jsonServers); err != nil {
		fmt.Printf("⚠️  Failed to parse JSON: %v, using fallback sample data\n", err)
		return s.loadFallbackSampleData()
	}

	// Convert to current Server model and create
	successCount := 0
	for _, jsonServer := range jsonServers {
		// Extract tags from package registry names
		tags := make([]string, 0)
		for _, pkg := range jsonServer.Packages {
			if pkg.RegistryName != "" && pkg.RegistryName != "unknown" {
				tags = append(tags, pkg.RegistryName)
			}
		}

		// Add some basic tags
		tags = append(tags, "mcp", "server")

		// Create Server with current model structure
		server := models.Server{
			ID:          jsonServer.ID,
			Name:        jsonServer.Name,
			Description: jsonServer.Description,
			Version:     jsonServer.VersionDetail.Version,
			Repository:  jsonServer.Repository.URL,
			Author:      extractAuthorFromRepoURL(jsonServer.Repository.URL),
			Tags:        tags,
			IsActive:    true, // Default to active
			CreatedAt:   time.Now().Format(time.RFC3339),
		}

		// Create the server
		if err := s.Create(server); err != nil {
			fmt.Printf("⚠️  Failed to create server %s: %v\n", server.Name, err)
		} else {
			successCount++
		}
	}

	fmt.Printf("✅ Loaded %d servers from %s\n", successCount, filePath)
	return nil
}

// loadFallbackSampleData provides fallback sample data if JSON loading fails
func (s *SQLiteStore) loadFallbackSampleData() error {
	// Sample servers (existing implementation)
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
		},
		// ... rest of sample servers
	}

	// Insert sample data
	for _, server := range sampleServers {
		if err := s.Create(server); err != nil {
			return fmt.Errorf("failed to create sample server %s: %w", server.ID, err)
		}
	}

	return nil
}
