package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/sourcegraph/mux"

	"sourcegraph.com/sourcegraph/srclib/cvg"

	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/util/handlerutil"
)

func serveCoverage(w http.ResponseWriter, r *http.Request) error {
	if strings.ToLower(r.Header.Get("content-type")) != "application/json" {
		w.WriteHeader(http.StatusBadRequest)
		return errors.New("requires Content-Type: application/json")
	}

	ctx, cl := handlerutil.Client(r)

	_, repoRev, _, err := handlerutil.GetRepoAndRev(ctx, mux.Vars(r))
	if err != nil {
		return err
	}

	var cov cvg.Coverage
	if err := json.NewDecoder(r.Body).Decode(&cov); err != nil {
		return err
	}

	covJSON, err := json.Marshal(cov)
	if err != nil {
		return err
	}

	var statusUpdate sourcegraph.RepoStatusesCreateOp
	statusUpdate.Repo = repoRev
	statusUpdate.Status = sourcegraph.RepoStatus{
		Context:     "coverage",
		Description: string(covJSON),
	}
	if _, err = cl.RepoStatuses.Create(ctx, &statusUpdate); err != nil {
		return err
	}

	return nil
}
