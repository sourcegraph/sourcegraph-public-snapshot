// Package symbols implements the symbol search service.
package symbols

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/inconshreveable/log15"

	"github.com/sourcegraph/sourcegraph/cmd/symbols/internal/parser"
	"github.com/sourcegraph/sourcegraph/cmd/symbols/internal/search"
	"github.com/sourcegraph/sourcegraph/cmd/symbols/internal/sqlite"
	"github.com/sourcegraph/sourcegraph/cmd/symbols/internal/types"
	"github.com/sourcegraph/sourcegraph/internal/diskcache"
)

type symbolsHandler struct {
	gitserverClient sqlite.GitserverClient
	parser          parser.Parser
	cache           *diskcache.Store
}

func NewHandler(
	gitserverClient sqlite.GitserverClient,
	parser parser.Parser,
	cache *diskcache.Store,
) http.Handler {
	h := &symbolsHandler{
		gitserverClient: gitserverClient,
		parser:          parser,
		cache:           cache,
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

	result, err := search.Search(r.Context(), h.gitserverClient, h.parser, h.cache, args, sqlite.WriteDBFile)
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
