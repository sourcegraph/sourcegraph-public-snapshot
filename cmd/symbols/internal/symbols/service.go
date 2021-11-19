// Package symbols implements the symbol search service.
package symbols

import (
	"log"
	"net/http"

	"github.com/sourcegraph/sourcegraph/internal/diskcache"
)

// Service is the symbols service.
type Service struct {
	GitserverClient GitserverClient

	// MaxConcurrentFetchTar is the maximum number of concurrent calls allowed
	// to FetchTar. It defaults to 15.
	MaxConcurrentFetchTar int

	// Path is the directory in which to store the cache.
	Path string

	// Cache is the disk backed Cache.
	Cache *diskcache.Store

	ParserPool ParserPool

	// fetchSem is a semaphore to limit concurrent calls to FetchTar. The
	// semaphore size is controlled by MaxConcurrentFetchTar
	fetchSem chan int
}

// Start must be called before any requests are handled.
func (s *Service) Init() error {
	if s.MaxConcurrentFetchTar == 0 {
		s.MaxConcurrentFetchTar = 15
	}
	s.fetchSem = make(chan int, s.MaxConcurrentFetchTar)
	return nil
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
