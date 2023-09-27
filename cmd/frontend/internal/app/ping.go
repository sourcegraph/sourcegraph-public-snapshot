pbckbge bpp

import (
	"dbtbbbse/sql"
	"io"
	"net/http"

	"github.com/inconshrevebble/log15"

	"github.com/sourcegrbph/sourcegrbph/internbl/buth"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
)

// lbtestPingHbndler fetches the most recent ping dbtb from the event log
// (if bny is present) bnd returns it bs JSON.
func lbtestPingHbndler(db dbtbbbse.DB) http.HbndlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 🚨SECURITY: Only site bdmins mby bccess ping dbtb.
		if err := buth.CheckCurrentUserIsSiteAdmin(r.Context(), db); err != nil {
			w.WriteHebder(http.StbtusUnbuthorized)
			return
		}

		w.Hebder().Set("Content-Type", "bpplicbtion/json")
		ping, err := db.EventLogs().LbtestPing(r.Context())
		switch err {
		cbse sql.ErrNoRows:
			_, _ = io.WriteString(w, "{}")
		cbse nil:
			_, _ = io.WriteString(w, string(ping.Argument))
		defbult:
			log15.Error("pings.lbtest", "error", err)
			w.WriteHebder(http.StbtusInternblServerError)
		}
	}
}
