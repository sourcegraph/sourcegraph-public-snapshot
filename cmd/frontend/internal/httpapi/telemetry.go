package httpapi

import (
	"fmt"
	"net/http"
	"net/http/httputil"

	"github.com/gorilla/mux"

	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/pkg/siteid"
	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/telemetry"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/conf"
)

var telemetryReverseProxy http.Handler

func init() {
	// If telemetry is disabled, we still want to collect samples, so we can show the
	// site admin what *would* be collected if it were enabled.
	if conf.Get().DisableTelemetry {
		telemetryReverseProxy = http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			stripTelemetryRequest(req)
			req.URL.Scheme, req.URL.Host = "https", "example.com" // needed for DumpRequestOut
			telemetry.Sample(req)
			fmt.Fprintln(w, "telemetry is disabled")
		})
	} else {
		telemetryReverseProxy = &httputil.ReverseProxy{
			Director: func(req *http.Request) {
				stripTelemetryRequest(req)
				req.URL.Scheme, req.URL.Host = "https", "example.com" // needed for DumpRequestOut
				telemetry.Sample(req)

				if useBILogger(siteid.Get()) {
					req.URL.Scheme = "http"
					req.URL.Host = req.Host
					req.URL.Path = "/.bi-logger"
				} else {
					req.URL.Scheme = "https"
					req.URL.Host = "sourcegraph-logging.telligentdata.com"
					req.Host = "sourcegraph-logging.telligentdata.com"
					req.URL.Path = "/" + mux.Vars(req)["TelemetryPath"]
				}
			},
		}
	}
}

var keepHeadersInProxiedTelemetry = map[string]struct{}{
	"Host":            struct{}{},
	"User-Agent":      struct{}{},
	"Content-Length":  struct{}{},
	"Accept":          struct{}{},
	"Accept-Encoding": struct{}{},
	"Accept-Language": struct{}{},
	"Cache-Control":   struct{}{},
	"Connection":      struct{}{},
	"Content-Type":    struct{}{},
	"Pragma":          struct{}{},
}

// stripTelemetryRequest removes sensitive and unnecessary data from the client request
// before forwarding it up to the telemetry collector, such as the CSRF token.
func stripTelemetryRequest(req *http.Request) {
	for name := range req.Header {
		if _, keep := keepHeadersInProxiedTelemetry[http.CanonicalHeaderKey(name)]; !keep {
			req.Header.Del(name)
		}
	}
}

// useBILogger indicates if the given siteID represents a deployment that uses a bi-logger
// service for on-prem telemetry logging
func useBILogger(siteID string) bool {
	return siteID == "Uber" || siteID == "UmamiWeb"
}
