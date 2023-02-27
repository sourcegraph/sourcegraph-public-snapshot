package httpapi

import (
	"encoding/json"
	"net/http"

	"github.com/inconshreveable/log15"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/deviceid"
	"github.com/sourcegraph/sourcegraph/internal/eventlogger"
	"github.com/sourcegraph/sourcegraph/internal/featureflag"
	"github.com/sourcegraph/sourcegraph/internal/usagestats"
)

func telemetryHandler(db database.DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var tr eventlogger.TelemetryRequest
		err := json.NewDecoder(r.Body).Decode(&tr)
		if err != nil {
			log15.Error("telemetryHandler: Decode", "error", err)
		}
		err = usagestats.LogBackendEvent(db, tr.UserID, deviceid.FromContext(r.Context()), tr.EventName, tr.Argument, tr.PublicArgument, featureflag.GetEvaluatedFlagSet(r.Context()), nil)
		if err != nil {
			log15.Error("telemetryHandler: usagestats.LogBackendEvent", "error", err)
		}
		w.WriteHeader(http.StatusNoContent)
	})
}
