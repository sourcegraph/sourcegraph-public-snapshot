package app

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/sourcegraph/log"

	oce "github.com/sourcegraph/sourcegraph/cmd/frontend/oneclickexport"
	"github.com/sourcegraph/sourcegraph/internal/auth"
	"github.com/sourcegraph/sourcegraph/internal/database"
)

func oneClickExportHandler(db database.DB, logger log.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 🚨SECURITY: Only site admins may get this archive.
		ctx := r.Context()
		if err := auth.CheckCurrentUserIsSiteAdmin(ctx, db); err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		w.Header().Set("Content-Type", "application/zip")
		w.Header().Set("Content-Disposition", "attachment; filename=\"SourcegraphDataExport.zip\"")

		var request oce.ExportRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		archive, err := oce.GlobalExporter.Export(ctx, request)
		if err != nil {
			logger.Error("OneClickExport", log.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		_, err = io.Copy(w, archive)
		if err != nil {
			logger.Error("Writing archive to HTTP response", log.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}
}
