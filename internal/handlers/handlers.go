// Package handlers provides HTTP request handlers for the MCP registry API
package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"registry/internal/models"
)

// Handler contains the dependencies needed for handling HTTP requests
type Handler struct {
	store models.ServerStore // Interface, not concrete type!
}

// NewHandler creates a new handler instance with the given store
func NewHandler(store models.ServerStore) *Handler {
	return &Handler{
		store: store,
	}
}

// HealthHandler handles GET /health requests
func (h *Handler) HealthHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	response := map[string]string{
		"status": "ok",
	}

	h.writeJSON(w, response, http.StatusOK)
}

// ServersHandler handles GET /servers and POST /servers requests
func (h *Handler) ServersHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.handleGetServers(w, r)
	case http.MethodPost:
		h.handleCreateServer(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleGetServers handles GET /servers requests
func (h *Handler) handleGetServers(w http.ResponseWriter, r *http.Request) {
	servers, err := h.store.GetAll()
	if err != nil {
		h.writeError(w, err, "Failed to retrieve servers", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"servers": servers,
		"count":   len(servers),
	}

	h.writeJSON(w, response, http.StatusOK)
}

// handleCreateServer handles POST /servers requests
func (h *Handler) handleCreateServer(w http.ResponseWriter, r *http.Request) {
	var server models.Server
	if err := json.NewDecoder(r.Body).Decode(&server); err != nil {
		h.writeError(w, err, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if err := h.store.Create(server); err != nil {
		// Check error type to determine response
		switch err.(type) {
		case models.ValidationErrors:
			h.writeError(w, err, "Validation failed", http.StatusBadRequest)
		default:
			if strings.Contains(err.Error(), "already exists") {
				h.writeError(w, err, "Server already exists", http.StatusConflict)
			} else {
				h.writeError(w, err, "Failed to create server", http.StatusInternalServerError)
			}
		}
		return
	}

	h.writeJSON(w, server, http.StatusCreated)
}

// ServerDetailHandler handles GET /servers/{id} requests
func (h *Handler) ServerDetailHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract ID from URL path
	id := h.extractIDFromPath(r.URL.Path, "/servers/")
	if id == "" {
		http.Error(w, "Server ID is required", http.StatusBadRequest)
		return
	}

	server, err := h.store.GetByID(id)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			h.writeError(w, err, "Server not found", http.StatusNotFound)
		} else {
			h.writeError(w, err, "Failed to retrieve server", http.StatusInternalServerError)
		}
		return
	}

	h.writeJSON(w, server, http.StatusOK)
}

// CountHandler handles GET /servers/count requests
func (h *Handler) CountHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	total, active, err := h.store.Count()
	if err != nil {
		h.writeError(w, err, "Failed to get server count", http.StatusInternalServerError)
		return
	}

	response := map[string]int{
		"total":  total,
		"active": active,
	}

	h.writeJSON(w, response, http.StatusOK)
}

// SearchHandler handles GET /servers/search?name=xyz requests
func (h *Handler) SearchHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	nameQuery := r.URL.Query().Get("name")
	if nameQuery == "" {
		http.Error(w, "Search term 'name' is required", http.StatusBadRequest)
		return
	}

	servers, err := h.store.Search(nameQuery)
	if err != nil {
		h.writeError(w, err, "Failed to search servers", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"servers":     servers,
		"count":       len(servers),
		"search_term": nameQuery,
	}

	h.writeJSON(w, response, http.StatusOK)
}

// Helper methods

// extractIDFromPath extracts the ID from a URL path
func (h *Handler) extractIDFromPath(path, prefix string) string {
	if !strings.HasPrefix(path, prefix) {
		return ""
	}
	return strings.TrimPrefix(path, prefix)
}

// writeJSON writes a JSON response with the given status code
func (h *Handler) writeJSON(w http.ResponseWriter, data interface{}, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		// If we can't encode the response, write a simple error
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

// writeError writes an error response in JSON format
func (h *Handler) writeError(w http.ResponseWriter, err error, message string, statusCode int) {
	response := map[string]string{
		"error":   message,
		"details": err.Error(),
	}
	h.writeJSON(w, response, statusCode)
}
