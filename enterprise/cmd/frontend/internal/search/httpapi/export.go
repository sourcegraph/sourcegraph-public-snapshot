package httpapi

import (
	"encoding/csv"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/graph-gophers/graphql-go"

	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/search/resolvers"
	"github.com/sourcegraph/sourcegraph/internal/auth"
	"github.com/sourcegraph/sourcegraph/internal/search/exhaustive/store"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func ServeSearchJobDownload(store *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		nodeID := mux.Vars(r)["id"]

		jobID, err := resolvers.UnmarshalSearchJobID(graphql.ID(nodeID))
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// 🚨 SECURITY: only the initiator, internal or site admins may view a job
		_, err = store.GetExhaustiveSearchJob(r.Context(), jobID)
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
