package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"registry/internal/config"
	"registry/internal/handlers"
	"registry/internal/models"
	"syscall"
	"time"
)

// Server wraps the HTTP server with lifecycle management
type Server struct {
	config     *config.Config
	httpServer *http.Server
	store      models.ServerStore
	handler    *handlers.Handler
}

// New creates a new server instance
func New(cfg *config.Config, store models.ServerStore) *Server {
	// Create handler
	handler := handlers.NewHandler(store)

	// Create HTTP server
	httpServer := &http.Server{
		Addr:           cfg.Address,
		Handler:        setupRoutes(handler, cfg),
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		IdleTimeout:    15 * time.Second,
		MaxHeaderBytes: 1 << 20, // 1 MB
	}

	return &Server{
		config:     cfg,
		httpServer: httpServer,
		store:      store,
		handler:    handler,
	}
}

// setupRoutes configures all HTTP routes
func setupRoutes(handler *handlers.Handler, cfg *config.Config) http.Handler {
	mux := http.NewServeMux()

	// API routes
	mux.HandleFunc("/health", handler.HealthHandler)
	mux.HandleFunc("/servers", handler.ServersHandler)
	mux.HandleFunc("/servers/", handler.ServerDetailHandler)
	mux.HandleFunc("/servers/count", handler.CountHandler)
	mux.HandleFunc("/servers/search", handler.SearchHandler)

	// Development routes (only in dev environment)
	if cfg.IsDevelopment() {
		mux.HandleFunc("/debug/config", debugConfigHandler(cfg))
	}

	// Add middleware
	var finalHandler http.Handler = mux

	// CORS middleware (if enabled)
	if cfg.EnableCORS {
		finalHandler = corsMiddleware(finalHandler)
	}

	// Logging middleware
	finalHandler = loggingMiddleware(finalHandler, cfg)

	return finalHandler
}

// Start begins serving HTTP requests
func (s *Server) Start() error {
	log.Printf("🚀 Starting MCP Registry server")
	log.Printf("📡 Server listening on http://%s", s.config.Address)
	log.Printf("🌍 Environment: %s", s.config.Environment)

	// Print available endpoints
	s.printEndpoints()

	// Start server in a goroutine
	serverErrors := make(chan error, 1)
	go func() {
		serverErrors <- s.httpServer.ListenAndServe()
	}()

	// Setup signal handling for graceful shutdown
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

	// Block until we receive a signal or server error
	select {
	case err := <-serverErrors:
		if err != nil && err != http.ErrServerClosed {
			return fmt.Errorf("server failed to start: %w", err)
		}
	case sig := <-shutdown:
		log.Printf("🛑 Received shutdown signal: %v", sig)

		// Create context with timeout for graceful shutdown
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		// Shutdown server gracefully
		log.Println("🔄 Shutting down server gracefully...")
		if err := s.httpServer.Shutdown(ctx); err != nil {
			log.Printf("❌ Server shutdown error: %v", err)
			return err
		}

		log.Println("✅ Server shutdown complete")
	}

	return nil
}

// Stop shuts down the server gracefully
func (s *Server) Stop(ctx context.Context) error {
	log.Println("🔄 Stopping server...")
	return s.httpServer.Shutdown(ctx)
}

func (s *Server) printEndpoints() {
	baseURL := fmt.Sprintf("http://%s", s.config.Address)

	log.Println("📋 Available endpoints:")
	log.Printf("   GET  %s/health", baseURL)
	log.Printf("   GET  %s/servers", baseURL)
	log.Printf("   POST %s/servers", baseURL)
	log.Printf("   GET  %s/servers/{id}", baseURL)
	log.Printf("   GET  %s/servers/count", baseURL)
	log.Printf("   GET  %s/servers/search?name=xyz", baseURL)

	if s.config.IsDevelopment() {
		log.Printf("   GET  %s/debug/config (dev only)", baseURL)
	}

	log.Println("💡 Try these commands:")
	log.Printf("   curl %s/health", baseURL)
	log.Printf("   curl %s/servers", baseURL)
}

// Debug handlers

// debugConfigHandler returns current configuration (development only)
func debugConfigHandler(cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		w.Header().Set("Content-Type", "application/json")

		// Create safe config copy (remove sensitive data)
		safeConfig := map[string]interface{}{
			"environment":    cfg.Environment,
			"address":        cfg.Address,
			"log_level":      cfg.LogLevel,
			"storage_type":   cfg.StorageType,
			"enable_cors":    cfg.EnableCORS,
			"enable_metrics": cfg.EnableMetrics,
			"version":        cfg.Version,
		}

		if err := json.NewEncoder(w).Encode(safeConfig); err != nil {
			http.Error(w, "Failed to encode config", http.StatusInternalServerError)
		}
	}
}

// Middleware functions

// loggingMiddleware logs HTTP requests
func loggingMiddleware(next http.Handler, cfg *config.Config) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Create a response writer wrapper to capture status code
		ww := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		// Call the next handler
		next.ServeHTTP(ww, r)

		// Log the request (structured logging)
		duration := time.Since(start)

		if cfg.LogLevel == "debug" {
			log.Printf("HTTP %s %s %d %v %s",
				r.Method, r.URL.Path, ww.statusCode, duration, r.RemoteAddr)
		} else if ww.statusCode >= 400 {
			// Always log errors
			log.Printf("HTTP ERROR %s %s %d %v",
				r.Method, r.URL.Path, ww.statusCode, duration)
		}
	})
}

// corsMiddleware adds CORS headers
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		// Handle preflight requests
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}
