package httpapi

import (
	"net/http"

	"sourcegraph.com/sourcegraph/sourcegraph/util/handlerutil"
	"sourcegraph.com/sqs/pbtypes"
)

// serveCoverage returns coverage data for a given repository at the
// default branch revision.
func serveCoverage(w http.ResponseWriter, r *http.Request) error {
	ctx, cl := handlerutil.Client(r)

	list, err := cl.RepoStatuses.GetCoverage(ctx, &pbtypes.Void{})
	if err != nil {
		return err
	}

	return writeJSON(w, list)
}
