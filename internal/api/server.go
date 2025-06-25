package api

import (
	"context"
	"log"
	"net/http"
	"registry/internal/api/router"
	"time"
)

// Server represents the HTTP server
type Server struct {
	// config      *config.Config
	// registry service.RegistryService
	// authService auth.Service
	// router *http.ServeMux
	server *http.Server
}

const SERVER_ADDRESS = ":8080"

// NewServer creates a new HTTP server
// func NewServer(cfg *config.Config, registryService service.RegistryService, authService auth.Service) *Server {
func NewServer() *Server {
	mux := router.New()

	server := &Server{
		// config:      cfg,
		// registry:    registryService,
		// authService: authService,
		// router:      mux,
		server: &http.Server{
			// Addr:              cfg.ServerAddress,
			Addr:              SERVER_ADDRESS,
			Handler:           mux,
			ReadHeaderTimeout: 10 * time.Second,
		},
	}

	return server
}

// Start begins listening for incoming HTTP requests
func (s *Server) Start() error {
	// log.Printf("HTTP server starting on %s", s.config.ServerAddress)
	log.Printf("HTTP server starting on %s", SERVER_ADDRESS)
	return s.server.ListenAndServe()
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}
