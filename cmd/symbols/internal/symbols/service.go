// Package symbols implements the symbol search service.
package symbols

import (
	"log"
	"net/http"

	"github.com/sourcegraph/sourcegraph/internal/diskcache"
)

// Service is the symbols service.
type Service struct {
	// Path is the directory in which to store the cache.
	Path string

	GitserverClient GitserverClient

	// Cache is the disk backed Cache.
	Cache *diskcache.Store

	ParserPool ParserPool

	fetchSem chan int
}

type HandlerFactory interface {
	Handler() http.Handler
}

func NewService(
	path string,
	gitserverClient GitserverClient,
	cache *diskcache.Store,
	parserPool ParserPool,
	maxConcurrentFetchTar int,
) HandlerFactory {
	return newService(path, gitserverClient, cache, parserPool, maxConcurrentFetchTar)
}

func newService(
	path string,
	gitserverClient GitserverClient,
	cache *diskcache.Store,
	parserPool ParserPool,
	maxConcurrentFetchTar int,
) *Service {
	return &Service{
		Path:            path,
		GitserverClient: gitserverClient,
		Cache:           cache,
		ParserPool:      parserPool,
		fetchSem:        make(chan int, maxConcurrentFetchTar),
	}
}

// Handler returns the http.Handler that should be used to serve requests.
func (s *Service) Handler() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/search", s.handleSearch)
	mux.HandleFunc("/healthz", s.handleHealthCheck)

	return mux
}

func (s *Service) handleHealthCheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)

	_, err := w.Write([]byte("Ok"))
	if err != nil {
		log.Printf("failed to write response to health check, err: %s", err)
	}
}
