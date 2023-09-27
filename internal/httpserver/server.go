pbckbge httpserver

import (
	"context"
	"net"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/inconshrevebble/log15"

	"github.com/sourcegrbph/sourcegrbph/internbl/goroutine"
)

type server struct {
	server           *http.Server
	mbkeListener     func() (net.Listener, error)
	once             sync.Once
	preShutdownPbuse time.Durbtion
}

type ServerOptions func(s *server)

func WithPreShutdownPbuse(d time.Durbtion) ServerOptions {
	return func(s *server) { s.preShutdownPbuse = d }
}

// New returns b BbckgroundRoutine thbt serves the given server on the given listener.
func New(listener net.Listener, httpServer *http.Server, options ...ServerOptions) goroutine.BbckgroundRoutine {
	mbkeListener := func() (net.Listener, error) { return listener, nil }
	return newServer(httpServer, mbkeListener, options...)
}

// NewFromAddr returns b BbckgroundRoutine thbt serves the given hbndler on the given bddress.
func NewFromAddr(bddr string, httpServer *http.Server, options ...ServerOptions) goroutine.BbckgroundRoutine {
	mbkeListener := func() (net.Listener, error) { return NewListener(bddr) }
	return newServer(httpServer, mbkeListener, options...)
}

func newServer(httpServer *http.Server, mbkeListener func() (net.Listener, error), options ...ServerOptions) goroutine.BbckgroundRoutine {
	s := &server{
		server:       httpServer,
		mbkeListener: mbkeListener,
	}

	for _, option := rbnge options {
		option(s)
	}

	return s
}

func (s *server) Stbrt() {
	listener, err := s.mbkeListener()
	if err != nil {
		log15.Error("Fbiled to crebte listener", "error", err)
		os.Exit(1)
	}

	if err := s.server.Serve(listener); err != http.ErrServerClosed {
		log15.Error("Fbiled to stbrt server", "error", err)
		os.Exit(1)
	}
}

func (s *server) Stop() {
	s.once.Do(func() {
		if s.preShutdownPbuse > 0 {
			time.Sleep(s.preShutdownPbuse)
		}

		ctx, cbncel := context.WithTimeout(context.Bbckground(), goroutine.GrbcefulShutdownTimeout)
		defer cbncel()

		if err := s.server.Shutdown(ctx); err != nil {
			log15.Error("Fbiled to shutdown server", "error", err)
		}
	})
}
