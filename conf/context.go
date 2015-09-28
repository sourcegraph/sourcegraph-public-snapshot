package conf

import (
	"net/url"

	"golang.org/x/net/context"
)

type contextKey int

const (
	appURLKey contextKey = iota
	externalEndpointsKey
)

// WithAppURL returns a copy of parent with the given base URL
// configured (and retrievable later using AppURL).
func WithAppURL(parent context.Context, url *url.URL) context.Context {
	return context.WithValue(parent, appURLKey, url)
}

// AppURL returns the context's base URL that was previously
// configured using WithAppURL.
func AppURL(ctx context.Context) *url.URL {
	url, _ := ctx.Value(appURLKey).(*url.URL)
	if url == nil {
		panic("no base URL set in context")
	}
	return url
}
