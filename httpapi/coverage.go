package httpapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/gorilla/mux"

	"sourcegraph.com/sourcegraph/srclib/cvg"

	"sourcegraph.com/sourcegraph/sourcegraph/go-sourcegraph/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/util/errcode"
	"sourcegraph.com/sourcegraph/sourcegraph/util/handlerutil"
)

// Thresholds under which we should fail the build
const (
	FileScoreThresh = 0.2
	RefScoreThresh  = 0.5
)

var errCoverageIsBad = errcode.HTTPErr{Status: http.StatusNotAcceptable, Err: fmt.Errorf("coverage did not meet minimum thresholds")}

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

	var cov map[string]*cvg.Coverage
	if err := json.NewDecoder(r.Body).Decode(&cov); err != nil {
		return err
	}

	foundGoodLang := false
	for _, langcov := range cov {
		if langcov.FileScore >= FileScoreThresh && langcov.RefScore >= RefScoreThresh {
			foundGoodLang = true
			break
		}
	}
	if !foundGoodLang {
		return &errCoverageIsBad
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
