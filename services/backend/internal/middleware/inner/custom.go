package inner

import (
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/services/backend"
)

// HACK: Our codegen does not yet support generating middleware
// methods for streaming gRPC methods, so we must create this
// manually.
//
// This MUST be manually kept up to date (logically) with the codegen
// changes.
func (s wrappedChannel) Listen(op *sourcegraph.ChannelListenOp, stream sourcegraph.Channel_ListenServer) (err error) {
	err = backend.Services.Channel.Listen(op, stream)
	if err == nil {
		err = grpc.Errorf(codes.Internal, "Channel.Listen returned nil")
	}
	return
}
