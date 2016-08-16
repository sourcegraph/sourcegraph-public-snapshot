// Package appdashctx stores and retrieves Appdash config from the
// context.
//
// It is a separate package from traceutil to avoid an import cycle
// between sgx, traceutil, and app.
package appdashctx

import (
	"net/url"

	"gopkg.in/inconshreveable/log15.v2"

	"sourcegraph.com/sourcegraph/appdash"

	"context"
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
	url := AppdashURLSafe(ctx)
	if url == nil {
		panic("no Appdash URL in context")
	}
	return url
}

// AppdashURLSafe does not panic in the case of the url missing. This should
// only be used in top-level error handling routines, not normal code.
func AppdashURLSafe(ctx context.Context) *url.URL {
	url, _ := ctx.Value(urlKey).(*url.URL)
	if url == nil {
		log15.Crit("no Appdash URL in context")
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
