package grpc

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type Propagator interface {
	ExtractContext(context.Context) metadata.MD
	InjectContext(context.Context, metadata.MD) context.Context
}

func StreamClientPropagator(prop Propagator) grpc.StreamClientInterceptor {
	return func(
		ctx context.Context,
		desc *grpc.StreamDesc,
		cc *grpc.ClientConn,
		method string,
		streamer grpc.Streamer,
		opts ...grpc.CallOption,
	) (grpc.ClientStream, error) {
		md := prop.ExtractContext(ctx)
		for k, vals := range md {
			for _, val := range vals {
				ctx = metadata.AppendToOutgoingContext(ctx, k, val)
			}
		}
		return streamer(ctx, desc, cc, method, opts...)
	}
}

func UnaryClientPropagator(prop Propagator) grpc.UnaryClientInterceptor {
	return func(
		ctx context.Context,
		method string,
		req, reply interface{},
		cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker,
		opts ...grpc.CallOption,
	) error {
		md := prop.ExtractContext(ctx)
		for k, vals := range md {
			for _, val := range vals {
				ctx = metadata.AppendToOutgoingContext(ctx, k, val)
			}
		}
		return invoker(ctx, method, req, reply, cc, opts...)
	}
}

func StreamServerPropagator(prop Propagator) grpc.StreamServerInterceptor {
	return func(
		srv interface{},
		ss grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error {
		ctx := ss.Context()
		md, ok := metadata.FromIncomingContext(ss.Context())
		if ok {
			ctx = prop.InjectContext(ss.Context(), md)
			ss = contextedServerStream{ss, ctx}
		}
		return handler(srv, ss)
	}
}

func UnaryServerPropagator(prop Propagator) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (resp interface{}, err error) {
		md, ok := metadata.FromIncomingContext(ctx)
		if ok {
			ctx = prop.InjectContext(ctx, md)
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
