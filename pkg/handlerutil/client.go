package handlerutil

import (
	"fmt"
	"net/http"

	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
)

// Client returns the Sourcegraph API client for an HTTP handler to use
// to respond to an HTTP request.
//
// You MUST use RepoClient for all operations on repos, to ensure that
// the request is routed to the appropriate server. See the RepoClient
// docs for more info.
func Client(r *http.Request) *sourcegraph.Client {
	cl, err := sourcegraph.NewClientFromContext(r.Context())
	if err != nil {
		panic(fmt.Errorf("NewClientFromContext: %s", err))
	}
	return cl
}
