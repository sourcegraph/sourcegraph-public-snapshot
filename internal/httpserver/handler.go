package httpserver

import (
	"net/http"

	"github.com/gorilla/mux"
)

// TODO - document
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
