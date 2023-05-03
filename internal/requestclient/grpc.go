package requestclient

import (
	"context"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/peer"
)

// UnaryServerInterceptor returns a grpc.UnaryServerInterceptor that sets the
// Client in the context based on the peer address. See https://pkg.go.dev/google.golang.org/grpc/peer
// for more information.
//
// Note: This interceptor doesn't set the ForwardedFor field of the Client, as our gRPC implementation
// only handles requests from internal services.
func UnaryServerInterceptor(ctx context.Context, req any, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
	p, ok := peer.FromContext(ctx)
	if !ok || p == nil {
		return handler(ctx, req)
	}

	client := &Client{
		IP: baseIP(p.Addr),
	}

	ctx = WithClient(ctx, client)
	return handler(ctx, req)
}

// StreamServerInterceptor returns a grpc.StreamServerInterceptor that sets the
// Client in the context based on the peer address. See https://pkg.go.dev/google.golang.org/grpc/peer
// for more information.
//
// Note: This interceptor doesn't set the ForwardedFor field of the Client, as our gRPC implementation
// only handles requests from internal services.
func StreamServerInterceptor(srv any, ss grpc.ServerStream, _ *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	p, ok := peer.FromContext(ss.Context())
	if !ok || p == nil {
		return handler(srv, ss)
	}

	client := &Client{
		IP: baseIP(p.Addr),
	}

	ctx := WithClient(ss.Context(), client)
	return handler(srv, &customContextServerStream{
		ServerStream:  ss,
		customContext: ctx,
	})
}

// customContextServerStream is a wrapper around grpc.ServerStream that returns a custom context
// from the Context method.
type customContextServerStream struct {
	grpc.ServerStream
	customContext context.Context
}

func (s *customContextServerStream) Context() context.Context {
	return s.customContext
}

var _ grpc.ServerStream = &customContextServerStream{}

// baseIP returns the base IP address of the given net.Addr
func baseIP(addr net.Addr) string {
	switch a := addr.(type) {
	case *net.TCPAddr:
		return a.IP.String()
	case *net.UDPAddr:
		return a.IP.String()
	default:
		return addr.String()
	}
}
