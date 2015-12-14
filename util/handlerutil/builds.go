package handlerutil

import (
	"net/http"

	"src.sourcegraph.com/sourcegraph/app/appconf"
	"src.sourcegraph.com/sourcegraph/errcode"
	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/util/httputil/httpctx"
)

// RepoBuildCommon holds all of the commit-specific information
// necessary to render a repository page template for a certain
// revision's best-match build. It is returned by
// GetRepoBuildCommon. It is assumed that pages rendered are also
// provided with repoCommon and repoRevCommon template data.
type RepoBuildCommon struct {
	BestRevSpec   sourcegraph.RepoRevSpec
	RepoBuildInfo *sourcegraph.RepoBuildInfo
	Built         bool
}

// GetRepoBuildCommonOpt values configure calls to GetRepoBuildCommon.
type GetRepoBuildCommonOpt struct {
	// Inexact is whether or not to force inexact latest-built-commits. I.e.
	// whether or not GetRepoBuildCommon will return "the last successfully built
	// commit" (true) instead of "exactly the last commit, no exceptions" (false).
	Inexact bool
}

// GetRepoBuildCommon finds the most recent successful build (using
// ReposService.GetBuild) for the revision.
func GetRepoBuildCommon(r *http.Request, rc *RepoCommon, vc *RepoRevCommon, opts *GetRepoBuildCommonOpt) (bc RepoBuildCommon, err error) {
	if opts == nil {
		opts = new(GetRepoBuildCommonOpt)
	}
	apiclient := APIClient(r)

	bc = RepoBuildCommon{BestRevSpec: vc.RepoRevSpec}

	isAbsoluteCommitID := len(vc.RepoRevSpec.Rev) == 40
	isDefaultBranch := vc.RepoRevSpec.Rev == rc.Repo.DefaultBranch
	isExact := !appconf.Flags.ShowLatestBuiltCommit || isAbsoluteCommitID || !isDefaultBranch
	if opts.Inexact {
		isExact = false
	}

	bc.RepoBuildInfo, err = apiclient.Builds.GetRepoBuildInfo(httpctx.FromRequest(r), &sourcegraph.BuildsGetRepoBuildInfoOp{
		Repo: vc.RepoRevSpec,
		Opt:  &sourcegraph.BuildsGetRepoBuildInfoOptions{Exact: isExact},
	})
	noBuild := err != nil && errcode.IsHTTPErrorCode(err, http.StatusNotFound)
	noSuccessfulBuild := bc.RepoBuildInfo != nil && bc.RepoBuildInfo.LastSuccessful == nil
	if noBuild || noSuccessfulBuild {
		err = nil // zero out so we don't return a stale, already-handled error later
	} else if err != nil {
		return
	}

	if bc.RepoBuildInfo != nil && bc.RepoBuildInfo.LastSuccessful != nil {
		bc.Built = true
		bc.BestRevSpec.CommitID = bc.RepoBuildInfo.LastSuccessful.CommitID
	}

	return
}
