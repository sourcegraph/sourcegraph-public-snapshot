package outer

import (
	"runtime"

	"golang.org/x/net/context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/services/svc"
)

// HACK: Our codegen does not yet support generating middleware
// methods for streaming gRPC methods, so we must create this
// manually.
//
// This MUST be manually kept up to date (logically) with the codegen
// changes.
func (s wrappedChannel) Listen(op *sourcegraph.ChannelListenOp, stream sourcegraph.Channel_ListenServer) (returnedError error) {
	defer func() {
		if err := recover(); err != nil {
			const size = 64 << 10
			buf := make([]byte, size)
			buf = buf[:runtime.Stack(buf, false)]
			returnedError = grpc.Errorf(codes.Internal, "panic in Channel.Listen: %v\n\n%s", err, buf)
		}
	}()

	ctx := stream.Context()

	var err error
	ctx, err = initContext(ctx, s.ctxFunc, s.services)
	if err != nil {
		return wrapErr(err)
	}

	innerSvc := svc.ChannelOrNil(ctx)
	if innerSvc == nil {
		return grpc.Errorf(codes.Unimplemented, "Channel")
	}

	if err := innerSvc.Listen(op, channel_listenServer{stream, ctx}); err != nil {
		return wrapErr(err)
	}

	return nil
}

// channel_ListenServer lets us override the ctx to be the modified ctx
// from initContext.
type channel_listenServer struct {
	sourcegraph.Channel_ListenServer
	ctx context.Context
}

// Context overrides grpc.Stream's Context method.
func (s channel_listenServer) Context() context.Context { return s.ctx }
