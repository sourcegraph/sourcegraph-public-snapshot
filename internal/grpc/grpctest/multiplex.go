package grpctest

import (
	"context"
	"net"
	"net/http"

	"google.golang.org/grpc"

	internalgrpc "github.com/sourcegraph/sourcegraph/internal/grpc"
)

type MultiplexedServer struct {
	l net.Listener
	s *internalgrpc.MultiplexedServer
}

// NewMultiplexedServer creates and starts a server that listens on an arbitrary port selected by the OS.
// It will dynamically handle requests with either grpcServer or httpHandler depending on whether the
// header `content-type: application/grpc` is set.
func NewMultiplexedServer(grpcServer *grpc.Server, httpHandler http.Handler) *MultiplexedServer {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		panic("grpctest: failed to listen: " + err.Error())
	}

	s := internalgrpc.NewMultiplexedServer(l.Addr().String(), grpcServer, &http.Server{Handler: httpHandler})
	go s.Serve(l)

	return &MultiplexedServer{
		l: l,
		s: s,
	}
}

// Stop shuts down the server and waits for all connections to be closed.
func (s *MultiplexedServer) Stop() {
	s.s.Shutdown(context.Background())
}

// Addr returns the address this server is listening on. The address is generated dynamically, so
// any use of grptest.MultiplexedServer will need to call this to be effective.
func (s *MultiplexedServer) Addr() string {
	return s.l.Addr().String()
}
