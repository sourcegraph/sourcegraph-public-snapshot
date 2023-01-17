// Package grpc is a set of shared code for implementing gRPC.
package grpc

import (
	"net/http"
	"strings"

	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"google.golang.org/grpc"
)

// MultiplexHandlers takes a gRPC server and a plain HTTP handler and multiplexes the
// request handling. Any requests that declare themselves as gRPC requests are routed
// to the gRPC server, all others are routed to the httpHandler.
func MultiplexHandlers(grpcServer *grpc.Server, httpHandler http.Handler) http.Handler {
	newHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.ProtoMajor == 2 && strings.Contains(r.Header.Get("Content-Type"), "application/grpc") {
			grpcServer.ServeHTTP(w, r)
		} else {
			httpHandler.ServeHTTP(w, r)
		}
	})

	// Until we enable TLS, we need to fall back to the h2c protocol, which is
	// basically HTTP2 without TLS. The standard library does not implement the
	// h2s protocol, so this hijacks h2s requests and handles them correctly.
	return h2c.NewHandler(newHandler, &http2.Server{})
}

func DefaultDialOptions() []grpc.DialOption {
	// Generate the options dynamically rather than using a static slice
	// because the tracer will not be initialized during init time.
	return []grpc.DialOption{
		grpc.WithStreamInterceptor(otelgrpc.StreamClientInterceptor()),
		grpc.WithUnaryInterceptor(otelgrpc.UnaryClientInterceptor()),
	}
}

func DefaultServerOptions() []grpc.ServerOption {
	// Generate the options dynamically rather than using a static slice
	// because the tracer will not be initialized during init time.
	return []grpc.ServerOption{
		grpc.ChainStreamInterceptor(otelgrpc.StreamServerInterceptor()),
		grpc.ChainUnaryInterceptor(otelgrpc.UnaryServerInterceptor()),
	}
}
