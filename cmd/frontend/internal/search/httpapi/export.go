package httpapi

import (
	"context"
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
func filenamePrefix(jobID int64) string {
	return fmt.Sprintf("search-jobs_%d_%s", jobID, time.Now().Format("2006-01-02_150405"))
}

func ServeSearchJobDownload(logger log.Logger, svc *service.Service) http.HandlerFunc {
	logger = logger.With(log.String("handler", "ServeSearchJobDownload"))

	return func(w http.ResponseWriter, r *http.Request) {
		jobIDStr := mux.Vars(r)["id"]
		jobID, err := strconv.ParseInt(jobIDStr, 10, 64)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		writerTo, err := svc.GetSearchJobResultsWriterTo(r.Context(), int64(jobID))
		if err != nil {
			httpError(w, err)
			return
		}

		filename := filenamePrefix(jobID) + ".jsonl"
		writeJSON(logger.With(log.Int64("jobID", jobID)), w, filename, writerTo)
	}
}

func ServeSearchJobLogs(logger log.Logger, svc *service.Service) http.HandlerFunc {
	logger = logger.With(log.String("handler", "ServeSearchJobLogs"))

	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		jobIDStr := mux.Vars(r)["id"]
		jobID, err := strconv.ParseInt(jobIDStr, 10, 64)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		filename := filenamePrefix(jobID) + ".log.csv"

		// Jobs in a terminal state are aggregated. As part of the aggregation, the logs
		// are stored in the blobstore. If the job is not in a terminal state, the logs
		// are still in the database. It's possible that this call races with the
		// aggregation process, but the chances are slim and the user can always retry
		// to download the logs.
		job, err := svc.GetSearchJob(ctx, jobID)
		if err != nil {
			httpError(w, err)
			return
		}
		if job.IsAggregated {
			serveLogFromBlobstore(ctx, logger, svc, filename, jobID, w)
		} else {
			serveLogFromDB(ctx, logger, svc, filename, jobID, w)
		}
	}
}

func serveLogFromDB(ctx context.Context, logger log.Logger, svc *service.Service, filename string, jobID int64, w http.ResponseWriter) {
	csvWriterTo, err := svc.GetSearchJobLogsWriterTo(ctx, jobID)
	if err != nil {
		httpError(w, err)
		return
	}

	writeCSV(logger.With(log.Int64("jobID", jobID)), w, filename, csvWriterTo)
}

func serveLogFromBlobstore(ctx context.Context, logger log.Logger, svc *service.Service, filenameNoQuotes string, jobID int64, w http.ResponseWriter) {
	rc, err := svc.GetJobLogs(ctx, jobID)
	if err != nil {
		httpError(w, err)
		return
	}
	defer rc.Close()
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filenameNoQuotes))
	w.WriteHeader(200)
	n, err := io.Copy(w, rc)
	if err != nil {
		logger.Warn("failed while writing search job csv response", log.String("filename", filenameNoQuotes), log.Int64("bytesWritten", n), log.Error(err))
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

func writeJSON(logger log.Logger, w http.ResponseWriter, filenameNoQuotes string, writerTo io.WriterTo) {
	w.Header().Set("Content-Type", "application/jsonlines")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filenameNoQuotes))
	w.WriteHeader(200)
	n, err := writerTo.WriteTo(w)
	if err != nil {
		logger.Warn("failed while writing search job response", log.String("filename", filenameNoQuotes), log.Int64("bytesWritten", n), log.Error(err))
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
