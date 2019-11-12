package httpapi

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httputil"

	log15 "gopkg.in/inconshreveable/log15.v2"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/envvar"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/pkg/usagestats2"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/eventlogger"
)

var telemetryHandler http.Handler

func init() {
	if envvar.SourcegraphDotComMode() {
		telemetryHandler = &httputil.ReverseProxy{
			Director: func(req *http.Request) {
				// Removed due to our event logging ETL pipeline sunsetting schedule.
				// TODO(Dan): update with new logging URL.
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
			err = usagestats2.LogEvent(context.Background(), tr.EventName, "", tr.UserID, "backend", "BACKEND", nil)
			if err != nil {
				log15.Error("telemetryHandler: usagestats2.LogEvent", "error", err)
			}
			w.WriteHeader(http.StatusNoContent)
		})
	}
}
