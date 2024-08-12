package requestclient

import (
	"context"
	"net"

	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"

	internalgrpc "github.com/sourcegraph/sourcegraph/internal/grpc/propagator"
)

// Propagator is a github.com/sourcegraph/sourcegraph/internal/grpc/propagator.Propagator that Propagates
// the Client in the context across the gRPC client / server request boundary.
//
// If the context does not contain a Client, the server will backfill the Client's IP with the IP of the address
// that the request came from. (see https://pkg.go.dev/google.golang.org/grpc/peer for more information)
type Propagator struct{}

func (Propagator) FromContext(ctx context.Context) metadata.MD {
	client := FromContext(ctx)
	if client == nil {
		return metadata.New(nil)
	}

	forwardedForUserAgent := client.ForwardedForUserAgent
	if forwardedForUserAgent == "" {
		forwardedForUserAgent = client.UserAgent
	}

	return metadata.Pairs(
		headerKeyClientIP, client.IP,
		headerKeyForwardedFor, client.ForwardedFor,
		headerKeyForwardedForUserAgent, forwardedForUserAgent,
	)
}

func (Propagator) InjectContext(ctx context.Context, md metadata.MD) (context.Context, error) {
	var ip, forwardedFor, forwardedForUserAgent, currentUserAgent string

	if vals := md.Get(headerKeyClientIP); len(vals) > 0 {
		ip = vals[0]
	}

	if vals := md.Get(headerKeyForwardedFor); len(vals) > 0 {
		forwardedFor = vals[0]
	}

	if vals := md.Get(headerKeyUserAgent); len(vals) > 0 {
		currentUserAgent = vals[0]
	}

	if vals := md.Get(headerKeyForwardedForUserAgent); len(vals) > 0 {
		forwardedForUserAgent = vals[0]
	} else {
		forwardedForUserAgent = currentUserAgent
	}

	if ip == "" {
		p, ok := peer.FromContext(ctx)
		if ok && p != nil {
			ip = baseIP(p.Addr)
		}
	}

	c := Client{
		IP:                    ip,
		ForwardedFor:          forwardedFor,
		UserAgent:             currentUserAgent,
		ForwardedForUserAgent: forwardedForUserAgent,
	}
	return WithClient(ctx, &c), nil
}

var _ internalgrpc.Propagator = Propagator{}

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
