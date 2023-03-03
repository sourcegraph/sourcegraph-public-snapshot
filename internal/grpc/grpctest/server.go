package grpctest

import (
	"net"
	"net/http"

	"google.golang.org/grpc"

	internalgrpc "github.com/sourcegraph/sourcegraph/internal/grpc"
)

type MultiplexedServer struct {
	l net.Listener
	s *internalgrpc.MultiplexedServer
}

func NewMultiplexedServer(grpcServer *grpc.Server, handler http.Handler) *MultiplexedServer {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		panic("grpctest: failed to listen: " + err.Error())
	}

	s := internalgrpc.NewMultiplexedServer(l.Addr().String(), grpcServer, &http.Server{Handler: handler})
	go s.Serve(l)

	return &MultiplexedServer{
		l: l,
		s: s,
	}
}

func (s *MultiplexedServer) Stop() {
	s.l.Close()
}

func (s *MultiplexedServer) Addr() string {
	return s.l.Addr().String()
}
