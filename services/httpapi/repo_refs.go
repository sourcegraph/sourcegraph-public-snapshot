package httpapi

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/handlerutil"
	"sourcegraph.com/sourcegraph/sourcegraph/services/backend"
)

func serveRepoRefs(w http.ResponseWriter, r *http.Request) error {
	repo, _, err := handlerutil.GetRepoAndRev(r.Context(), mux.Vars(r))
	if err != nil {
		return errors.Wrap(err, "GetRepoAndRev")
	}

	totalRefs, err := backend.Defs.TotalRefs(r.Context(), repo.URI)
	if err != nil {
		return errors.Wrap(err, "Defs.TotalRefs")
	}

	deprTotalRefs, deprErr := backend.Defs.DeprecatedTotalRefs(r.Context(), repo.URI)
	if deprErr != nil && totalRefs == 0 {
		return errors.Wrap(deprErr, "Defs.DeprecatedTotalRefs")
	}
	if deprTotalRefs > totalRefs {
		totalRefs = deprTotalRefs
	}

	return writeJSON(w, &struct {
		TotalRefs int
	}{
		TotalRefs: totalRefs,
	})
}
