package fed

import (
	"net/url"

	"golang.org/x/net/context"
	"sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/svc"
	"src.sourcegraph.com/sourcegraph/svc/middleware/remote"
)

// NewRemoteContext creates a copy of ctx that accesses services on
// the given endpoint.
func NewRemoteContext(ctx context.Context, endpoint *url.URL) context.Context {
	// TODO: Temporary workaround until sourcegraph.com is updated.
	if endpoint.Host == "sourcegraph.com" || endpoint.Host == "sourcegraph.com:443" {
		tmp := *endpoint
		tmp.Host = "sourcegraph.com:3100"
		endpoint = &tmp
	}

	ctx = sourcegraph.WithGRPCEndpoint(ctx, endpoint)
	return svc.WithServices(ctx, remote.Services)
}
