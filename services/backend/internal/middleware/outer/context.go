package outer

import (
	"context"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"gopkg.in/inconshreveable/log15.v2"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/errcode"
	"sourcegraph.com/sourcegraph/sourcegraph/services/backend/internal/oauth2util"
	"sourcegraph.com/sourcegraph/sourcegraph/services/backend/serverctx"
	"sourcegraph.com/sourcegraph/sourcegraph/services/svc"
)

// A ContextFunc is called before a method executes and lets you customize its context.
type ContextFunc func(context.Context) context.Context

func initContext(ctx context.Context, ctxFunc ContextFunc, services svc.Services) (context.Context, error) {
	var err error

	// Initialize from command-line args.
	ctx = ctxFunc(ctx)

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

	// Run ctx funcs that (may) require use of values we just stored
	// in the ctx.
	for _, f := range serverctx.LastFuncs {
		ctx, err = f(ctx)
		if err != nil {
			return nil, err
		}
	}

	return ctx, nil
}

// wrapErr returns non-nil error iff err is non-nil.
func wrapErr(err error) error {
	if err == nil {
		return nil
	}

	// Don't double-wrap errors that are already gRPC errors.
	if strings.HasPrefix(err.Error(), "rpc error: code = ") {
		return err
	}

	code := errcode.GRPC(err)
	if code == codes.OK {
		// grpc.Errorf returns nil error if code is OK, so replace it with unknown if we get here.
		log15.Warn("wrapErr: err resulted in errcode.GRPC(err) returning codes.OK; using codes.Unknown", "err", err)
		code = codes.Unknown
	}
	return grpc.Errorf(code, "%s", err.Error())
}
