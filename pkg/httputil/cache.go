package httputil

import (
	"net/http"

	"github.com/sourcegraph/httpcache"
)

var (
	// Cache is an in-memory HTTP cache.
	Cache = httpcache.NewMemoryCache()

	// CachingClient is an HTTP client that caches responses in memory
	// (using Cache).
	CachingClient = &http.Client{Transport: httpcache.NewTransport(Cache)}
)
