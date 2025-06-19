package models

type Server struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Version     string   `json:"version"`
	Repository  string   `json:"repository"`
	Author      string   `json:"author"`
	Tags        []string `json:"tags"`
	IsActive    bool     `json:"is_active"`
	CreatedAt   string   `json:"created_at"`
}

// ServerStore defines the interface for server storage operations
// This is a contract that any storage implementation must fulfill
type ServerStore interface {
	// GetAll returns all servers in the registry
	GetAll() ([]Server, error)

	// GetByID returns a specific server by its ID
	GetByID(id string) (*Server, error)

	// Create adds a new server to the registry
	Create(server Server) error

	// Update modifies an existing server
	Update(server Server) error

	// Delete removes a server from the registry
	Delete(id string) error

	// Search finds servers by name (case-insensitive)
	Search(nameQuery string) ([]Server, error)

	// Count returns the total number of servers and active servers
	Count() (total int, active int, err error)
}

// ValidationError represents a validation error with details
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func (e *ValidationError) Error() string {
	return e.Field + ": " + e.Message
}

type ValidationErrors []ValidationError

func (ve ValidationErrors) Error() string {
	if len(ve) == 0 {
		return "no validation errors"
	}

	result := "validation failed: "
	for i, err := range ve {
		if i > 0 {
			result += ", "
		}
		result += err.Error()
	}
	return result
}

// ValidateServer validates a server struct and returns any validation errors
func ValidateServer(server Server) error {
	var errors ValidationErrors

	if server.ID == "" {
		errors = append(errors, ValidationError{
			Field:   "id",
			Message: "is required",
		})
	}

	if server.Name == "" {
		errors = append(errors, ValidationError{
			Field:   "name",
			Message: "is required",
		})
	}

	if server.Version == "" {
		errors = append(errors, ValidationError{
			Field:   "version",
			Message: "is required",
		})
	}

	if len(errors) > 0 {
		return errors
	}

	return nil
}
