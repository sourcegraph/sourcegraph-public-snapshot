package httpserver

import (
	"context"
	"net"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/inconshreveable/log15"

	"github.com/sourcegraph/sourcegraph/internal/conf/deploy"
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
		// On kubernetes, we want to wait an additional 5 seconds after we receive a
		// shutdown request to give some additional time for the endpoint changes
		// to propagate to services talking to this server like the LB or ingress
		// controller. We only do this in frontend and not on all services, because
		// frontend is the only publicly exposed service where we don't control
		// retries on connection failures (see httpcli.InternalClient).
		if deploy.Type() == deploy.Kubernetes {
			time.Sleep(5 * time.Second)
		}

		ctx, cancel := context.WithTimeout(context.Background(), goroutine.GracefulShutdownTimeout)
		defer cancel()

		if err := s.server.Shutdown(ctx); err != nil {
			log15.Error("Failed to shutdown server", "error", err)
		}
	})
}
