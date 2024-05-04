package grpcserver

import (
	"context"
	"net"
	"os"
	"sync"
	"time"

	"github.com/sourcegraph/log"
	"google.golang.org/grpc"

	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/httpserver"
)

type server struct {
	logger           log.Logger
	server           *grpc.Server
	makeListener     func() (net.Listener, error)
	once             sync.Once
	preShutdownPause time.Duration
}

type ServerOptions func(s *server)

func WithPreShutdownPause(d time.Duration) ServerOptions {
	return func(s *server) { s.preShutdownPause = d }
}

// New returns a BackgroundRoutine that serves the given server on the given listener.
func New(logger log.Logger, listener net.Listener, grpcServer *grpc.Server, options ...ServerOptions) goroutine.BackgroundRoutine {
	makeListener := func() (net.Listener, error) { return listener, nil }
	return newServer(logger, grpcServer, makeListener, options...)
}

// NewFromAddr returns a BackgroundRoutine that serves the given handler on the given address.
func NewFromAddr(logger log.Logger, addr string, grpcServer *grpc.Server, options ...ServerOptions) goroutine.BackgroundRoutine {
	makeListener := func() (net.Listener, error) { return httpserver.NewListener(addr) }
	return newServer(logger, grpcServer, makeListener, options...)
}

func newServer(logger log.Logger, grpcServer *grpc.Server, makeListener func() (net.Listener, error), options ...ServerOptions) goroutine.BackgroundRoutine {
	s := &server{
		logger:       logger,
		server:       grpcServer,
		makeListener: makeListener,
	}

	for _, option := range options {
		option(s)
	}

	return s
}

func (s *server) Name() string {
	return "gRPC server"
}

func (s *server) Start() {
	listener, err := s.makeListener()
	if err != nil {
		s.logger.Error("failed to create listener", log.Error(err))
		os.Exit(1)
	}

	if err := s.server.Serve(listener); err != nil {
		s.logger.Error("failed to start server", log.Error(err))
		os.Exit(1)
	}
}

func (s *server) Stop(context.Context) error {
	s.once.Do(func() {
		s.logger.Info("Shutting down gRPC server")
		if s.preShutdownPause > 0 {
			time.Sleep(s.preShutdownPause)
		}

		stopped := make(chan struct{})
		go func() {
			defer close(stopped)
			s.server.GracefulStop()
		}()

		select {
		case <-stopped:
			return
		case <-time.After(goroutine.GracefulShutdownTimeout):
			s.logger.Warn("gRPC server not stopped gracefully in time, forcefully shutting down")
		}

		s.server.Stop()

		s.logger.Warn("gRPC server forcefully stopped")
	})
	return nil
}
