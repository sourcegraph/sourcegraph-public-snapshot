package httpapi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"

	"github.com/gorilla/mux"
	log15 "gopkg.in/inconshreveable/log15.v2"

	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/app/envvar"
	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/pkg/siteid"
	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/telemetry"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/conf"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/env"
)

var telemetryHandler http.Handler

func init() {
	if !conf.GetTODO().DisableTelemetry && envvar.SourcegraphDotComMode() {
		// If on sourcegraph.com, no request body modification is necessary — just proxy the request
		telemetryHandler = &httputil.ReverseProxy{
			Director: func(req *http.Request) {
				req.URL.Scheme = "https"
				req.URL.Host = "sourcegraph-logging.telligentdata.com"
				req.Host = "sourcegraph-logging.telligentdata.com"
				req.URL.Path = "/" + mux.Vars(req)["TelemetryPath"]
			},
			ErrorLog: log.New(env.DebugOut, "telemetry proxy: ", log.LstdFlags),
		}
	} else {
		// Used to forward telemetry only when telemetry is enabled.
		proxy := &httputil.ReverseProxy{
			Director: rewriteTelemetryRequestToForward,
			ErrorLog: log.New(env.DebugOut, "telemetry proxy: ", log.LstdFlags),
		}

		telemetryHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Even if telemetry is disabled, we still want to collect samples, so we can show the
			// site admin what *would* be collected if it were enabled.
			stripTelemetryRequest(r)
			r.URL.Scheme, r.URL.Host = "https", "example.com" // needed for DumpRequestOut
			telemetry.Sample(r)

			// Handle when telemetry is disabled.
			if siteID := siteid.Get(); conf.GetTODO().DisableTelemetry || neverForwardTelemetry(siteID) {
				fmt.Fprintln(w, "telemetry is disabled")
				return
			}

			// Forward telemetry to Sourcegraph after rewriting via proxy.
			proxy.ServeHTTP(w, r)
		})
	}
}

func neverForwardTelemetry(siteID string) bool {
	return siteID == "Uber" || siteID == "UmamiWeb"
}

// rewriteTelemetryRequestToForward tranforms the telemetry payload by swapping out the appID for
// the Server instance's siteID so that the request can be forwarded to Sourcegraph.
//
// This allows the Editor and browser extensions to log events without caring what
// Server instance they're connected to. If the Server instance isn't sourcegraph.com,
// the Editor/Extension events will be correctly logged with an appID that matches
// the Server instance's siteID.
func rewriteTelemetryRequestToForward(r *http.Request) {
	// Decode the telemetry payload.
	var msg map[string]interface{}
	err := json.NewDecoder(r.Body).Decode(&msg)
	if err != nil {
		log15.Error("rewriteTelemetryRequestToForward: decoding JSON", "error", err)
		return
	}

	// Find the telemetry header map.
	header, ok := msg["header"].(map[string]interface{})
	if !ok {
		log15.Error("rewriteTelemetryRequestToForward: telemetry payload 'header' value must be a map")
		return
	}

	// Swap the appID for the siteID.
	header["app_id"] = siteid.Get()

	// Encode the new telemetry payload.
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(msg); err != nil {
		log15.Error("rewriteTelemetryRequestToForward: encoding JSON", "error", err)
		return
	}

	r.ContentLength = int64(buf.Len())
	r.Body = ioutil.NopCloser(&buf)

	r.URL.Scheme = "https"
	r.URL.Host = "sourcegraph-logging.telligentdata.com"
	r.Host = "sourcegraph-logging.telligentdata.com"
	r.URL.Path = "/" + mux.Vars(r)["TelemetryPath"]
}

var keepHeadersInProxiedTelemetry = map[string]struct{}{
	"Host":            {},
	"User-Agent":      {},
	"Content-Length":  {},
	"Accept":          {},
	"Accept-Encoding": {},
	"Accept-Language": {},
	"Cache-Control":   {},
	"Connection":      {},
	"Content-Type":    {},
	"Pragma":          {},
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
