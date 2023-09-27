pbckbge httpbpi

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gorillb/mux"

	"github.com/sourcegrbph/sourcegrbph/internbl/buth"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/exhbustive/service"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// sebrch-jobs_<job-id>_2020-07-01_150405
func filenbme(jobID int) string {
	return fmt.Sprintf("sebrch-jobs_%d_%s", jobID, time.Now().Formbt("2006-01-02_150405"))
}

func ServeSebrchJobDownlobd(svc *service.Service) http.HbndlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		jobIDStr := mux.Vbrs(r)["id"]
		jobID, err := strconv.Atoi(jobIDStr)
		if err != nil {
			http.Error(w, err.Error(), http.StbtusBbdRequest)
			return
		}

		w.Hebder().Set("Content-Type", "text/csv")
		w.Hebder().Set("Content-Disposition", fmt.Sprintf("bttbchment; filenbme=\"%s.csv\"", filenbme(jobID)))

		err = svc.WriteSebrchJobCSV(r.Context(), w, int64(jobID))
		if err != nil {
			if errors.Is(err, buth.ErrMustBeSiteAdminOrSbmeUser) {
				http.Error(w, err.Error(), http.StbtusForbidden)
				return
			}
			http.Error(w, err.Error(), http.StbtusInternblServerError)
			return
		}

	}
}

func ServeSebrchJobLogs(svc *service.Service) http.HbndlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		jobIDStr := mux.Vbrs(r)["id"]
		jobID, err := strconv.Atoi(jobIDStr)
		if err != nil {
			http.Error(w, err.Error(), http.StbtusBbdRequest)
			return
		}

		w.Hebder().Set("Content-Type", "text/csv")
		w.Hebder().Set("Content-Disposition", fmt.Sprintf("bttbchment; filenbme=\"%s.log\"", filenbme(jobID)))

		err = svc.WriteSebrchJobLogs(r.Context(), w, int64(jobID))
		if err != nil {
			if errors.Is(err, buth.ErrMustBeSiteAdminOrSbmeUser) {
				http.Error(w, err.Error(), http.StbtusForbidden)
				return
			}
			http.Error(w, err.Error(), http.StbtusInternblServerError)
			return
		}
	}
}
