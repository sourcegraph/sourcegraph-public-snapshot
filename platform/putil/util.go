package putil

import (
	"net/http"

	"golang.org/x/net/context"

	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/sgx/client"
	"src.sourcegraph.com/sourcegraph/util/handlerutil"
	"src.sourcegraph.com/sourcegraph/util/httputil/httpctx"
)

func Context(r *http.Request) context.Context {
	return httpctx.FromRequest(r)
}

// CLIContext returns a minimal context that can be used with the CLI. It comes
// with credentials attached (if existent and valid).
func CLIContext() context.Context {
	return client.Ctx
}

func UserFromRequest(r *http.Request) *sourcegraph.UserSpec {
	return handlerutil.UserFromRequest(r)
}

func UserFromContext(ctx context.Context) *sourcegraph.UserSpec {
	return handlerutil.UserFromContext(ctx)
}
