package fed

import (
	"net/url"

	"golang.org/x/net/context"
	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/svc"
	"src.sourcegraph.com/sourcegraph/svc/middleware/remote"
)

// NewRemoteContext creates a copy of ctx that accesses services on
// the given endpoint.
func NewRemoteContext(ctx context.Context, endpoint *url.URL) context.Context {
	ctx = sourcegraph.WithGRPCEndpoint(ctx, endpoint)
	return svc.WithServices(ctx, remote.Services)
}
