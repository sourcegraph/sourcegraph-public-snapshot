package putil

import (
	"net/http"

	"golang.org/x/net/context"

	"sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/sgx"
	"src.sourcegraph.com/sourcegraph/util/handlerutil"
	"src.sourcegraph.com/sourcegraph/util/httputil/httpctx"
)

func Context(r *http.Request) context.Context {
	return httpctx.FromRequest(r)
}

// CLIContext returns a minimal context that can be used with the CLI. It comes
// with credentials attached (if existent and valid).
func CLIContext() context.Context {
	return sgx.WithClientContext(context.Background())
}

func UserFromRequest(r *http.Request) *sourcegraph.User {
	return handlerutil.UserFromRequest(r)
}

func UserFromContext(ctx context.Context) *sourcegraph.User {
	return handlerutil.UserFromContext(ctx)
}
