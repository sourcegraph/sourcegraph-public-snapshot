package handlerutil

import (
	"fmt"
	"net/http"

	"golang.org/x/net/context"

	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/util/httputil/httpctx"
)

// Client returns the Sourcegraph API client and context for an HTTP
// handler to use to respond to an HTTP request.
func Client(r *http.Request) (context.Context, *sourcegraph.Client) {
	ctx := httpctx.FromRequest(r)
	cl, err := sourcegraph.NewClientFromContext(ctx)
	if err != nil {
		panic(fmt.Sprintf("NewClientFromContext: %s", err))
	}
	return ctx, cl
}
