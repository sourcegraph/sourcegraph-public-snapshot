package requestclient

import (
	"net/http"
	"strings"

	"github.com/sourcegraph/sourcegraph/internal/env"
)

var useCloudflareHeaders = env.MustGetBool("SRC_USE_CLOUDFLARE_HEADERS", false, "Use Cloudflare headers for request metadata")

const (
	// Sourcegraph-specific client IP header key
	headerKeyClientIP = "X-Sourcegraph-Client-IP"
	// De-facto standard for identifying original IP address of a client:
	// https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/X-Forwarded-For
	headerKeyForwardedFor = "X-Forwarded-For"
	// headerKeyForwardedForUserAgent propagates the first headerKeyUserAgent
	// seen.
	headerKeyForwardedForUserAgent = "X-Forwarded-For-User-Agent"
	// Standard for identifyying the application, operating system, vendor,
	// and/or version of the requesting user agent.
	// https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/User-Agent
	headerKeyUserAgent = "User-Agent"
)

// HTTPTransport is a roundtripper that sets client IP information within request context as
// headers on outgoing requests. The attached headers can be picked up and attached to
// incoming request contexts with client.HTTPMiddleware.
//
// TODO(@bobheadxi): Migrate to httpcli.Doer and httpcli.Middleware
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
		forwardedForUserAgent := client.ForwardedForUserAgent
		if forwardedForUserAgent == "" {
			forwardedForUserAgent = client.UserAgent
		}
		req = req.Clone(req.Context()) // RoundTripper should not modify original request
		req.Header.Set(headerKeyClientIP, client.IP)
		req.Header.Set(headerKeyForwardedFor, client.ForwardedFor)
		req.Header.Set(headerKeyForwardedForUserAgent, forwardedForUserAgent)
	}

	return t.RoundTripper.RoundTrip(req)
}

// ExternalHTTPMiddleware wraps the given handle func and attaches client IP
// data indicated in incoming requests to the request header.
//
// This is meant to be used by http handlers which sit behind a reverse proxy
// receiving user traffic. IE sourcegraph-frontend.
func ExternalHTTPMiddleware(next http.Handler) http.Handler {
	return httpMiddleware(next, useCloudflareHeaders)
}

// InternalHTTPMiddleware wraps the given handle func and attaches client IP
// data indicated in incoming requests to the request header.
//
// This is meant to be used by http handlers which receive internal traffic.
// EG gitserver.
func InternalHTTPMiddleware(next http.Handler) http.Handler {
	// We don't use Cloudflare headers in internal handlers because we don't have
	// access to the Cloudflare WAF.
	return httpMiddleware(next, false)
}

// httpMiddleware wraps the given handle func and attaches client IP data indicated in
// incoming requests to the request header.
//
// In Sourcegraph.com and Sourcegraph Cloud, we have a more reliable headers from
// Cloudflare via Cloudflare WAF, so environments that use Cloudflare can opt-in
// to Cloudflare-provided headers on external handlers. We have a debug endpoint
// that lets you confirm the presence of various headers:
//
//	curl --silent https://sourcegraph.com/-/debug/headers | grep Cf-
//
// Documentation for available Cloudflare headers is available at
// https://developers.cloudflare.com/fundamentals/reference/http-request-headers
func httpMiddleware(next http.Handler, useCloudflareHeaders bool) http.Handler {
	ipSourceFuncs := []ipSourceFunc{
		getForwardedFor,
	}

	if useCloudflareHeaders {
		// Try to find trusted Cloudflare-provided connecting client IP.
		// https://developers.cloudflare.com/fundamentals/reference/http-request-headers/#x-forwarded-for
		ipSourceFuncs = []ipSourceFunc{getCloudFlareIP, getForwardedFor}
	}

	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		forwardedFor := ""
		for _, f := range ipSourceFuncs {
			forwardedFor = f(req)
			if forwardedFor != "" {
				break
			}
		}

		var wafIPCountryCode string
		if useCloudflareHeaders {
			// Try to find trusted Cloudflare-provided country code of the request.
			// https://developers.cloudflare.com/fundamentals/reference/http-request-headers/#cf-ipcountry
			//
			// Cloudflare uses the "XX" code to indicate that the country info
			// is unknown.
			if cfIPCountry := req.Header.Get("CF-IPCountry"); cfIPCountry != "" && cfIPCountry != "XX" {
				wafIPCountryCode = cfIPCountry
			}
		}

		currentUserAgent := req.Header.Get(headerKeyUserAgent)
		forwardedForUserAgent := currentUserAgent
		if agent := req.Header.Get(headerKeyForwardedForUserAgent); agent != "" {
			forwardedForUserAgent = agent
		}

		ctxWithClient := WithClient(req.Context(), &Client{
			IP:                    strings.Split(req.RemoteAddr, ":")[0],
			ForwardedFor:          forwardedFor,
			UserAgent:             currentUserAgent,
			ForwardedForUserAgent: forwardedForUserAgent,

			wafIPCountryCode: wafIPCountryCode,
		})
		next.ServeHTTP(rw, req.WithContext(ctxWithClient))
	})
}

type ipSourceFunc func(*http.Request) string

func getCloudFlareIP(req *http.Request) string {
	return req.Header.Get("Cf-Connecting-Ip")
}

func getForwardedFor(req *http.Request) string {
	// According to HTTP1.1/RFC 2616: Headers may be repeated, and any comma-separated
	// list-headers (like X-Forwarded-For) should be treated as a single value.
	//
	// "...It MUST be possible to combine the multiple header fields into one “field-name: field-value”
	// pair, without changing the semantics of the message, by appending each subsequent field-value to
	// the first, each separated by a comma. The order in which header fields with the same field-name
	// are received is therefore significant to the interpretation of the combined field value, and thus
	// a proxy MUST NOT change the order of these field values when a message is forwarded."
	values := req.Header.Values(headerKeyForwardedFor)
	if len(values) == 0 {
		return ""
	}

	return strings.Join(values, ",")
}
