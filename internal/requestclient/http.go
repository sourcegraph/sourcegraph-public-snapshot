package requestclient

import (
	"net/http"
	"strings"
)

const (
	// Sourcegraph-specific client IP header key
	headerKeyClientIP = "X-Sourcegraph-Client-IP"
	// De-facto standard for identifying original IP address of a client:
	// https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/X-Forwarded-For
	headerKeyForwardedFor = "X-Forwarded-For"
)

// HTTPTransport is a roundtripper that sets client IP information within request context as
// headers on outgoing requests. The attached headers can be picked up and attached to
// incoming request contexts with client.HTTPMiddleware.
type HTTPTransport struct {
	RoundTripper http.RoundTripper
}

var _ http.RoundTripper = &HTTPTransport{}

func (t *HTTPTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if t.RoundTripper == nil {
		t.RoundTripper = http.DefaultTransport
	}

	client := FromContext(req.Context())
	if client != nil {
		req.Header.Set(headerKeyClientIP, client.IP)
		req.Header.Set(headerKeyForwardedFor, client.ForwardedFor)
	}

	return t.RoundTripper.RoundTrip(req)
}

// HTTPMiddleware wraps the given handle func and attaches client IP data indicated in
// incoming requests to the request header.
func HTTPMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		ctxWithClient := WithClient(req.Context(), &Client{
			IP:           strings.Split(req.RemoteAddr, ":")[0],
			ForwardedFor: req.Header.Get(headerKeyForwardedFor),
		})
		next.ServeHTTP(rw, req.WithContext(ctxWithClient))
	})
}
