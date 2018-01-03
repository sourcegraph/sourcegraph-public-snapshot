package httpapi

import (
	"fmt"
	"net/http"
	"net/http/httputil"

	"github.com/gorilla/mux"
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

				req.URL.Scheme = "https"
				req.URL.Host = "sourcegraph-logging.telligentdata.com"
				req.Host = "sourcegraph-logging.telligentdata.com"
				req.URL.Path = "/" + mux.Vars(req)["TelemetryPath"]
			},
		}
	}
}

// stripTelemetryRequest removes sensitive and unnecessary data from the client request
// before forwarding it up to the telemetry collector, such as the CSRF token.
func stripTelemetryRequest(req *http.Request) {
	req.Header.Del("cookie")
	req.Header.Del("origin")
	req.Header.Del("referer")
}
