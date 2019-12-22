package httpapi

import (
	"context"
	"encoding/json"
	"net/http"

	log15 "gopkg.in/inconshreveable/log15.v2"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/pkg/usagestats"
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
		err = usagestats.LogEvent(context.Background(), tr.EventName, "internal:backend", tr.UserID, "", "BACKEND", &tr.Argument)
		if err != nil {
			log15.Error("telemetryHandler: usagestats.LogBackendEvent", "error", err)
		}
		if tr.UserID != 0 && tr.EventName == "SavedSearchEmailNotificationSent" {
			err = usagestats.LogActivity(true, tr.UserID, "", "STAGEVERIFY")
			if err != nil {
				log15.Error("telemetryHandler: usagestats.LogBackendEvent", "error", err)
			}
		}
		w.WriteHeader(http.StatusNoContent)
	})
}
