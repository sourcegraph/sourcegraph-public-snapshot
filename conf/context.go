package conf

import (
	"net/url"

	"golang.org/x/net/context"
)

type contextKey int

const (
	appURLKey contextKey = iota
)

// WithURL returns a copy of parent with the given base URL
// configured (and retrievable later using AppURL).
func WithURL(parent context.Context, appURL *url.URL) context.Context {
	return context.WithValue(parent, appURLKey, appURL)
}

// AppURL returns the context's base URL that was previously
// configured using WithURL.
func AppURL(ctx context.Context) *url.URL {
	url, _ := ctx.Value(appURLKey).(*url.URL)
	if url == nil {
		panic("no base URL set in context")
	}
	return url
}
