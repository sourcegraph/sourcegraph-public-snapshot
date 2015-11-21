package handlerutil

import (
	"net/http"

	"sourcegraph.com/sourcegraph/grpccache"
	"src.sourcegraph.com/sourcegraph/util/httputil/httpctx"
)

// CacheMiddleware propagates the cache headers from the HTTP request by setting
// the cache control key in the gRPC context.
func CacheMiddleware(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	ctx := httpctx.FromRequest(r)

	if hasNoCacheHeader(r) {
		ctx = grpccache.NoCache(ctx)
		httpctx.SetForRequest(r, ctx)
	}

	next(w, r)
}

func hasNoCacheHeader(r *http.Request) bool {
	cacheControl := r.Header.Get("Cache-Control")
	if cacheControl == "no-cache" {
		return true
	}

	// For compatibility with older HTTP 1.0 clients.
	// See http://www.w3.org/Protocols/rfc2616/rfc2616-sec14.html#sec14.32.
	pragma := r.Header.Get("Pragma")
	if pragma == "no-cache" {
		return true
	}

	return false
}
