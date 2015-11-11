package conf

import (
	"net/url"

	"golang.org/x/net/context"
)

// ExternalEndpointsOpts configures the server's externally advertised
// endpoint URLs for the HTTP API and gRPC API. These are the URLs at
// which external clients should contact the server. The server itself
// (and the app, worker, and HTTP API) communicates with itself using
// the endpoints specified in conf.EndpointOpts.
type ExternalEndpointsOpts struct {
	HTTPEndpoint string `long:"external-http-endpoint" description:"externally accessible base URL to HTTP API" default-mask:"default: <AppURL>/.api/"`
	GRPCEndpoint string `long:"external-grpc-endpoint" description:"externally accessible base URL to gRPC API" default-mask:"default: --grpc-endpoint value"`
}

// HTTPEndpointURL parses and returns o.HTTPEndpoint. If parsing
// fails, it panics (since it is assumed to have been sanitized at
// creation time).
func (o ExternalEndpointsOpts) HTTPEndpointURL() *url.URL {
	u, err := url.Parse(o.HTTPEndpoint)
	if err != nil {
		panic(err)
	}
	return u
}

// GRPCEndpointURL parses and returns o.GRPCEndpoint. If parsing
// fails, it panics (since it is assumed to have been sanitized at
// creation time).
func (o ExternalEndpointsOpts) GRPCEndpointURL() *url.URL {
	u, err := url.Parse(o.GRPCEndpoint)
	if err != nil {
		panic(err)
	}
	return u
}

// WithExternalEndpoints returns a copy of parent with the given
// external endpoints configured (and retrievable later using
// ExternalEndpoints).
//
// Only the server should call this.
func WithExternalEndpoints(parent context.Context, v ExternalEndpointsOpts) context.Context {
	return context.WithValue(parent, externalEndpointsKey, v)
}

// ExternalEndpoints returns the context's external endpoints that
// were previously configured using WithExternalEndpoints. If no
// external endpoints were previously configured, it panics.
//
// Only the server should call this.
func ExternalEndpoints(ctx context.Context) ExternalEndpointsOpts {
	v, ok := ctx.Value(externalEndpointsKey).(ExternalEndpointsOpts)
	if ok {
		return v
	}
	panic("no ExternalEndpoints set")
}
