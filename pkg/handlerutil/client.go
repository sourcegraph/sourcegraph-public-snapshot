package handlerutil

import (
	"fmt"
	"net/http"

	"context"

	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/httputil/httpctx"
)

// Client returns the Sourcegraph API client and context for an HTTP
// handler to use to respond to an HTTP request.
//
// You MUST use RepoClient for all operations on repos, to ensure that
// the request is routed to the appropriate server. See the RepoClient
// docs for more info.
func Client(r *http.Request) (context.Context, *sourcegraph.Client) {
	ctx := httpctx.FromRequest(r)
	cl, err := sourcegraph.NewClientFromContext(ctx)
	if err != nil {
		panic(fmt.Sprintf("NewClientFromContext: %s", err))
	}
	return ctx, cl
}
