package server

import (
	"net/http"

	"github.com/gorilla/mux"
)

func (s *Server) handler() http.Handler {
	mux := mux.NewRouter()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	return mux
}
