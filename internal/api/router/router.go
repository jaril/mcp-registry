package router

import "net/http"

func New() *http.ServeMux {
	mux := http.NewServeMux()

	// Register routes for all API versions
	RegisterV0Routes(mux)

	return mux
}
