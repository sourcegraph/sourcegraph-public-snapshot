package cli

import "src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"

// Client returns a Sourcegraph API client configured to use the
// specified endpoints and authentication info.
func Client() *sourcegraph.Client {
	return sourcegraph.NewClientFromContext(Ctx)
}
