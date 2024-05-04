package httpserver

import (
	"context"
	"net"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/inconshreveable/log15" //nolint:logging // TODO move all logging to sourcegraph/log

	"github.com/sourcegraph/sourcegraph/internal/goroutine"
)

type server struct {
	server           *http.Server
	makeListener     func() (net.Listener, error)
	once             sync.Once
	preShutdownPause time.Duration
}

type ServerOptions func(s *server)

func WithPreShutdownPause(d time.Duration) ServerOptions {
	return func(s *server) { s.preShutdownPause = d }
}

// New returns a BackgroundRoutine that serves the given server on the given listener.
func New(listener net.Listener, httpServer *http.Server, options ...ServerOptions) goroutine.BackgroundRoutine {
	makeListener := func() (net.Listener, error) { return listener, nil }
	return newServer(httpServer, makeListener, options...)
}

// NewFromAddr returns a BackgroundRoutine that serves the given handler on the given address.
func NewFromAddr(addr string, httpServer *http.Server, options ...ServerOptions) goroutine.BackgroundRoutine {
	makeListener := func() (net.Listener, error) { return NewListener(addr) }
	return newServer(httpServer, makeListener, options...)
}

func newServer(httpServer *http.Server, makeListener func() (net.Listener, error), options ...ServerOptions) goroutine.BackgroundRoutine {
	s := &server{
		server:       httpServer,
		makeListener: makeListener,
	}

	for _, option := range options {
		option(s)
	}

	return s
}

func (s *server) Name() string {
	return "HTTP server"
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

func (s *server) Stop(ctx context.Context) error {
	s.once.Do(func() {
		if s.preShutdownPause > 0 {
			time.Sleep(s.preShutdownPause)
		}

		ctx, cancel := context.WithTimeout(ctx, goroutine.GracefulShutdownTimeout)
		defer cancel()

		if err := s.server.Shutdown(ctx); err != nil {
			log15.Error("Failed to shutdown server", "error", err)
		}
	})
	return nil
}
