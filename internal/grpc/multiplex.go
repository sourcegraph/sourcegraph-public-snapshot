package grpc

import (
	"context"
	"net"
	"net/http"

	"github.com/soheilhy/cmux"
	"github.com/sourcegraph/conc/pool"
	"google.golang.org/grpc"
)

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

func (s *MultiplexedServer) Addr() string {
	return s.addr
}

func (s *MultiplexedServer) ListenAndServe() error {
	l, err := net.Listen("tcp", s.addr)
	if err != nil {
		return err
	}

	return s.Serve(l)
}

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
