package httpserver

import (
	"context"
	"net"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/inconshreveable/log15"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/trace/ot"
)

type server struct {
	listener net.Listener
	server   *http.Server
	once     sync.Once
}

type Options struct {
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

// New returns a BackgroundRoutine that serves the given handler on the given listener.
func New(listener net.Listener, handler http.Handler, options Options) goroutine.BackgroundRoutine {
	httpServer := &http.Server{
		Handler:      ot.Middleware(handler),
		ReadTimeout:  options.ReadTimeout,
		WriteTimeout: options.WriteTimeout,
	}

	return &server{
		listener: listener,
		server:   httpServer,
	}
}

// New returns a BackgroundRoutine that serves the given handler on the given address.
func NewFromAddr(addr string, handler http.Handler, options Options) (goroutine.BackgroundRoutine, error) {
	listener, err := NewListener(addr)
	if err != nil {
		return nil, err
	}

	return New(listener, handler, options), nil
}

func (s *server) Start() {
	if err := s.server.Serve(s.listener); err != http.ErrServerClosed {
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
