package propagator

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// Propagator is a type that can extract some information from a context.Context,
// returning it in the form of metadata.MD and can also inject that same metadata
// back into a context on the server side of an RPC call.
type Propagator interface {
	// FromContext extracts the information to be propagated from a context,
	// converting it to a metadata.MD. This will be called on the client side
	// of an RPC.
	FromContext(context.Context) metadata.MD

	// InjectContext takes a context and some metadata and creates a new context
	// with the information from the metadata injected into the context.
	// This will be called on the server side of an RPC.
	InjectContext(context.Context, metadata.MD) (context.Context, error)
}

// StreamClientPropagator returns an interceptor that will use the given propagator
// to forward some information from the context across the RPC call. The server
// should be configured with an interceptor that uses the same propagator.
func StreamClientPropagator(prop Propagator) grpc.StreamClientInterceptor {
	return func(
		ctx context.Context,
		desc *grpc.StreamDesc,
		cc *grpc.ClientConn,
		method string,
		streamer grpc.Streamer,
		opts ...grpc.CallOption,
	) (grpc.ClientStream, error) {
		md := prop.FromContext(ctx)
		for k, vals := range md {
			for _, val := range vals {
				ctx = metadata.AppendToOutgoingContext(ctx, k, val)
			}
		}
		return streamer(ctx, desc, cc, method, opts...)
	}
}

// UnaryClientPropagator returns an interceptor that will use the given propagator
// to forward some information from the context across the RPC call. The server
// should be configured with an interceptor that uses the same propagator.
func UnaryClientPropagator(prop Propagator) grpc.UnaryClientInterceptor {
	return func(
		ctx context.Context,
		method string,
		req, reply interface{},
		cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker,
		opts ...grpc.CallOption,
	) error {
		md := prop.FromContext(ctx)
		for k, vals := range md {
			for _, val := range vals {
				ctx = metadata.AppendToOutgoingContext(ctx, k, val)
			}
		}
		return invoker(ctx, method, req, reply, cc, opts...)
	}
}

// StreamServerPropagator returns an interceptor that will use the given propagator
// to translate some metadata back into the context for the RPC handler. The client
// should be configured with an interceptor that uses the same propagator.
func StreamServerPropagator(prop Propagator) grpc.StreamServerInterceptor {
	return func(
		srv interface{},
		ss grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error {
		ctx := ss.Context()
		md, ok := metadata.FromIncomingContext(ctx)
		if ok {
			var err error
			ctx, err = prop.InjectContext(ss.Context(), md)
			if err != nil {
				return err
			}
			ss = contextedServerStream{ss, ctx}
		}
		return handler(srv, ss)
	}
}

// UnaryServerPropagator returns an interceptor that will use the given propagator
// to translate some metadata back into the context for the RPC handler. The client
// should be configured with an interceptor that uses the same propagator.
func UnaryServerPropagator(prop Propagator) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (resp interface{}, err error) {
		md, ok := metadata.FromIncomingContext(ctx)
		if ok {
			ctx, err = prop.InjectContext(ctx, md)
			if err != nil {
				return nil, err
			}
		}
		return handler(ctx, req)
	}
}

type contextedServerStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (css contextedServerStream) Context() context.Context {
	return css.ctx
}
