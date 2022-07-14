package userip

import (
	"net/http"
	"strings"
)

const (
	// Sourcegraph-specific user IP header key
	headerKeyUserIP = "X-Sourcegraph-User-IP"
	// De-facto standard for identifying original IP address of a client:
	// https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/X-Forwarded-For
	headerKeyForwardedFor = "X-Forwarded-For"
)

// HTTPTransport is a roundtripper that sets user IP information within request context as
// headers on outgoing requests. The attached headers can be picked up and attached to
// incoming request contexts with userip.HTTPMiddleware.
type HTTPTransport struct {
	RoundTripper http.RoundTripper
}

var _ http.RoundTripper = &HTTPTransport{}

func (t *HTTPTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if t.RoundTripper == nil {
		t.RoundTripper = http.DefaultTransport
	}

	userIP := FromContext(req.Context())
	if userIP != nil {
		req.Header.Set(headerKeyUserIP, userIP.IP)
		req.Header.Set(headerKeyForwardedFor, userIP.XForwardedFor)
	}

	return t.RoundTripper.RoundTrip(req)
}

// HTTPMiddleware wraps the given handle func and attaches user IP data indicated in
// incoming requests to the request header.
func HTTPMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		var userIP UserIP
		userIP.IP = strings.Split(req.RemoteAddr, ":")[0]
		userIP.XForwardedFor = req.Header.Get(headerKeyForwardedFor)

		ctx := req.Context()
		ctxWithIP := WithUserIP(ctx, &userIP)

		next.ServeHTTP(rw, req.WithContext(ctxWithIP))
	})
}
