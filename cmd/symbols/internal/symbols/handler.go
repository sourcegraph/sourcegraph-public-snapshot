// Package symbols implements the symbol search service.
package symbols

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/inconshreveable/log15"

	"github.com/sourcegraph/sourcegraph/cmd/symbols/internal/search"
	"github.com/sourcegraph/sourcegraph/cmd/symbols/internal/types"
)

type symbolsHandler struct {
	searcher search.Searcher
}

func NewHandler(
	searcher search.Searcher,
) http.Handler {
	h := &symbolsHandler{
		searcher: searcher,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/search", h.handleSearch)
	mux.HandleFunc("/healthz", h.handleHealthCheck)
	return mux
}

func (h *symbolsHandler) handleSearch(w http.ResponseWriter, r *http.Request) {
	var args types.SearchArgs
	if err := json.NewDecoder(r.Body).Decode(&args); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	result, err := h.searcher.Search(r.Context(), args)
	if err != nil {
		if err == context.Canceled && r.Context().Err() == context.Canceled {
			return // client went away
		}
		log15.Error("Symbol search failed", "args", args, "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(result); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *symbolsHandler) handleHealthCheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)

	_, err := w.Write([]byte("Ok"))
	if err != nil {
		log.Printf("failed to write response to health check, err: %s", err)
	}
}
