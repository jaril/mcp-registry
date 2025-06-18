// This file implements a simple HTTP server with JSON endpoints

package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"
)

// Server represents an MCP server in our registry
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

// In-memory storage for our servers
var servers []Server

// Initialize some sample data
func initSampleData() {
	servers = []Server{
		{
			ID:          "1",
			Name:        "filesystem-server",
			Description: "A server for accessing local filesystem",
			Version:     "1.0.0",
			Repository:  "https://github.com/example/filesystem-server",
			Tags:        []string{"filesystem", "local", "files"},
			IsActive:    true,
			CreatedAt:   time.Now().Format(time.RFC3339),
			Author:      "John Doe",
		},
		{
			ID:          "2",
			Name:        "web-scraper-server",
			Description: "A server for web scraping operations",
			Version:     "2.1.0",
			Repository:  "https://github.com/example/web-scraper",
			Tags:        []string{"web", "scraping", "http"},
			IsActive:    true,
			CreatedAt:   time.Now().Format(time.RFC3339),
			Author:      "Jane Doe",
		},
		{
			ID:          "3",
			Name:        "database-server",
			Description: "A server for database operations",
			Version:     "1.5.2",
			Repository:  "https://github.com/example/database-server",
			Tags:        []string{"database", "sql", "postgres"},
			IsActive:    true,
			CreatedAt:   time.Now().Format(time.RFC3339),
			Author:      "John Doe",
		},
		{
			ID:          "4",
			Name:        "email-server",
			Description: "A server for email operations",
			Version:     "0.9.0",
			Repository:  "https://github.com/example/email-server",
			Tags:        []string{"email", "smtp", "imap"},
			IsActive:    true,
			CreatedAt:   time.Now().Format(time.RFC3339),
			Author:      "Jane Doe",
		},
	}
}

// healthHandler handles GET /health requests
// This is a simple endpoint to check if our server is running
func healthHandler(w http.ResponseWriter, r *http.Request) {
	// Only allow GET requests
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Create a simple response
	response := map[string]string{
		"status": "ok",
		"time":   time.Now().Format(time.RFC3339),
	}

	// Set the content type to JSON
	w.Header().Set("Content-Type", "application/json")

	// Encode the response as JSON and send it
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// serversHandler handles GET /servers requests
// This returns a list of all servers in our registry
func serversHandler(w http.ResponseWriter, r *http.Request) {
	// Only allow GET requests for now
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Create the response structure
	response := map[string]interface{}{
		"servers": servers,
		"count":   len(servers),
	}

	// Set content type and send JSON response
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

func countHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Count total servers and active servers
	totalCount := len(servers)
	activeCount := 0

	for _, server := range servers {
		if server.IsActive {
			activeCount++
		}
	}

	response := map[string]int{
		"total":  totalCount,
		"active": activeCount,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// searchHandler handles GET /servers/search?name=xyz requests
func searchHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get the search term from query parameters
	searchTerm := r.URL.Query().Get("name")
	if searchTerm == "" {
		http.Error(w, "Search term 'name' is required", http.StatusBadRequest)
		return
	}

	// Search for servers containing the search term
	var matchingServers []Server
	for _, server := range servers {
		// Check if the server name contains the search term (case-insensitive)
		if strings.Contains(strings.ToLower(server.Name), strings.ToLower(searchTerm)) {
			matchingServers = append(matchingServers, server)
		}
	}

	response := map[string]interface{}{
		"servers":     matchingServers,
		"count":       len(matchingServers),
		"search_term": searchTerm,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// serverDetailHandler handles GET /servers/{id} requests
// This returns details for a specific server
func serverDetailHandler(w http.ResponseWriter, r *http.Request) {
	// Only allow GET requests
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract the server ID from the URL path
	// URL will be like "/servers/1", so we remove "/servers/" to get the ID
	path := r.URL.Path
	if !strings.HasPrefix(path, "/servers/") {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}

	id := strings.TrimPrefix(path, "/servers/")
	if id == "" {
		http.Error(w, "Server ID is required", http.StatusBadRequest)
		return
	}

	// Find the server with the matching ID
	for _, server := range servers {
		if server.ID == id {
			// Found the server, return it
			w.Header().Set("Content-Type", "application/json")
			if err := json.NewEncoder(w).Encode(server); err != nil {
				http.Error(w, "Failed to encode response", http.StatusInternalServerError)
				return
			}
			return
		}
	}

	// Server not found
	http.Error(w, "Server not found", http.StatusNotFound)
}

func main() {
	// Initialize our sample data
	initSampleData()

	// Register our HTTP handlers
	// These map URL patterns to handler functions
	http.HandleFunc("/health", healthHandler)
	http.HandleFunc("/servers/", serverDetailHandler)
	http.HandleFunc("/servers/search", searchHandler)
	http.HandleFunc("/servers", serversHandler)
	http.HandleFunc("/count", countHandler)

	// Start the server
	port := ":8080"
	fmt.Printf("🚀 MCP Registry server starting on http://localhost%s\n", port)
	fmt.Println("📡 Available endpoints:")
	fmt.Println("   GET /health - Check server status")
	fmt.Println("   GET /servers - List all servers")
	fmt.Println("   GET /servers/search?name=<term> - Search servers by name")
	fmt.Println("   GET /servers/{id} - Get server details")
	fmt.Println("   GET /count - Get server counts")
	fmt.Println("")
	fmt.Println("💡 Try these commands:")
	fmt.Println("   curl http://localhost:8080/health")
	fmt.Println("   curl http://localhost:8080/servers")
	fmt.Println("   curl http://localhost:8080/servers/1")
	fmt.Println("   curl 'http://localhost:8080/servers/search?name=filesystem'")

	// ListenAndServe starts the HTTP server and blocks
	// If there's an error starting the server, log.Fatal will exit the program
	log.Fatal(http.ListenAndServe(port, nil))
}
