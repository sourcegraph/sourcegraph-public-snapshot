package grpc

import (
	"context"
	"net"
	"net/http"
	"sync"

	"github.com/cockroachdb/errors"
	"github.com/soheilhy/cmux"
	"github.com/sourcegraph/conc/pool"
	"google.golang.org/grpc"
)

// NewMultiplexedServer will create a server that dynamically switches between
// the provided gRPC server and HTTP server depending on whether the header
// `content-type: application/grpc` is set.
func NewMultiplexedServer(addr string, grpcServer *grpc.Server, httpServer *http.Server) *MultiplexedServer {
	ctx, cancel := context.WithCancel(context.Background())

	return &MultiplexedServer{
		ctx: ctx,

		shutdownCMuxListener: cancel,

		backgroundTasksDone:  make(chan struct{}),
		serverShutdownCalled: make(chan struct{}),

		addr:       addr,
		httpServer: httpServer,
		grpcServer: grpcServer,
	}
}

type MultiplexedServer struct {
	ctx context.Context

	shutdownCMuxListener func()

	backgroundTasksDone  chan struct{}
	serverShutdownCalled chan struct{}

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

	// possibleShutdownErrors are errors that can possibly be returned from any of the
	// underlying server's Serve() when they are torn down. If someone explicitly calls MultiplexedServer.Shutdown(),
	// we avoid returning these errors from Serve() itself to avoid confusion.
	possibleShutdownErrors := []error{
		net.ErrClosed,          // returned when someone tries to write to a closed connection
		cmux.ErrListenerClosed, // returned by the cmux listener when the underlying listener is closed
		cmux.ErrServerClosed,   // returned by the cmux listener when the server itself is closed
	}

	p.Go(func() error {
		err := s.grpcServer.Serve(grpcL)

		select {
		case <-s.serverShutdownCalled:
			if errors.IsAny(err, possibleShutdownErrors...) {
				return nil
			}

			return err
		default:
			return err
		}
	})

	p.Go(func() error {
		err := s.httpServer.Serve(httpL)

		select {
		case <-s.serverShutdownCalled:
			if errors.IsAny(err, possibleShutdownErrors...) {
				return nil
			}

			return err
		default:
			return err
		}
	})

	p.Go(func() error {
		err := m.Serve()

		select {
		case <-s.serverShutdownCalled:
			if errors.IsAny(err, possibleShutdownErrors...) {
				return nil
			}

			return err
		default:
			return err
		}
	})

	p.Go(func() error {
		<-s.ctx.Done()
		m.Close()
		return nil
	})

	return p.Wait()
}

// Shutdown gracefully shuts down the server without interrupting any
// active connections.

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
	close(s.serverShutdownCalled) // signal that the server has been shutdown

	// First, shutdown the underlying gRPC and HTTP servers
	s.grpcServer.GracefulStop()
	httpErr := s.httpServer.Shutdown(ctx)

	// Then, close the underlying listener
	s.shutdownCMuxListener()

	if httpErr != nil {
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
