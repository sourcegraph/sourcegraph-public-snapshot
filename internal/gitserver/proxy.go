package gitserver

import (
	"net/http"
	"net/http/httputil"

	"github.com/neelance/parallel"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/trace/ot"
)

// DefaultReverseProxy is the default ReverseProxy. It uses the same transport and HTTP
// limiter as the default client.
var DefaultReverseProxy = NewReverseProxy(defaultTransport, DefaultClient.HTTPLimiter)

// NewReverseProxy returns a new gitserver.ReverseProxy instantiated with the given
// transport and HTTP limiter.
func NewReverseProxy(transport http.RoundTripper, httpLimiter *parallel.Run) *ReverseProxy {
	return &ReverseProxy{
		Transport:   transport,
		HTTPLimiter: httpLimiter,
	}
}

// ReverseProxy is a gitserver reverse proxy.
type ReverseProxy struct {
	Transport http.RoundTripper

	// Limits concurrency of outstanding HTTP posts
	HTTPLimiter *parallel.Run
}

// ServeHTTP creates a one-shot proxy with the given director and proxies the given request
// to gitserver. The director must rewrite the request to the correct gitserver address, which
// should be obtained via a gitserver client's AddrForRepo method.
func (p *ReverseProxy) ServeHTTP(repo api.RepoName, method, op string, director func(req *http.Request), res http.ResponseWriter, req *http.Request) {
	span, _ := ot.StartSpanFromContext(req.Context(), "ReverseProxy.ServeHTTP")
	defer func() {
		span.LogKV("repo", string(repo), "method", method, "op", op)
		span.Finish()
	}()

	if p.HTTPLimiter != nil {
		p.HTTPLimiter.Acquire()
		defer p.HTTPLimiter.Release()
		span.LogKV("event", "Acquired HTTP limiter")
	}

	proxy := &httputil.ReverseProxy{
		Director:  director,
		Transport: p.Transport,
	}

	proxy.ServeHTTP(res, req)
}
