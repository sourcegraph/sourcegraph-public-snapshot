package server

import (
	"net/http"

	"github.com/sourcegraph/sourcegraph/internal/api"

	"github.com/sourcegraph/sourcegraph/cmd/gitserver/domain"
)

func (s *Server) handleGetObject(svc domain.GetObjectService) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// TODO: Actually parse input
		repo := api.RepoName("parsedRepo")
		objectName := "parsedObject"

		// All work actually done here:
		_, _ = svc.GetObject(r.Context(), repo, objectName)

		// TODO: Marshal object
	}
}
