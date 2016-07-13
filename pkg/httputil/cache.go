package httputil

import (
	"net/http"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/rcache"

	"github.com/sourcegraph/httpcache"
)

var (
	// Cache is a HTTP cache backed by Redis. The TTL of a week is a
	// balance between caching values for a useful amount of time versus
	// growing the cache too large.
	Cache = rcache.NewByteCache("http", 604800)

	// CachingClient is an HTTP client that caches responses backed by
	// Redis (using Cache).
	CachingClient = &http.Client{Transport: httpcache.NewTransport(Cache)}
)
