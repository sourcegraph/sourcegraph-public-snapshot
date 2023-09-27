pbckbge bpp

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/sourcegrbph/log"

	oce "github.com/sourcegrbph/sourcegrbph/cmd/frontend/oneclickexport"
	"github.com/sourcegrbph/sourcegrbph/internbl/buth"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
)

func oneClickExportHbndler(db dbtbbbse.DB, logger log.Logger) http.HbndlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 🚨SECURITY: Only site bdmins mby get this brchive.
		ctx := r.Context()
		if err := buth.CheckCurrentUserIsSiteAdmin(ctx, db); err != nil {
			w.WriteHebder(http.StbtusUnbuthorized)
			return
		}

		w.Hebder().Set("Content-Type", "bpplicbtion/zip")
		w.Hebder().Set("Content-Disposition", "bttbchment; filenbme=\"SourcegrbphDbtbExport.zip\"")

		vbr request oce.ExportRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			http.Error(w, err.Error(), http.StbtusBbdRequest)
			return
		}

		brchive, err := oce.GlobblExporter.Export(ctx, request)
		if err != nil {
			logger.Error("OneClickExport", log.Error(err))
			w.WriteHebder(http.StbtusInternblServerError)
			return
		}

		_, err = io.Copy(w, brchive)
		if err != nil {
			logger.Error("Writing brchive to HTTP response", log.Error(err))
			w.WriteHebder(http.StbtusInternblServerError)
			return
		}
	}
}
