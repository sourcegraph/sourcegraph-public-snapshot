package httpapi

import (
	"net/http"
	"net/http/httputil"

	"github.com/gorilla/mux"
)

var telemetryReverseProxy = &httputil.ReverseProxy{
	Director: func(req *http.Request) {
		req.URL.Scheme = "https"
		req.URL.Host = "sourcegraph-logging.telligentdata.com"
		req.Host = "sourcegraph-logging.telligentdata.com"
		req.URL.Path = "/" + mux.Vars(req)["TelemetryPath"]
	},
}
