package http

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"

	"github.com/opentracing/opentracing-go/log"

	uploads "github.com/sourcegraph/sourcegraph/internal/codeintel/uploads"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type Handler struct {
	svc        *uploads.Service
	operations *operations
}

var _ http.Handler = &Handler{}

func newHandler(svc *uploads.Service, observationContext *observation.Context) *Handler {
	return &Handler{
		svc:        svc,
		operations: newOperations(observationContext),
	}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.serveJSON(w, r)
}

func (h *Handler) serveJSON(w http.ResponseWriter, r *http.Request) {
	payload, statusCode, err := h.handleRequest(r)
	if err != nil {
		handleErr(w, err, "request failed", statusCode)
		return
	}
	if payload == nil {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	data, err := json.Marshal(payload)
	if err != nil {
		handleErr(w, err, "failed to serialize result", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(statusCode)

	if _, err := io.Copy(w, bytes.NewReader(data)); err != nil {
		handleErr(nil, err, "failed to write payload to client", http.StatusInternalServerError)
		return
	}
}

func (h *Handler) handleRequest(r *http.Request) (payload any, statusCode int, err error) {
	ctx, trace, endObservation := h.operations.todo.With(r.Context(), &err, observation.Args{})
	defer func() {
		endObservation(1, observation.Args{LogFields: []log.Field{
			log.Int("statusCode", statusCode),
		}})
	}()

	// To be implemented in https://github.com/sourcegraph/sourcegraph/issues/33375
	_, _ = ctx, trace
	return nil, 0, nil
}
