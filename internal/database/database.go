package database

import (
	"context"
	"errors"
	"registry/internal/model"
)

// Common database errors
var (
	ErrNotFound       = errors.New("record not found")
	ErrAlreadyExists  = errors.New("record already exists")
	ErrInvalidInput   = errors.New("invalid input")
	ErrDatabase       = errors.New("database error")
	ErrInvalidVersion = errors.New("invalid version: cannot publish older version after newer version")
)

// Database defines the interface for database operations on MCPRegistry entries
type Database interface {
	// List retrieves all MCPRegistry entries with optional filtering
	List(ctx context.Context, filter map[string]interface{}, cursor string, limit int) ([]*model.Server, string, error)
	// GetByID retrieves a single ServerDetail by it's ID
	GetByID(ctx context.Context, id string) (*model.ServerDetail, error)
	// Publish adds a new ServerDetail to the database
	Publish(ctx context.Context, serverDetail *model.ServerDetail) error
	// Close closes the database connection
	Close() error
}
