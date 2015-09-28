package conf

import (
	"net/url"

	"golang.org/x/net/context"
	"sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/fed/discover"
)

// EndpointOpts sets the URLs to use for contacting the Sourcegraph
// server's API.
//
// These endpoints may differ from the endpoints that a server wishes
// to advertise externally. For example, if internal API traffic is
// routed through a local network, you may want to use
// "http://10.1.2.3:3100" as the gRPC endpoint here, but you may want
// external clients to use "https://example.com:3100". The externally
// advertised endpoints may be passed as CLI flags to `src serve`.
type EndpointOpts struct {
	Endpoint string `short:"u" long:"endpoint" description:"URL or hostname of Sourcegraph endpoint for API client operations (auto-detects and overrides --grpc-endpoint)" default:"" env:"SRC_URL"`

	GRPCEndpoint string `long:"grpc-endpoint" description:"URL of gRPC API (use -u/--endpoint to auto-detect)" default:"http://localhost:3100" env:"SG_GRPC_URL"`
}

// WithEndpoints sets the HTTP and gRPC endpoint in the context. If a
// auto-detect endpoint is specified, it discovers the gRPC endpoint
// from that endpoint; otherwise it uses the GRPCEndpoint field.
func (c *EndpointOpts) WithEndpoints(ctx context.Context) (context.Context, error) {
	if c.Endpoint != "" {
		info, err := discover.SiteURL(ctx, c.Endpoint)
		if err != nil {
			return nil, err
		}
		return info.NewContext(ctx)
	}

	grpcEndpoint, err := url.Parse(c.GRPCEndpoint)
	if err != nil {
		return nil, err
	}
	ctx = sourcegraph.WithGRPCEndpoint(ctx, grpcEndpoint)

	return ctx, nil
}
