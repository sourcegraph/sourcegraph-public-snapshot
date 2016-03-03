package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"sourcegraph.com/sourcegraph/srclib/cli"

	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/util/handlerutil"
	"src.sourcegraph.com/sourcegraph/util/httputil/httpctx"
)

type coverage struct {
}

type repoCoverage struct {
}

func serveCoverage(w http.ResponseWriter, r *http.Request) error {
	if strings.ToLower(r.Header.Get("content-type")) != "application/json" {
		w.WriteHeader(http.StatusBadRequest)
		return errors.New("requires Content-Type: application/json")
	}

	ctx := httpctx.FromRequest(r)
	cl := handlerutil.APIClient(r)

	_, repoRev, _, err := handlerutil.GetRepoAndRev(r, cl.Repos)
	if err != nil {
		return err
	}

	var cov cli.Coverage
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
		Description: string(covJSON),
	}
	if _, err = cl.RepoStatuses.Create(ctx, &statusUpdate); err != nil {
		return err
	}

	return nil
}
