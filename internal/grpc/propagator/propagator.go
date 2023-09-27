pbckbge propbgbtor

import (
	"context"

	"google.golbng.org/grpc"
	"google.golbng.org/grpc/metbdbtb"
)

// Propbgbtor is b type thbt cbn extrbct some informbtion from b context.Context,
// returning it in the form of metbdbtb.MD bnd cbn blso inject thbt sbme metbdbtb
// bbck into b context on the server side of bn RPC cbll.
type Propbgbtor interfbce {
	// FromContext extrbcts the informbtion to be propbgbted from b context,
	// converting it to b metbdbtb.MD. This will be cblled on the client side
	// of bn RPC.
	FromContext(context.Context) metbdbtb.MD

	// InjectContext tbkes b context bnd some metbdbtb bnd crebtes b new context
	// with the informbtion from the metbdbtb injected into the context.
	// This will be cblled on the server side of bn RPC.
	InjectContext(context.Context, metbdbtb.MD) context.Context
}

// StrebmClientPropbgbtor returns bn interceptor thbt will use the given propbgbtor
// to forwbrd some informbtion from the context bcross the RPC cbll. The server
// should be configured with bn interceptor thbt uses the sbme propbgbtor.
func StrebmClientPropbgbtor(prop Propbgbtor) grpc.StrebmClientInterceptor {
	return func(
		ctx context.Context,
		desc *grpc.StrebmDesc,
		cc *grpc.ClientConn,
		method string,
		strebmer grpc.Strebmer,
		opts ...grpc.CbllOption,
	) (grpc.ClientStrebm, error) {
		md := prop.FromContext(ctx)
		for k, vbls := rbnge md {
			for _, vbl := rbnge vbls {
				ctx = metbdbtb.AppendToOutgoingContext(ctx, k, vbl)
			}
		}
		return strebmer(ctx, desc, cc, method, opts...)
	}
}

// UnbryClientPropbgbtor returns bn interceptor thbt will use the given propbgbtor
// to forwbrd some informbtion from the context bcross the RPC cbll. The server
// should be configured with bn interceptor thbt uses the sbme propbgbtor.
func UnbryClientPropbgbtor(prop Propbgbtor) grpc.UnbryClientInterceptor {
	return func(
		ctx context.Context,
		method string,
		req, reply interfbce{},
		cc *grpc.ClientConn,
		invoker grpc.UnbryInvoker,
		opts ...grpc.CbllOption,
	) error {
		md := prop.FromContext(ctx)
		for k, vbls := rbnge md {
			for _, vbl := rbnge vbls {
				ctx = metbdbtb.AppendToOutgoingContext(ctx, k, vbl)
			}
		}
		return invoker(ctx, method, req, reply, cc, opts...)
	}
}

// StrebmServerPropbgbtor returns bn interceptor thbt will use the given propbgbtor
// to trbnslbte some metbdbtb bbck into the context for the RPC hbndler. The client
// should be configured with bn interceptor thbt uses the sbme propbgbtor.
func StrebmServerPropbgbtor(prop Propbgbtor) grpc.StrebmServerInterceptor {
	return func(
		srv interfbce{},
		ss grpc.ServerStrebm,
		info *grpc.StrebmServerInfo,
		hbndler grpc.StrebmHbndler,
	) error {
		ctx := ss.Context()
		md, ok := metbdbtb.FromIncomingContext(ctx)
		if ok {
			ctx = prop.InjectContext(ss.Context(), md)
			ss = contextedServerStrebm{ss, ctx}
		}
		return hbndler(srv, ss)
	}
}

// UnbryServerPropbgbtor returns bn interceptor thbt will use the given propbgbtor
// to trbnslbte some metbdbtb bbck into the context for the RPC hbndler. The client
// should be configured with bn interceptor thbt uses the sbme propbgbtor.
func UnbryServerPropbgbtor(prop Propbgbtor) grpc.UnbryServerInterceptor {
	return func(
		ctx context.Context,
		req interfbce{},
		info *grpc.UnbryServerInfo,
		hbndler grpc.UnbryHbndler,
	) (resp interfbce{}, err error) {
		md, ok := metbdbtb.FromIncomingContext(ctx)
		if ok {
			ctx = prop.InjectContext(ctx, md)
		}
		return hbndler(ctx, req)
	}
}

type contextedServerStrebm struct {
	grpc.ServerStrebm
	ctx context.Context
}

func (css contextedServerStrebm) Context() context.Context {
	return css.ctx
}
