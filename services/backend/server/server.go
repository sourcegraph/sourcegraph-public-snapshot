package server

import (
	"context"

	"google.golang.org/grpc"
	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/services/backend/internal/middleware/inner"
	"sourcegraph.com/sourcegraph/sourcegraph/services/backend/internal/middleware/outer"
	"sourcegraph.com/sourcegraph/sourcegraph/services/svc"
)

// New creates a new gRPC server with all RPC services
// registered. Callers are responsible for setting up the listener and
// any other server options.
func New(svcs svc.Services, opts ...grpc.ServerOption) *grpc.Server {
	opts = append(opts, grpc.CustomCodec(sourcegraph.GRPCCodec))
	s := grpc.NewServer(opts...)
	svc.RegisterAll(s, svcs)
	return s
}

func Config(ctxFunc func(context.Context) context.Context) svc.Services {
	// Construct the inner services. The inner services are the
	// services as they appear in the context of service method
	// implementations. Below we wrap them with metadata, config,
	// etc., handlers that only need to be run once per external
	// request.
	services := inner.Services()

	// Wrap in middleware for context initialization. This is the
	// outermost wrapper because it performs the most
	// expensive work, and we only want it to be run once per external
	// request (it does not need to be re-run when services make
	// internal requests to their own methods or other services'
	// methods).
	return outer.Services(ctxFunc, services)
}
