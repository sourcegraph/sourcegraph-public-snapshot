package handlerutil

import (
	"fmt"
	"net/http"

	"github.com/sourcegraph/mux"

	"golang.org/x/net/context"

	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/util/httputil/httpctx"
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

// RepoClient returns the (possibly federated) Sourcegraph API client
// and context for an HTTP handler to use when performing operations
// on a repo.
//
// It MUST be used in order to route the API requests to the fallback
// endpoint if the main server does not have the repo. (E.g., if a
// user tries to access a public GitHub repo on a Sourcegraph Server
// installation that doesn't have that repo locally, the operation
// will fall back to the fallback endpoint (usually Sourcegraph.com),
// which has the repo.)
func RepoClient(r *http.Request) (ctx context.Context, cl *sourcegraph.Client, fallback bool, err error) {
	repo, err := sourcegraph.UnmarshalRepoSpec(mux.Vars(r))
	if err != nil {
		return nil, nil, false, err
	}

	ctx = httpctx.FromRequest(r)
	return repoClient(ctx, repo)
}

func repoClient(ctx context.Context, repo sourcegraph.RepoSpec) (ctx2 context.Context, cl *sourcegraph.Client, fallback bool, err error) {
	ctx, fallback, err = ctxForRepo(ctx, repo)
	if err != nil {
		return nil, nil, fallback, err
	}

	cl, err = sourcegraph.NewClientFromContext(ctx)
	return ctx, cl, fallback, err
}
