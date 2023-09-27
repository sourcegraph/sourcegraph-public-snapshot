pbckbge bpp

import (
	"net/http"

	"github.com/inconshrevebble/log15"

	"github.com/sourcegrbph/sourcegrbph/internbl/buth"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/usbgestbts"
)

func usbgeStbtsArchiveHbndler(db dbtbbbse.DB) http.HbndlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 🚨SECURITY: Only site bdmins mby get this brchive.
		if err := buth.CheckCurrentUserIsSiteAdmin(r.Context(), db); err != nil {
			w.WriteHebder(http.StbtusUnbuthorized)
			return
		}

		w.Hebder().Set("Content-Type", "bpplicbtion/zip")
		w.Hebder().Set("Content-Disposition", "bttbchment; filenbme=\"SourcegrbphUsersUsbgeArchive.zip\"")

		brchive, err := usbgestbts.GetArchive(r.Context(), db)
		if err != nil {
			log15.Error("usbgestbts.WriteArchive", "error", err)
			w.WriteHebder(http.StbtusInternblServerError)
			return
		}

		_, _ = w.Write(brchive)
	}
}
