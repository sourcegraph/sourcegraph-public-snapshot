package app

import (
	"encoding/json"
	"net/http"

	"github.com/sourcegraph/log"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	oce "github.com/sourcegraph/sourcegraph/cmd/frontend/oneclickexport"
	"github.com/sourcegraph/sourcegraph/internal/database"
)

func oneClickExportHandler(db database.DB, logger log.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 🚨SECURITY: Only site admins may get this archive.
		ctx := r.Context()
		if err := backend.CheckCurrentUserIsSiteAdmin(ctx, db); err != nil {
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

		// TODO: change when Exporter.Export is refactored to return io.Reader
		for len(archive) > 0 {
			bytesWritten, err := w.Write(archive)
			if err != nil {
				logger.Error("OneClickExport output write", log.Error(err))
				return
			}

			// all bytes written, exiting the function
			if bytesWritten == len(archive) {
				break
			}

			// writing remaining bytes
			archive = archive[bytesWritten:]
		}
	}
}
