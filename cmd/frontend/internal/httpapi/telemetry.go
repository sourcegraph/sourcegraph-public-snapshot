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

	"github.com/sourcegraph/sourcegraph/cmd/frontend/envvar"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/pkg/usagestats"
	"github.com/sourcegraph/sourcegraph/pkg/env"
	"github.com/sourcegraph/sourcegraph/pkg/eventlogger"
)

var telemetryHandler http.Handler

func init() {
	if envvar.SourcegraphDotComMode() {
		telemetryHandler = &httputil.ReverseProxy{
			Director: func(req *http.Request) {
				req.URL.Scheme = "https"
				req.URL.Host = "sourcegraph-logging.telligentdata.com"
				req.Host = "sourcegraph-logging.telligentdata.com"
				req.URL.Path = "/" + mux.Vars(req)["TelemetryPath"]

				var tr eventlogger.TelemetryRequest
				err := json.NewDecoder(req.Body).Decode(&tr)
				if err != nil {
					log15.Error("telemetryHandler: Decode", "error", err)
				}

				var buf bytes.Buffer
				if err := json.NewEncoder(&buf).Encode(tr.Payload); err != nil {
					log15.Error("telemetryHandler: Encode", "error", err)
				}
				req.Body = ioutil.NopCloser(&buf)
			},
			ErrorLog: log.New(env.DebugOut, "telemetry proxy: ", log.LstdFlags),
		}
	} else {
		telemetryHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var tr eventlogger.TelemetryRequest
			err := json.NewDecoder(r.Body).Decode(&tr)
			if err != nil {
				log15.Error("telemetryHandler: Decode(2)", "error", err)
			}
			if tr.UserID != 0 && tr.EventLabel == "SavedSearchEmailNotificationSent" {
				err = usagestats.LogActivity(true, tr.UserID, "", "STAGEVERIFY")
				if err != nil {
					log15.Error("telemetryHandler: usagestats.LogActivity", "error", err)
				}
			}

			fmt.Fprintln(w, "event-level telemetry is disabled")
			w.WriteHeader(http.StatusNoContent)
		})
	}
}
