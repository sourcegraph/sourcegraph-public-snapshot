package api

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/inconshreveable/log15"

	"github.com/sourcegraph/go-ctags"

	"github.com/sourcegraph/sourcegraph/cmd/symbols/internal/database/writer"
	"github.com/sourcegraph/sourcegraph/cmd/symbols/internal/types"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type apiHandler struct {
	cachedDatabaseWriter writer.CachedDatabaseWriter
	ctagsBinary          string
	operations           *operations
}

func NewHandler(
	cachedDatabaseWriter writer.CachedDatabaseWriter,
	ctagsBinary string,
	observationContext *observation.Context,
) http.Handler {
	h := newAPIHandler(cachedDatabaseWriter, ctagsBinary, observationContext)

	mux := http.NewServeMux()
	mux.HandleFunc("/search", h.handleSearch)
	mux.HandleFunc("/healthz", h.handleHealthCheck)
	mux.HandleFunc("/list-languages", h.handleListLanguages)
	return mux
}

func newAPIHandler(
	cachedDatabaseWriter writer.CachedDatabaseWriter,
	ctagsBinary string,
	observationContext *observation.Context,
) *apiHandler {
	return &apiHandler{
		cachedDatabaseWriter: cachedDatabaseWriter,
		ctagsBinary:          ctagsBinary,
		operations:           newOperations(observationContext),
	}
}

const maxNumSymbolResults = 500

func (h *apiHandler) handleSearch(w http.ResponseWriter, r *http.Request) {
	var args types.SearchArgs
	if err := json.NewDecoder(r.Body).Decode(&args); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if args.First < 0 || args.First > maxNumSymbolResults {
		args.First = maxNumSymbolResults
	}

	result, err := h.handleSearchInternal(r.Context(), args)
	if err != nil {
		// Ignore reporting errors where client disconnected
		if r.Context().Err() == context.Canceled && errors.Is(err, context.Canceled) {
			return
		}

		log15.Error("Symbol search failed", "args", args, "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(result); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (h *apiHandler) handleListLanguages(w http.ResponseWriter, r *http.Request) {
	mapping, err := ctags.ListLanguageMappings(r.Context(), h.ctagsBinary)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := json.NewEncoder(w).Encode(mapping); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (h *apiHandler) handleHealthCheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)

	if _, err := w.Write([]byte("OK")); err != nil {
		log15.Error("failed to write response to health check, err: %s", err)
	}
}
