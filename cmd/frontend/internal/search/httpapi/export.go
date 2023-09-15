package httpapi

import (
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

		w.Header().Set("Content-Type", "text/csv")
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%d.csv\"", jobID))

		err = svc.WriteSearchJobCSV(r.Context(), w, int64(jobID))
		if err != nil {
			if errors.Is(err, auth.ErrMustBeSiteAdminOrSameUser) {
				http.Error(w, err.Error(), http.StatusForbidden)
				return
			}
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

	}
}
