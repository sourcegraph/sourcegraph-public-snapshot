package httpapi

import (
	"encoding/csv"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	"github.com/sourcegraph/sourcegraph/internal/auth"
	"github.com/sourcegraph/sourcegraph/internal/search/exhaustive/service"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func ServeSearchJobDownload(svc *service.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		jobIDStr := mux.Vars(r)["id"]
		jobID, err := strconv.Atoi(jobIDStr)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		_, err = svc.GetSearchJob(r.Context(), int64(jobID))
		if err != nil {
			if errors.Is(err, auth.ErrMustBeSiteAdminOrSameUser) {
				http.Error(w, err.Error(), http.StatusForbidden)
				return
			} else {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}

		w.Header().Set("Content-Type", "text/csv")
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%d.csv\"", jobID))

		// dummy data
		csvWriter := csv.NewWriter(w)
		_ = csvWriter.Write([]string{"repo", "revspec", "revision"})
		_ = csvWriter.Write([]string{"1", "spec", "2"})
		csvWriter.Flush()
	}
}
