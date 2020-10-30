package httpserver

import (
	"net/http"

	"github.com/gorilla/mux"
)

// NewHandler creates an HTTP handler with a default /healthz endpoint.
// If a function is provided, it will be invoked with a router on which
// additional routes can be installed.
func NewHandler(setupRoutes func(router *mux.Router)) http.Handler {
	router := mux.NewRouter()
	router.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	if setupRoutes != nil {
		setupRoutes(router)
	}

	return router
}
