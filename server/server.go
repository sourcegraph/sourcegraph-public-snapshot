package server

import (
	"strings"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/auth/authutil"
	"src.sourcegraph.com/sourcegraph/errcode"
	"src.sourcegraph.com/sourcegraph/server/internal/middleware"
	"src.sourcegraph.com/sourcegraph/server/internal/middleware/auth"
	"src.sourcegraph.com/sourcegraph/server/internal/middleware/cached"
	"src.sourcegraph.com/sourcegraph/server/internal/middleware/ctxfunc"
	"src.sourcegraph.com/sourcegraph/server/internal/oauth2util"
	"src.sourcegraph.com/sourcegraph/server/local"
	"src.sourcegraph.com/sourcegraph/server/serverctx"
	"src.sourcegraph.com/sourcegraph/svc"
	"src.sourcegraph.com/sourcegraph/util/traceutil"
)

// NewServer creates a new gRPC server with all RPC services
// registered. Callers are responsible for setting up the listener and
// any other server options.
func NewServer(svcs svc.Services, opts ...grpc.ServerOption) *grpc.Server {
	opts = append(opts, grpc.CustomCodec(sourcegraph.GRPCCodec))
	s := grpc.NewServer(opts...)
	svc.RegisterAll(s, svcs)
	return s
}

func Config(ctxFunc func(context.Context) context.Context) svc.Services {
	authConfig := &auth.Config{
		AllowAnonymousReaders: authutil.ActiveFlags.AllowAnonymousReaders,
		DebugLog:              false, /* TODO(sqs:cleanup) globalOpt.Verbose*/
	}

	// Construct the inner services. The inner services are the
	// services as they appear in the context of service method
	// implementations. Below (in the "return middleware.Services"
	// statement) we wrap them with authentication, metadata, config,
	// etc., handlers that only need to be run once per external
	// request.
	services := middleware.Wrap(local.Services, authConfig)

	// Wrap in middleware for context initialization. This is the
	// outermost wrapper (except caching) because it performs the most
	// expensive work, and we only want it to be run once per external
	// request (it does not need to be re-run when services make
	// internal requests to their own methods or other services'
	// methods).
	outerServices := ctxfunc.Services(func(ctx context.Context) (context.Context, error) {
		var err error

		// Initialize from command-line args.
		ctx = ctxFunc(ctx)

		// Propagate span ID for tracing.
		ctx, err = traceutil.MiddlewareGRPC(ctx)
		if err != nil {
			return nil, err
		}

		for _, f := range serverctx.Funcs {
			ctx, err = f(ctx)
			if err != nil {
				return nil, err
			}
		}

		// Set the services in the context so they are available to
		ctx = svc.WithServices(ctx, services)

		// Check for and verify OAuth2 credentials.
		ctx, err = oauth2util.GRPCMiddleware(ctx)
		if err != nil {
			return nil, err
		}

		return ctx, nil
	}, func(err error) error {
		if err == nil {
			return nil
		}

		// Don't double-wrap errors that are already gRPC errors.
		if strings.HasPrefix(err.Error(), "rpc error: code = ") {
			return err
		}

		return grpc.Errorf(errcode.GRPC(err), "%s", err.Error())
	})

	outerServices = cached.Wrap(outerServices)
	return outerServices
}
