package outer

import (
	"strings"

	"golang.org/x/net/context"

	"google.golang.org/grpc"
	"src.sourcegraph.com/sourcegraph/errcode"
	"src.sourcegraph.com/sourcegraph/server/internal/oauth2util"
	"src.sourcegraph.com/sourcegraph/server/serverctx"
	"src.sourcegraph.com/sourcegraph/svc"
	"src.sourcegraph.com/sourcegraph/util/traceutil"
)

// A ContextFunc is called before a method executes and lets you customize its context.
type ContextFunc func(context.Context) context.Context

func initContext(ctx context.Context, ctxFunc ContextFunc, services svc.Services) (context.Context, error) {
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
}

func wrapErr(err error) error {
	if err == nil {
		return nil
	}

	// Don't double-wrap errors that are already gRPC errors.
	if strings.HasPrefix(err.Error(), "rpc error: code = ") {
		return err
	}

	return grpc.Errorf(errcode.GRPC(err), "%s", err.Error())
}
