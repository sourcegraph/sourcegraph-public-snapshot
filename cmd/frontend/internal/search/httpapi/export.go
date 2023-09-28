package httpapi

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"

	"github.com/sourcegraph/sourcegraph/internal/auth"
	"github.com/sourcegraph/sourcegraph/internal/search/exhaustive/service"
	"github.com/sourcegraph/sourcegraph/internal/search/exhaustive/store"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// search-jobs_<job-id>_2020-07-01_150405
func filename(jobID int) string {
	return fmt.Sprintf("search-jobs_%d_%s", jobID, time.Now().Format("2006-01-02_150405"))
}

func ServeSearchJobDownload(svc *service.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		jobIDStr := mux.Vars(r)["id"]
		jobID, err := strconv.Atoi(jobIDStr)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "text/csv")
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s.csv\"", filename(jobID)))

		n, err := svc.WriteSearchJobCSV(r.Context(), w, int64(jobID))
		if err != nil {
			httpError(w, err)
			return
		}
		if n == 0 {
			w.WriteHeader(http.StatusNoContent)
		}
	}
}

func ServeSearchJobLogs(svc *service.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		jobIDStr := mux.Vars(r)["id"]
		jobID, err := strconv.Atoi(jobIDStr)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "text/csv")
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s.log\"", filename(jobID)))

		err = svc.WriteSearchJobLogs(r.Context(), w, int64(jobID))
		if err != nil {
			httpError(w, err)
		}

	}
}

func httpError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, auth.ErrMustBeSiteAdminOrSameUser):
		http.Error(w, err.Error(), http.StatusForbidden)
	case errors.Is(err, store.ErrNoResults), errors.Is(err, sql.ErrNoRows):
		http.Error(w, err.Error(), http.StatusNotFound)
	default:
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
