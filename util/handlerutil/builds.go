package handlerutil

import (
	"net/http"

	"sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/auth"
	"src.sourcegraph.com/sourcegraph/util/buildutil"
	"src.sourcegraph.com/sourcegraph/util/httputil/httpctx"
)

// RepoBuildCommon holds all of the commit-specific information
// necessary to render a repository page template for a certain
// revision's best-match build. It is returned by
// getBestMatchBuild. It is assumed that pages rendered are also
// provided with repoCommon and repoRevCommon template data.
type RepoBuildCommon struct {
	BestRevSpec   sourcegraph.RepoRevSpec
	RepoBuildInfo *sourcegraph.RepoBuildInfo
	Built         bool
}

// GetRepoBuildCommonOpt values configure calls to getBestMatchBuild.
type GetRepoBuildCommonOpt struct {
	// allowUnbuilt causes getBestMatchBuild to NOT display the "repo
	// revision must be built" interstitial for revisions that lack a
	// successful build.
	AllowUnbuilt bool
}

// AllowBrowsingUnbuiltRepo is whether getBestMatchBuild should apply
// allowUnbuilt to the given repo. In general, customer repos should
// be able to be browsed even when unbuilt, to provide full repo
// browsing functionality.
//
// AllowBrowsingUnbuiltRepo is overridden by tests.
var AllowBrowsingUnbuiltRepo = func(repo *sourcegraph.Repo) bool {
	// Allow browsing unbuilt non-GitHub repos for customers.
	return !repo.IsGitHubRepo() || true // TODO(sqs!): remove the "true"
}

// GetRepoBuildCommon finds the most recent successful build (using
// ReposService.GetBuild) for the revision.  May return custom
// error NoBuildError to allow special handling of this case.
func GetRepoBuildCommon(r *http.Request, rc *RepoCommon, vc *RepoRevCommon, opts *GetRepoBuildCommonOpt) (bc RepoBuildCommon, err error) {
	if opts == nil {
		opts = new(GetRepoBuildCommonOpt)
	}
	apiclient := APIClient(r)

	bc = RepoBuildCommon{BestRevSpec: vc.RepoRevSpec}

	isAbsoluteCommitID := len(vc.RepoRevSpec.Rev) == 40
	isDefaultBranch := vc.RepoRevSpec.Rev == rc.Repo.DefaultBranch

	bc.RepoBuildInfo, err = apiclient.Builds.GetRepoBuildInfo(httpctx.FromRequest(r), &sourcegraph.BuildsGetRepoBuildInfoOp{
		Repo: vc.RepoRevSpec,
		Opt:  &sourcegraph.BuildsGetRepoBuildInfoOptions{Exact: isAbsoluteCommitID || !isDefaultBranch},
	})
	noBuild := err != nil && IsHTTPErrorCode(err, http.StatusNotFound)
	noSuccessfulBuild := bc.RepoBuildInfo != nil && bc.RepoBuildInfo.LastSuccessful == nil
	if noBuild || noSuccessfulBuild {
		err = nil // zero out so we don't return a stale, already-handled error later

		// Do we allow the user to browse unbuilt repo revs?
		allowUnbuilt := opts.AllowUnbuilt || AllowBrowsingUnbuiltRepo(rc.Repo)

		if !allowUnbuilt {
			// Trigger build if no build exists, and show "repo not built" interstitial.
			var (
				b          *sourcegraph.Build
				needsLogin bool
			)
			if bc.RepoBuildInfo != nil && bc.RepoBuildInfo.Exact != nil {
				// Existing but failed or not-yet-completed build.
				b = bc.RepoBuildInfo.Exact
			} else if a := auth.ActorFromContext(httpctx.FromRequest(r)); !a.IsAuthenticated() {
				needsLogin = true
			} else {
				// No build, so trigger a build (if user is logged in,
				// to avoid triggering builds for web crawlers).
				b, err = apiclient.Builds.Create(httpctx.FromRequest(r), &sourcegraph.BuildsCreateOp{
					RepoRev: vc.RepoRevSpec,
					Opt: &sourcegraph.BuildCreateOptions{
						BuildConfig: sourcegraph.BuildConfig{
							Import:   true,
							Queue:    true,
							Priority: int32(buildutil.DefaultPriority(rc.Repo.Private, buildutil.Manual)),
						},
					},
				})
				if err != nil {
					return
				}
			}
			return RepoBuildCommon{}, &NoBuildError{
				RepoCommon:    *rc,
				RepoRevCommon: *vc,
				NeedsLogin:    needsLogin,
				Build:         b,
			}

		}
	} else if err != nil {
		return
	}

	if bc.RepoBuildInfo != nil && bc.RepoBuildInfo.LastSuccessful != nil {
		bc.Built = true
		bc.BestRevSpec.CommitID = bc.RepoBuildInfo.LastSuccessful.CommitID
	}

	return
}
