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

var telemetryReverseProxy http.Handler

func init() {
	// If telemetry is disabled, we still want to collect samples, so we can show the
	// site admin what *would* be collected if it were enabled.
	if conf.GetTODO().DisableTelemetry {
		telemetryReverseProxy = http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			stripTelemetryRequest(req)
			req.URL.Scheme, req.URL.Host = "https", "example.com" // needed for DumpRequestOut
			telemetry.Sample(req)
			fmt.Fprintln(w, "telemetry is disabled")
		})
	} else {
		// If on sourcegraph.com, no request body modification is necessary — just proxy the request
		if envvar.SourcegraphDotComMode() {
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
				ErrorLog: log.New(env.DebugOut, "telemetry proxy: ", log.LstdFlags),
			}
		} else {
			telemetryReverseProxy = &httputil.ReverseProxy{
				Director: serveOnPremTelemetryModification,
				ErrorLog: log.New(env.DebugOut, "telemetry proxy: ", log.LstdFlags),
			}
		}
	}
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

// useBILogger indicates if the given siteID represents a deployment that uses a bi-logger
// service for on-prem telemetry logging
func useBILogger(siteID string) bool {
	return siteID == "Uber" || siteID == "UmamiWeb"
}

// serveOnPremTelemetryModification tranforms the telemetry payload by swapping out
// the appID for the Server instance's siteID.
//
// This allows the Editor and browser extensions to log events without caring what
// Server instance they're connected to. If the Server instance isn't sourcegraph.com,
// the Editor/Extension events will be correctly logged with an appID that matches
// the Server instance's siteID.
func serveOnPremTelemetryModification(r *http.Request) {
	// Decode the telemetry payload.
	var msg map[string]interface{}
	err := json.NewDecoder(r.Body).Decode(&msg)
	if err != nil {
		log15.Error("serveOnPremTelemetryModification: decoding JSON", "error", err)
		return
	}

	// Find the telemetry header map.
	header, ok := msg["header"].(map[string]interface{})
	if !ok {
		log15.Error("serveOnPremTelemetryModification: telemetry payload 'header' value must be a map")
		return
	}

	// Swap the appID for the siteID.
	header["app_id"] = siteid.Get()

	// Encode the new telemtry payload.
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(msg); err != nil {
		log15.Error("serveOnPremTelemetryModification: encoding JSON", "error", err)
		return
	}

	r.ContentLength = int64(buf.Len())
	r.Body = ioutil.NopCloser(&buf)

	stripTelemetryRequest(r)
	r.URL.Scheme, r.URL.Host = "https", "example.com" // needed for DumpRequestOut
	telemetry.Sample(r)

	// Point the request to the ultimate telemetry endpoint.
	if useBILogger(siteid.Get()) {
		r.URL.Scheme = "http"
		r.URL.Host = r.Host
		r.URL.Path = "/.bi-logger"
	} else {
		r.URL.Scheme = "https"
		r.URL.Host = "sourcegraph-logging.telligentdata.com"
		r.Host = "sourcegraph-logging.telligentdata.com"
		r.URL.Path = "/" + mux.Vars(r)["TelemetryPath"]
	}
}
