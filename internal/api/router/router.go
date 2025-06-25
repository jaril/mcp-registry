package router

import (
	"net/http"
	"registry/internal/config"
)

func New(cfg *config.Config) *http.ServeMux {
	mux := http.NewServeMux()

	// Register routes for all API versions
	RegisterV0Routes(mux, cfg)

	return mux
}
