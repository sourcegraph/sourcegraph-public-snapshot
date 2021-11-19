// Package symbols implements the symbol search service.
package symbols

import (
	"log"
	"net/http"

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

func (s *service) handleHealthCheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)

	_, err := w.Write([]byte("Ok"))
	if err != nil {
		log.Printf("failed to write response to health check, err: %s", err)
	}
}
