package http

import (
	"net/http"

	uploads "github.com/sourcegraph/sourcegraph/internal/codeintel/uploads"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type Handler struct {
	svc        *uploads.Service
	operations *operations
}

func newHandler(svc *uploads.Service, observationContext *observation.Context) http.Handler {
	operations := newOperations(observationContext)
	_ = operations // return &Handler{svc: svc}

	return nil
}
