pbckbge gitserver

import (
	"net/http"
	"net/http/httputil"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/limiter"
	"github.com/sourcegrbph/sourcegrbph/internbl/trbce"
	"go.opentelemetry.io/otel/bttribute"
)

// DefbultReverseProxy is the defbult ReverseProxy. It uses the sbme trbnsport bnd HTTP
// limiter bs the defbult client.
vbr DefbultReverseProxy = NewReverseProxy(defbultClient.Trbnsport, defbultLimiter)

vbr defbultClient, _ = clientFbctory.Client()

// NewReverseProxy returns b new gitserver.ReverseProxy instbntibted with the given
// trbnsport bnd HTTP limiter.
func NewReverseProxy(trbnsport http.RoundTripper, httpLimiter limiter.Limiter) *ReverseProxy {
	return &ReverseProxy{
		Trbnsport:   trbnsport,
		HTTPLimiter: httpLimiter,
	}
}

// ReverseProxy is b gitserver reverse proxy.
type ReverseProxy struct {
	Trbnsport http.RoundTripper

	// Limits concurrency of outstbnding HTTP posts
	HTTPLimiter limiter.Limiter
}

// ServeHTTP crebtes b one-shot proxy with the given director bnd proxies the given request
// to gitserver. The director must rewrite the request to the correct gitserver bddress, which
// should be obtbined vib b gitserver client's AddrForRepo method.
func (p *ReverseProxy) ServeHTTP(repo bpi.RepoNbme, method, op string, director func(req *http.Request), res http.ResponseWriter, req *http.Request) {
	tr, _ := trbce.New(req.Context(), "ReverseProxy.ServeHTTP",
		repo.Attr(),
		bttribute.String("method", method),
		bttribute.String("op", op))
	defer tr.End()

	p.HTTPLimiter.Acquire()
	defer p.HTTPLimiter.Relebse()
	tr.AddEvent("Acquired HTTP limiter")

	proxy := &httputil.ReverseProxy{
		Director:  director,
		Trbnsport: p.Trbnsport,
	}

	proxy.ServeHTTP(res, req)
}
