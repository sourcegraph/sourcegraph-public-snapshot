package httpserver

import (
	"context"
	"net"
	"net/http"
	"os"
	"sync"

	"github.com/inconshreveable/log15"

	"github.com/sourcegraph/sourcegraph/internal/goroutine"
)

type server struct {
	server       *http.Server
	makeListener func() (net.Listener, error)
	once         sync.Once
}

// New returns a BackgroundRoutine that serves the given server on the given listener.
func New(listener net.Listener, httpServer *http.Server) goroutine.BackgroundRoutine {
	return &server{
		server:       httpServer,
		makeListener: func() (net.Listener, error) { return listener, nil },
	}
}

// New returns a BackgroundRoutine that serves the given handler on the given address.
func NewFromAddr(addr string, httpServer *http.Server) goroutine.BackgroundRoutine {
	return &server{
		server:       httpServer,
		makeListener: func() (net.Listener, error) { return NewListener(addr) },
	}
}

func (s *server) Start() {
	listener, err := s.makeListener()
	if err != nil {
		log15.Error("Failed to create listener", "error", err)
		os.Exit(1)
	}

	if err := s.server.Serve(listener); err != http.ErrServerClosed {
		log15.Error("Failed to start server", "error", err)
		os.Exit(1)
	}
}

func (s *server) Stop() {
	s.once.Do(func() {
		if err := s.server.Shutdown(context.Background()); err != nil {
			log15.Error("Failed to shutdown server", "error", err)
		}
	})
}
