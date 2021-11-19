// Package symbols implements the symbol search service.
package symbols

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/inconshreveable/log15"

	"github.com/sourcegraph/sourcegraph/cmd/symbols/internal/protocol"
	"github.com/sourcegraph/sourcegraph/internal/diskcache"
)

type service struct {
	gitserverClient GitserverClient
	cache           *diskcache.Store
	parserPool      ParserPool
	fetchSem        chan int
}

type HandlerFactory interface {
	Handler() http.Handler
}

func NewService(
	gitserverClient GitserverClient,
	cache *diskcache.Store,
	parserPool ParserPool,
	maxConcurrentFetchTar int,
) HandlerFactory {
	return newService(gitserverClient, cache, parserPool, maxConcurrentFetchTar)
}

func newService(
	gitserverClient GitserverClient,
	cache *diskcache.Store,
	parserPool ParserPool,
	maxConcurrentFetchTar int,
) *service {
	return &service{
		gitserverClient: gitserverClient,
		cache:           cache,
		parserPool:      parserPool,
		fetchSem:        make(chan int, maxConcurrentFetchTar),
	}
}

// Handler returns the http.Handler that should be used to serve requests.
func (s *service) Handler() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/search", s.handleSearch)
	mux.HandleFunc("/healthz", s.handleHealthCheck)

	return mux
}

func (s *service) handleSearch(w http.ResponseWriter, r *http.Request) {
	var args protocol.SearchArgs
	if err := json.NewDecoder(r.Body).Decode(&args); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	result, err := doSearch(r.Context(), s.gitserverClient, s.cache, s.parserPool, s.fetchSem, args)
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

func (s *service) handleHealthCheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)

	_, err := w.Write([]byte("Ok"))
	if err != nil {
		log.Printf("failed to write response to health check, err: %s", err)
	}
}
