package router

import (
	"net/http"
	"registry/internal/auth"
	"registry/internal/config"
	"registry/internal/service"
)

func New(cfg *config.Config, registry service.RegistryService, authService auth.Service) *http.ServeMux {
	mux := http.NewServeMux()

	// Register routes for all API versions
	RegisterV0Routes(mux, cfg, registry, authService)

	return mux
}
