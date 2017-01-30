package httpapi

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/handlerutil"
	"sourcegraph.com/sourcegraph/sourcegraph/services/backend"
)

func serveRepoShield(w http.ResponseWriter, r *http.Request) error {
	repo, _, err := handlerutil.GetRepoAndRev(r.Context(), mux.Vars(r))
	if err != nil {
		return errors.Wrap(err, "GetRepoAndRev")
	}

	totalRefs, err := backend.Defs.TotalRefs(r.Context(), repo.URI)
	if err != nil {
		return errors.Wrap(err, "Defs.TotalRefs")
	}

	w.Header().Set("X-Sourcegraph-Exact-Count", fmt.Sprint(totalRefs))

	// Format e.g. "1,399" as "1.3k".
	desc := fmt.Sprintf("%d projects", totalRefs)
	if totalRefs > 1000 {
		desc = fmt.Sprintf("%.1fk projects", float64(totalRefs)/1000.0)
	}

	return writeJSON(w, &struct {
		// Note: Named lowercase because the JSON is consumed by shields.io JS
		// code.
		Value string `json:"value"`
	}{
		// Note: We're adding a prefixed space because otherwise the shields.io
		// badge will be formatted badly (looks like `used by |12k projects`
		// instead of `used by | 12k projects`).
		Value: " " + desc,
	})
}
