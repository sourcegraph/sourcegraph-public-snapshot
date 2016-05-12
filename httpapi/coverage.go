package httpapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/gorilla/mux"

	"sourcegraph.com/sourcegraph/sourcegraph/services/ext/slack"
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

// serveCoverage imports coverage data for a given repository and
// revision. If it detects a regression in coverage, it triggers an
// alert.
func serveCoverage(w http.ResponseWriter, r *http.Request) error {
	if strings.ToLower(r.Header.Get("content-type")) != "application/json" {
		w.WriteHeader(http.StatusBadRequest)
		return errors.New("requires Content-Type: application/json")
	}
	ctx, cl := handlerutil.Client(r)

	_, repoRev, err := handlerutil.GetRepoAndRev(ctx, mux.Vars(r))
	if err != nil {
		return err
	}

	var cov map[string]*cvg.Coverage
	if err := json.NewDecoder(r.Body).Decode(&cov); err != nil {
		return err
	}

	// If previous coverage exists for HEAD commit, alert on failure.
	headCommit, err := cl.Repos.GetCommit(ctx, &sourcegraph.RepoRevSpec{RepoSpec: sourcegraph.RepoSpec{URI: repoRev.URI}})
	if err != nil {
		return err
	}
	if string(headCommit.ID) == repoRev.CommitID {
		if prevCov, dataVer, err := handlerutil.GetCoverage(cl, ctx, repoRev.URI); err == nil {
			if cvg.HasRegressed(prevCov, cov) {
				slack.PostMessage(slack.PostOpts{
					Msg: fmt.Sprintf(`Coverage for %s has regressed.
Bfore, commit %s had %s.
After, commit %s has %s.`, repoRev.URI, dataVer.CommitID, summary(prevCov), repoRev.CommitID, summary(cov)),
					IconEmoji: ":warning:",
					Channel:   "global-graph",
				})
			}
		}
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

func summary(c map[string]*cvg.Coverage) string {
	s := make(map[string]*cvg.Coverage)

	for lang, cov := range c {
		covCopy := *cov
		covCopy.UncoveredFiles = nil
		s[lang] = &covCopy
	}

	b, _ := json.MarshalIndent(s, "", "  ")
	return string(b)
}
