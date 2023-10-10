package httpapi

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/auth"
	"github.com/sourcegraph/sourcegraph/internal/search/exhaustive/service"
	"github.com/sourcegraph/sourcegraph/internal/search/exhaustive/store"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// search-jobs_<job-id>_2020-07-01_150405
func filenamePrefix(jobID int) string {
	return fmt.Sprintf("search-jobs_%d_%s", jobID, time.Now().Format("2006-01-02_150405"))
}

func ServeSearchJobDownload(logger log.Logger, svc *service.Service) http.HandlerFunc {
	logger = logger.With(log.String("handler", "ServeSearchJobDownload"))

	return func(w http.ResponseWriter, r *http.Request) {
		jobIDStr := mux.Vars(r)["id"]
		jobID, err := strconv.Atoi(jobIDStr)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		csvWriterTo, err := svc.GetSearchJobCSVWriterTo(r.Context(), int64(jobID))
		if err != nil {
			httpError(w, err)
			return
		}

		filename := filenamePrefix(jobID) + ".csv"
		writeCSV(logger.With(log.Int("jobID", jobID)), w, filename, csvWriterTo)
	}
}

func ServeSearchJobLogs(logger log.Logger, svc *service.Service) http.HandlerFunc {
	logger = logger.With(log.String("handler", "ServeSearchJobLogs"))

	return func(w http.ResponseWriter, r *http.Request) {
		jobIDStr := mux.Vars(r)["id"]
		jobID, err := strconv.Atoi(jobIDStr)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		csvWriterTo, err := svc.GetSearchJobLogsWriterTo(r.Context(), int64(jobID))
		if err != nil {
			httpError(w, err)
			return
		}

		filename := filenamePrefix(jobID) + ".log.csv"
		writeCSV(logger.With(log.Int("jobID", jobID)), w, filename, csvWriterTo)
	}
}

func writeCSV(logger log.Logger, w http.ResponseWriter, filenameNoQuotes string, writerTo io.WriterTo) {
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filenameNoQuotes))
	w.WriteHeader(200)
	n, err := writerTo.WriteTo(w)
	if err != nil {
		logger.Warn("failed while writing search job csv response", log.String("filename", filenameNoQuotes), log.Int64("bytesWritten", n), log.Error(err))
	}
}

func httpError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, auth.ErrMustBeSiteAdminOrSameUser):
		http.Error(w, err.Error(), http.StatusForbidden)
	case errors.Is(err, store.ErrNoResults):
		http.Error(w, err.Error(), http.StatusNotFound)
	default:
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
