// This file implements a simple HTTP server with JSON endpoints

package main

import (
	"fmt"
	"log"
	"net/http"
	"registry/internal/handlers"
	"registry/internal/storage"
)

func main() {
	fmt.Println("🚀 MCP Registry server starting...")

	// Initialize storage
	store := storage.NewMemoryStore()

	// Add sample data
	if err := store.InitWithSampleData(); err != nil {
		log.Fatalf("Failed to initialize sample data: %v", err)
	}

	// Initialize handlers
	handler := handlers.NewHandler(store)

	// Register routes
	http.HandleFunc("/health", handler.HealthHandler)
	http.HandleFunc("/servers", handler.ServersHandler)
	http.HandleFunc("/servers/", handler.ServerDetailHandler)
	http.HandleFunc("/servers/count", handler.CountHandler)
	http.HandleFunc("/servers/search", handler.SearchHandler)

	// Start server
	port := ":8080"
	fmt.Printf("📡 Server running on http://localhost%s\n", port)
	fmt.Println("📋 Available endpoints:")
	fmt.Println("   GET  /health")
	fmt.Println("   GET  /servers")
	fmt.Println("   POST /servers")
	fmt.Println("   GET  /servers/{id}")
	fmt.Println("   GET  /servers/count")
	fmt.Println("   GET  /servers/search?name=xyz")

	log.Fatal(http.ListenAndServe(port, nil))
}
