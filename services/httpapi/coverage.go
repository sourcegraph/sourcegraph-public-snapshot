package httpapi

import (
	"net/http"

	"sourcegraph.com/sourcegraph/sourcegraph/services/backend"
)

// serveCoverage returns coverage data for a given repository at the
// default branch revision.
func serveCoverage(w http.ResponseWriter, r *http.Request) error {
	list, err := backend.RepoStatuses.GetCoverage(r.Context())
	if err != nil {
		return err
	}

	return writeJSON(w, list)
}
