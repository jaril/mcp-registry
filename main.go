package main

import (
	"log"
	"time"

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

	// Log configuration
	cfg.LogConfig()

	// Initialize storage based on configuration
	var store models.ServerStore
	var cleanup func() error

	switch cfg.StorageType {
	case "memory":
		log.Println("📦 Using in-memory storage")
		memStore := storage.NewMemoryStore()
		if err := memStore.InitWithSampleData(); err != nil {
			log.Fatalf("Failed to initialize sample data: %v", err)
		}
		store = memStore
		cleanup = func() error { return nil } // No cleanup needed for memory

	case "sqlite", "database":
		log.Printf("🗄️  Using SQLite database: %s", cfg.DatabaseURL)

		connLifetime := time.Duration(cfg.ConnMaxLifetime) * time.Minute
		sqliteStore, err := storage.NewSQLiteStore(
			cfg.DatabaseURL,
			cfg.MaxOpenConns,
			cfg.MaxIdleConns,
			connLifetime,
		)
		if err != nil {
			log.Fatalf("Failed to initialize SQLite storage: %v", err)
		}

		// Initialize with sample data if database is empty
		if err := sqliteStore.InitWithSampleData(); err != nil {
			log.Fatalf("Failed to initialize sample data: %v", err)
		}

		store = sqliteStore
		cleanup = sqliteStore.Close

	default:
		log.Fatalf("Unknown storage type: %s", cfg.StorageType)
	}

	// Ensure cleanup happens on exit
	defer func() {
		if err := cleanup(); err != nil {
			log.Printf("Error during cleanup: %v", err)
		}
	}()

	// Create and start server
	srv := server.New(cfg, store)

	// Start server (blocks until shutdown)
	if err := srv.Start(); err != nil {
		log.Fatalf("Server failed: %v", err)
	}

	log.Println("👋 Goodbye!")
}
