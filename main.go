// This file implements a simple HTTP server with JSON endpoints

package main

import (
	"log"
	"registry/internal/config"
	"registry/internal/models"
	"registry/internal/server"
	"registry/internal/storage"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Log configuration (excluding sensitive data)
	cfg.LogConfig()

	// Initialize storage based on configuration
	var store models.ServerStore
	switch cfg.StorageType {
	case "memory":
		memStore := storage.NewMemoryStore()
		if err := memStore.InitWithSampleData(); err != nil {
			log.Fatalf("Failed to initialize sample data: %v", err)
		}
		store = memStore
	default:
		log.Fatalf("Unknown storage type: %s", cfg.StorageType)
	}

	// Create and start server
	srv := server.New(cfg, store)

	// Start server (blocks until shutdown)
	if err := srv.Start(); err != nil {
		log.Fatalf("Server failed: %v", err)
	}

	log.Println("👋 Goodbye!")
}
