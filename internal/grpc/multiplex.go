package grpc

import (
	"context"
	"net"
	"net/http"

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
		ctx:        ctx,
		cancel:     cancel,
		done:       make(chan struct{}),
		addr:       addr,
		httpServer: httpServer,
		grpcServer: grpcServer,
	}
}

type MultiplexedServer struct {
	ctx    context.Context
	cancel context.CancelFunc
	done   chan struct{}

	addr       string
	httpServer *http.Server
	grpcServer *grpc.Server
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
	defer close(s.done)

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
		<-s.ctx.Done()
		m.Close()
		return nil
	})

	return p.Wait()
}

func (s *MultiplexedServer) Shutdown(ctx context.Context) error {
	s.cancel()
	s.grpcServer.Stop()
	err := s.httpServer.Shutdown(ctx)
	if err != nil {
		return err
	}

	select {
	case <-s.done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
