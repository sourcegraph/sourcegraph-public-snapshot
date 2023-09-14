package gitserver

import (
	"net/http"
	"net/http/httputil"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/limiter"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	"go.opentelemetry.io/otel/attribute"
)

// DefaultReverseProxy is the default ReverseProxy. It uses the same transport and HTTP
// limiter as the default client.
var DefaultReverseProxy = NewReverseProxy(defaultClient.Transport, defaultLimiter)

var defaultClient, _ = clientFactory.Client()

// NewReverseProxy returns a new gitserver.ReverseProxy instantiated with the given
// transport and HTTP limiter.
func NewReverseProxy(transport http.RoundTripper, httpLimiter limiter.Limiter) *ReverseProxy {
	return &ReverseProxy{
		Transport:   transport,
		HTTPLimiter: httpLimiter,
	}
}

// ReverseProxy is a gitserver reverse proxy.
type ReverseProxy struct {
	Transport http.RoundTripper

	// Limits concurrency of outstanding HTTP posts
	HTTPLimiter limiter.Limiter
}

// ServeHTTP creates a one-shot proxy with the given director and proxies the given request
// to gitserver. The director must rewrite the request to the correct gitserver address, which
// should be obtained via a gitserver client's AddrForRepo method.
func (p *ReverseProxy) ServeHTTP(repo api.RepoName, method, op string, director func(req *http.Request), res http.ResponseWriter, req *http.Request) {
	tr, _ := trace.New(req.Context(), "ReverseProxy.ServeHTTP",
		repo.Attr(),
		attribute.String("method", method),
		attribute.String("op", op))
	defer tr.End()

	p.HTTPLimiter.Acquire()
	defer p.HTTPLimiter.Release()
	tr.AddEvent("Acquired HTTP limiter")

	proxy := &httputil.ReverseProxy{
		Director:  director,
		Transport: p.Transport,
	}

	proxy.ServeHTTP(res, req)
}
