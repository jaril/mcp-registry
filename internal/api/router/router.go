package router

import (
	"net/http"
	"registry/internal/config"
	"registry/internal/service"
)

func New(cfg *config.Config, registry service.RegistryService) *http.ServeMux {
	mux := http.NewServeMux()

	// Register routes for all API versions
	RegisterV0Routes(mux, cfg, registry)

	return mux
}
