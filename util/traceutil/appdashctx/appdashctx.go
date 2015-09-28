// Package appdashctx stores and retrieves Appdash config from the
// context.
//
// It is a separate package from traceutil to avoid an import cycle
// between sgx, traceutil, and app.
package appdashctx

import (
	"net/url"

	"sourcegraph.com/sourcegraph/appdash"

	"golang.org/x/net/context"
)

type contextKey int

const (
	urlKey contextKey = iota
	collectorKey
)

func WithAppdashURL(ctx context.Context, url *url.URL) context.Context {
	return context.WithValue(ctx, urlKey, url)
}

func AppdashURL(ctx context.Context) *url.URL {
	url, _ := ctx.Value(urlKey).(*url.URL)
	if url == nil {
		panic("no Appdash URL in context")
	}
	return url
}

func WithCollector(ctx context.Context, c appdash.Collector) context.Context {
	return context.WithValue(ctx, collectorKey, c)
}

func Collector(ctx context.Context) appdash.Collector {
	c, _ := ctx.Value(collectorKey).(appdash.Collector)
	return c
}
