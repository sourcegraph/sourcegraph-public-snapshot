package requestclient

import (
	"net/http"
	"strings"
)

const (
	// Sourcegraph-specific client IP header key
	headerKeyClientIP = "X-Sourcegraph-Client-IP"
	// Sourcegraph-specific client-provided request identifier key
	headerKeyRequestID = "X-Sourcegraph-Request-ID"
	// De-facto standard for identifying original IP address of a client:
	// https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/X-Forwarded-For
	headerKeyForwardedFor = "X-Forwarded-For"
	// Standard for identifyying the application, operating system, vendor,
	// and/or version of the requesting user agent.
	// https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/User-Agent
	headerKeyUserAgent = "User-Agent"
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
		req.Header.Set(headerKeyRequestID, client.RequestID)
	}

	return t.RoundTripper.RoundTrip(req)
}

// ExternalHTTPMiddleware wraps the given handle func and attaches client IP
// data indicated in incoming requests to the request header.
//
// This is meant to be used by http handlers which sit behind a reverse proxy
// receiving user traffic. IE sourcegraph-frontend.
func ExternalHTTPMiddleware(next http.Handler, hasCloudflareProxy bool) http.Handler {
	return httpMiddleware(next, hasCloudflareProxy)
}

// InternalHTTPMiddleware wraps the given handle func and attaches client IP
// data indicated in incoming requests to the request header.
//
// This is meant to be used by http handlers which receive internal traffic.
// EG gitserver.
func InternalHTTPMiddleware(next http.Handler) http.Handler {
	return httpMiddleware(next, false)
}

// httpMiddleware wraps the given handle func and attaches client IP data indicated in
// incoming requests to the request header.
func httpMiddleware(next http.Handler, hasCloudflareProxy bool) http.Handler {
	forwardedForHeaders := []string{headerKeyForwardedFor}
	if hasCloudflareProxy {
		// On Sourcegraph.com we have a more reliable header from cloudflare,
		// since x-forwarded-for can be spoofed. So use that if available.
		forwardedForHeaders = []string{"Cf-Connecting-Ip", headerKeyForwardedFor}
	}

	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		forwardedFor := ""
		for _, k := range forwardedForHeaders {
			forwardedFor = req.Header.Get(k)
			if forwardedFor != "" {
				break
			}
		}

		ctxWithClient := WithClient(req.Context(), &Client{
			IP:           strings.Split(req.RemoteAddr, ":")[0],
			ForwardedFor: req.Header.Get(headerKeyForwardedFor),
			UserAgent:    req.Header.Get(headerKeyUserAgent),
			RequestID:    req.Header.Get(headerKeyRequestID),
		})
		next.ServeHTTP(rw, req.WithContext(ctxWithClient))
	})
}
