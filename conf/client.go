package conf

import (
	"log"
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
	// Note: intentional use of c.Endpoint instead of EndpointURL here. We want to
	// check if the user input an endpoint URL that we should discover from -- we
	// don't want the default value returned by EndpointURL.
	if c.Endpoint != "" {
		info, err := discover.SiteURL(ctx, c.EndpointURL().String())
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

// EndpointURL returns c.Endpoint as a *url.URL but with various modifications
// (e.g. a sensible default, no path component, etc). It is also responsible for
// erroring out when the user provides a garbage endpoint URL. Always use
// c.EndpointURL instead of c.Endpoint, even when you just need a string form
// (just call EndpointURL().String()).
func (c *EndpointOpts) EndpointURL() *url.URL {
	e := c.Endpoint
	if e == "" {
		e = "http://localhost:3000"
	}
	endpoint, err := url.Parse(e)
	if err != nil {
		log.Fatal(err, "invalid endpoint URL specified (in EndpointOpts.EndpointURL")
	}

	// This prevents users who might be using e.g. Sourcegraph under a reverse proxy
	// at mycompany.com/sourcegraph (a subdirectory) from logging in -- but this
	// is not a typical case and otherwise users who effectively run:
	//
	//  src --endpoint=http://localhost:3000 login
	//
	// will be unable to authenticate in the event that they add a slash suffix:
	//
	//  src --endpoint=http://localhost:3000/ login
	//
	endpoint.Path = ""

	if endpoint.Scheme == "" {
		log.Fatal("invalid endpoint URL specified, endpoint URL must start with a schema (e.g. http://mydomain.com)")
	}
	return endpoint
}
