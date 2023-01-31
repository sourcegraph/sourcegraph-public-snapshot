package app

import (
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/archive"
	"github.com/sourcegraph/sourcegraph/internal/auth"
	"github.com/sourcegraph/sourcegraph/internal/database"
)

const (
	MAX_SIZE = 10 << 20 // 10MB
	FORM_KEY = "archive"
)

func archiveDownloadHandler(logger log.Logger, db database.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 🚨SECURITY: Only site admins may get this archive.
		if err := auth.CheckCurrentUserIsSiteAdmin(r.Context(), db); err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		w.Header().Set("Content-Type", "application/zip")
		w.Header().Set("Content-Disposition", "attachment; filename=\"SourcegraphArchive.zip\"")

		archive, err := archive.CreateArchive(r.Context(), logger)
		if err != nil {
			logger.Error("archive.download", log.Error(err))
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

func archiveUploadHandler(logger log.Logger, db database.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// TODO(jac): also allow if the site hasn't been initialized yet
		// 🚨SECURITY: Only site admins can restore
		if err := auth.CheckCurrentUserIsSiteAdmin(r.Context(), db); err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		r.ParseMultipartForm(MAX_SIZE)
		file, handler, err := r.FormFile(FORM_KEY)
		if err != nil {
			logger.Error("archive.upload", log.Error(err))
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		defer file.Close()

		dir, err := os.MkdirTemp(os.TempDir(), "archive-*")
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		defer os.RemoveAll(dir)

		f, err := os.Create(filepath.Join(dir, handler.Filename))
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// Copy uploaded file from memory to file on disk
		if written, err := io.Copy(f, file); err != nil || written != handler.Size {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		f.Close()

		err = archive.RestoreFromArchive(r.Context(), logger, f.Name())
		if err != nil {
			//TODO(jac):
			// - type of errors e.g. bad archive (client), myriad errors (server)
			// - return the error to the user
			w.WriteHeader(http.StatusInternalServerError)
			logger.Error("Restore Failed", log.Error(err))
			return
		}
	}
}
