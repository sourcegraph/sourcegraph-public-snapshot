package grpc

import (
	"context"
	"net"
	"net/http"
	"sync"

	"github.com/soheilhy/cmux"
	"github.com/sourcegraph/conc/pool"
	"google.golang.org/grpc"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// NewMultiplexedServer will create a server that dynamically switches between
// the provided gRPC server and HTTP server depending on whether the header
// `content-type: application/grpc` is set.
func NewMultiplexedServer(addr string, grpcServer *grpc.Server, httpServer *http.Server) *MultiplexedServer {
	return &MultiplexedServer{
		shutdownCMuxListener: make(chan struct{}),
		backgroundTasksDone:  make(chan struct{}),

		addr:       addr,
		httpServer: httpServer,
		grpcServer: grpcServer,
	}
}

type MultiplexedServer struct {
	shutdownCMuxListener chan struct{}
	backgroundTasksDone  chan struct{}

	addr       string
	httpServer *http.Server
	grpcServer *grpc.Server

	shutdownOnce sync.Once
	shutdownErr  error
}

// ListenAndServe listens on the TCP network address addr and then calls
// Serve to handle requests on incoming connections.
func (s *MultiplexedServer) ListenAndServe() error {
	l, err := net.Listen("tcp", s.addr)
	if err != nil {
		return err
	}

	return s.Serve(l)
}

// Serve accepts connections on the listener and passes them to the
// appropriate underlying server.
func (s *MultiplexedServer) Serve(l net.Listener) error {
	defer close(s.backgroundTasksDone)

	m := cmux.New(l)

	grpcL := m.Match(cmux.HTTP2HeaderField("content-type", "application/grpc"))
	httpL := m.Match(cmux.Any())

	p := pool.New().WithErrors()

	p.Go(func() error {
		return s.grpcServer.Serve(grpcL)
	})

	p.Go(func() error {
		return s.httpServer.Serve(httpL)
	})

	p.Go(func() error {
		return m.Serve()
	})

	p.Go(func() error {
		<-s.shutdownCMuxListener
		m.Close()
		return nil
	})

	err := p.Wait()

	// possibleShutdownErrors are errors that can possibly be returned from any of the
	// underlying server's Serve() when they are torn down. If someone explicitly calls MultiplexedServer.Shutdown(),
	// we avoid returning these errors from Serve() itself to avoid confusion.
	possibleShutdownErrors := []error{
		net.ErrClosed,          // returned when someone tries to write to a closed connection
		cmux.ErrListenerClosed, // returned by the cmux listener when the underlying listener is closed
		cmux.ErrServerClosed,   // returned by the cmux listener when the server itself is closed
	}
	err = errors.Ignore(err, errors.IsAnyPred(possibleShutdownErrors...))

	return err
}

// Shutdown gracefully shuts down the server without interrupting any
// active connections.
//
// Shutdown works by first by preventing new connections from being accepted
// and then waiting for all existing gRPC RPCs to complete and all HTTP
// connections to become idle. Finally, the underlying listener is closed.
//
// If the provided context expires before the shutdown is complete, Shutdown
// returns the context's error, otherwise it returns any error returned from
// closing the underlying listener.
//
// Shutdown maybe only be called once. Subsequent calls will return the
// error returned from the first call (if any).
func (s *MultiplexedServer) Shutdown(ctx context.Context) error {
	s.shutdownOnce.Do(func() {
		s.shutdownErr = s.doShutdown(ctx)
	})

	return s.shutdownErr
}

func (s *MultiplexedServer) doShutdown(ctx context.Context) error {
	// First, shutdown the underlying gRPC and HTTP servers
	grpcErr := ShutdownGRPCServer(ctx, s.grpcServer)
	httpErr := s.httpServer.Shutdown(ctx)

	// Then, close the underlying listener
	close(s.shutdownCMuxListener)

	err := errors.Append(grpcErr, httpErr)
	if err != nil {
		return httpErr
	}

	select {
	case <-s.backgroundTasksDone:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (s *MultiplexedServer) AsBackgroundGoroutine() *BackgroundServer {
	return &BackgroundServer{s}
}

// ShutdownGRPCServer tries to stop the gRPC server gracefully, but if the
// context is canceled before that happens, it returns early with the context error.
func ShutdownGRPCServer(ctx context.Context, grpcServer *grpc.Server) error {
	stoppedGracefully := make(chan struct{})
	go func() {
		defer close(stoppedGracefully)
		grpcServer.GracefulStop()
	}()

	select {
	case <-stoppedGracefully:
		return nil
	case <-ctx.Done():
		grpcServer.Stop()
		return ctx.Err()
	}
}

// BackgroundServer is an implementation of goroutine.BackgroundGoroutine
// for MultiplexedServer
type BackgroundServer struct {
	s *MultiplexedServer
}

func (bs *BackgroundServer) Start() {
	go func() {
		bs.s.ListenAndServe()
	}()
}

func (bs *BackgroundServer) Stop() {
	bs.s.Shutdown(context.Background())
}
