package cli

import (
	"log"
	"net/url"

	"golang.org/x/net/context"
	"src.sourcegraph.com/sourcegraph/auth/userauth"
	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
)

var Endpoint EndpointOpts

func init() {
	CLI.AddGroup("Client API endpoint", "", &Endpoint)
}

// EndpointOpts sets the URL to use when contacting the Sourcegraph
// server.
//
// This endpoint may differ from the endpoint that a server wishes to
// advertise externally. For example, if internal API traffic is
// routed through a local network, you may want to use
// "http://10.1.2.3:3080" as the endpoint here, but you may want
// external clients to use "https://example.com:443". To do so, run
// the server with `src --endpoint http://10.1.2.3:3080 serve
// --app-url https://example.com`.
type EndpointOpts struct {
	// URL is the raw endpoint URL. Callers should almost never use
	// this; use the URLOrDefault method instead.
	URL string `short:"u" long:"endpoint" description:"URL to Sourcegraph server" default:"" env:"SRC_URL"`
}

// NewContext sets the server endpoint in the context.
func (c *EndpointOpts) NewContext(ctx context.Context) context.Context {
	return sourcegraph.WithGRPCEndpoint(ctx, c.URLOrDefault())
}

// URLOrDefault returns c.Endpoint as a *url.URL but with various
// modifications (e.g. a sensible default, no path component, etc). It
// is also responsible for erroring out when the user provides a
// garbage endpoint URL. Always use c.URLOrDefault instead of
// c.Endpoint, even when you just need a string form (just call
// URLOrDefault().String()).
func (c *EndpointOpts) URLOrDefault() *url.URL {
	e := c.URL
	if e == "" {
		// The user did not explicitly specify a endpoint URL, so use the default
		// found in the auth file.
		userAuth, err := userauth.Read(Credentials.AuthFile)
		if err != nil {
			log.Fatal(err, "failed to read user auth file (in EndpointOpts.URLOrDefault)")
		}
		e, _ = userAuth.GetDefault()
		if e == "" {
			// Auth file has no default, so just choose a sensible default value
			// instead.
			e = "http://localhost:3080"
		}
	}
	endpoint, err := url.Parse(e)
	if err != nil {
		log.Fatal(err, "invalid endpoint URL specified (in EndpointOpts.URLOrDefault)")
	}

	// This prevents users who might be using e.g. Sourcegraph under a reverse proxy
	// at mycompany.com/sourcegraph (a subdirectory) from logging in -- but this
	// is not a typical case and otherwise users who effectively run:
	//
	//  src --endpoint=http://localhost:3080 login
	//
	// will be unable to authenticate in the event that they add a slash suffix:
	//
	//  src --endpoint=http://localhost:3080/ login
	//
	endpoint.Path = ""

	if endpoint.Scheme == "" {
		log.Fatal("invalid endpoint URL specified, endpoint URL must start with a schema (e.g. http://mydomain.com)")
	}
	return endpoint
}
