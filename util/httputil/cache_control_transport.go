package httputil

import (
	"net/http"
	"strings"
	"sync"
)

// CacheControlTransport implements http.RoundTripper and sets the cache-control header of
// all GET requests to CacheControl. Typically, CacheControl is a value inherited from the
// HTTP request of the upstream client. If CacheControl is empty, then cacheControlTransport
// just forwards to the underlying RoundTripper.
//
// As an optimization, CacheControlTransport keeps track of GET URLs it has issued. For
// URLs that it has previously GOT, it does not set the CacheControl header, preferring to
// hit the local cache (if such a cache exists). Note that this behavior is correct
// assuming that all cache entries are fresh within the lifetime of CacheControlTransport.
// For this reason, CacheControlTransport should not outlive the lifetime of a single
// end-user request.
type CacheControlTransport struct {
	// Transport is the base transport that CacheControlTransport forwards requests to
	// after setting the Cache-Control header.
	Transport http.RoundTripper

	// The Cache-Control header to set on requests.
	CacheControl string

	// ShouldForwardCacheControl returns whether the Cache-Control header should be set on
	// a given request. This function should be stateless. Setting this to nil is
	// equivalent to setting a function that always returns true.
	ShouldSetCacheControl func(req *http.Request) bool

	// prevGets records previously seen GET request URLs.
	prevGets map[string]bool

	lock sync.Mutex // protects prevGets
}

func NewCacheControlTransport(cacheControl string, baseTransport http.RoundTripper, shouldSetCacheControl func(req *http.Request) bool) *CacheControlTransport {
	return &CacheControlTransport{
		CacheControl:          cacheControl,
		Transport:             baseTransport,
		ShouldSetCacheControl: shouldSetCacheControl,

		prevGets: make(map[string]bool),
	}
}

func (c *CacheControlTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	transport := c.Transport
	if transport == nil {
		transport = http.DefaultTransport
	}

	if c.CacheControl == "" || strings.ToUpper(req.Method) != "GET" {
		return transport.RoundTrip(req)
	}

	if c.ShouldSetCacheControl != nil && !c.ShouldSetCacheControl(req) {
		return transport.RoundTrip(req)
	}

	// don't set cache header if we've already seen this request
	c.lock.Lock()
	_, seen := c.prevGets[req.URL.String()]
	if !seen {
		c.prevGets[req.URL.String()] = true
		req = CloneRequest(req)
		req.Header.Set("Cache-Control", c.CacheControl)
	}
	c.lock.Unlock()
	return transport.RoundTrip(req)
}
