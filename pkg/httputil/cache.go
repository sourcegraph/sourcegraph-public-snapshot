package httputil

import (
	"net/http"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/rcache"

	"github.com/sourcegraph/httpcache"
)

var (
	// Cache is a HTTP cache backed by Redis
	Cache = rcache.NewByteCache("http")

	// CachingClient is an HTTP client that caches responses backed by
	// Redis (using Cache).
	CachingClient = &http.Client{Transport: httpcache.NewTransport(Cache)}
)
