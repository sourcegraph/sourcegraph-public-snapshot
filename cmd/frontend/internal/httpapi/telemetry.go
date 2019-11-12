package httpapi

import (
	"encoding/json"
	"net/http"

	log15 "gopkg.in/inconshreveable/log15.v2"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/pkg/usagestats2"
	"github.com/sourcegraph/sourcegraph/internal/eventlogger"
)

var telemetryHandler http.Handler

func init() {
	telemetryHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var tr eventlogger.TelemetryRequest
		err := json.NewDecoder(r.Body).Decode(&tr)
		if err != nil {
			log15.Error("telemetryHandler: Decode", "error", err)
		}
		err = usagestats2.LogBackendEvent(tr.UserID, tr.EventName, tr.Argument)
		if err != nil {
			log15.Error("telemetryHandler: usagestats2.LogBackendEvent", "error", err)
		}
		w.WriteHeader(http.StatusNoContent)
	})
}
