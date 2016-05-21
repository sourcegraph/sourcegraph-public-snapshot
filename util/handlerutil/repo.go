package handlerutil

import (
	"bytes"
	"fmt"
	"go/doc"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"gopkg.in/inconshreveable/log15.v2"
	"sourcegraph.com/sourcegraph/sourcegraph/go-sourcegraph/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/routevar"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/vcs"
	"sourcegraph.com/sourcegraph/sourcegraph/util/errcode"
	"sourcegraph.com/sourcegraph/sourcegraph/util/htmlutil"
	"sourcegraph.com/sourcegraph/sourcegraph/util/router_util"
	"sourcegraph.com/sourcegraph/srclib/graph"
)

// RepoCommon holds all of the information necessary to render a
// repository page template. It is returned by GetRepoFromRequest. See also
// RepoRevCommon.
type RepoCommon struct {
	Repo       *sourcegraph.Repo
	RepoConfig *sourcegraph.RepoConfig
}

// RepoRevCommon holds all of the commit-specific information
// necessary to render a repository page template for a certain
// commit. It is returned by GetRepoAndRevFromRequest. It is assumed that pages
// rendered are also provided with repoCommon template data.
type RepoRevCommon struct {
	RepoRevSpec sourcegraph.RepoRevSpec
}

// GetRepoAndRevCommon returns the repository and RepoRevSpec based on
// the route vars. It may also return custom error types
// URLMovedError, NoVCSDataError, which callers should ideally check
// for.
func GetRepoAndRevCommon(ctx context.Context, vars map[string]string) (rc *RepoCommon, vc *RepoRevCommon, err error) {
	rc, err = GetRepoCommon(ctx, vars)
	if err != nil {
		return
	}

	vc = &RepoRevCommon{}
	vc.RepoRevSpec.RepoSpec = rc.Repo.RepoSpec()

	vc.RepoRevSpec, err = getRepoRev(ctx, vars, rc.Repo.DefaultBranch)
	if err != nil {
		cloneInProgress := grpc.Code(err) == codes.Unavailable && grpc.ErrorDesc(err) == vcs.RepoNotExistError{CloneInProgress: true}.Error()
		if noVCSData := grpc.Code(err) == codes.NotFound ||
			cloneInProgress ||
			strings.Contains(err.Error(), "has no default branch"); noVCSData {

			if cloneInProgress {
				return rc, vc, err
			} else if rev := vars["Rev"]; rev != "" && rev != "@" {
				err = vcs.ErrRevisionNotFound
			} else {
				err = &NoVCSDataError{RepoCommon: rc}
			}
		}
		return
	}

	return
}

// GetRepoCommon returns the repository and RepoSpec based on the
// route vars. Callers should ideally handle the custom error type
// URLMovedError.
func GetRepoCommon(ctx context.Context, vars map[string]string) (rc *RepoCommon, err error) {
	cl, err := sourcegraph.NewClientFromContext(ctx)
	if err != nil {
		return nil, err
	}

	rc = &RepoCommon{}
	rc.Repo, _, err = GetRepo(ctx, vars)
	if err != nil {
		return
	}

	repoSpec := rc.Repo.RepoSpec()
	rc.RepoConfig, err = cl.Repos.GetConfig(ctx, &repoSpec)
	return
}

// GetRepo gets the repo (from the reposSvc) specified in the URL's
// RepoSpec route param. Callers should ideally check for a return error of type
// URLMovedError and handle this scenario by warning or redirecting the user.
func GetRepo(ctx context.Context, vars map[string]string) (repo *sourcegraph.Repo, repoSpec sourcegraph.RepoSpec, err error) {
	origRepoSpec, err := routevar.ToRepoSpec(vars)
	if err != nil {
		return nil, sourcegraph.RepoSpec{}, err
	}

	cl, err := sourcegraph.NewClientFromContext(ctx)
	if err != nil {
		return nil, sourcegraph.RepoSpec{}, err
	}

	repoSpec = origRepoSpec
	repo, err = cl.Repos.Get(ctx, &repoSpec)
	if err != nil {
		return nil, origRepoSpec, err
	}
	repoSpec = repo.RepoSpec()

	// Check for redirect.
	if origRepoSpec.URI != "" && origRepoSpec.URI != repoSpec.URI {
		return nil, repoSpec, &URLMovedError{repoSpec.URI}
	}

	return repo, repoSpec, nil
}

// getRepoRev resolves the RepoRevSpec and commit specified in the
// route vars. The provided defaultBranch is used if no rev is
// specified in the URL.
func getRepoRev(ctx context.Context, vars map[string]string, defaultRev string) (sourcegraph.RepoRevSpec, error) {
	repoRev, err := routevar.ToRepoRevSpec(vars)
	if err != nil {
		return sourcegraph.RepoRevSpec{}, err
	}

	cl, err := sourcegraph.NewClientFromContext(ctx)
	if err != nil {
		return sourcegraph.RepoRevSpec{}, err
	}

	commit, err := cl.Repos.GetCommit(ctx, &repoRev)
	if err != nil {
		return repoRev, err
	}
	repoRev.CommitID = string(commit.ID)

	if repoRev.Rev == "" {
		repoRev.Rev = defaultRev

		if defaultRev == "" {
			log15.Warn("GetRepoRev: no rev specified and repo has no default branch", "repo", repoRev.URI)
		}
	}

	if repoRev.Rev == "" {
		panic("empty Rev on repo " + repoRev.URI)
	}
	if repoRev.CommitID == "" {
		panic("empty CommitID on repo " + repoRev.URI + " rev " + repoRev.Rev)
	}

	return repoRev, nil
}

// GetRepoAndRev returns the Repo and the RepoRevSpec for a repository. It may
// also return custom error URLMovedError to allow special handling of this case,
// such as for example redirecting the user.
func GetRepoAndRev(ctx context.Context, vars map[string]string) (repo *sourcegraph.Repo, repoRevSpec sourcegraph.RepoRevSpec, err error) {
	repo, repoRevSpec.RepoSpec, err = GetRepo(ctx, vars)
	if err != nil {
		return repo, repoRevSpec, err
	}

	repoRevSpec, err = getRepoRev(ctx, vars, repo.DefaultBranch)
	return repo, repoRevSpec, err
}

// RedirectToNewRepoURI writes an HTTP redirect response with a
// Location that matches the request's location except with the
// RepoSpec route var updated to refer to newRepoURI (instead of the
// originally requested repo URI).
func RedirectToNewRepoURI(w http.ResponseWriter, r *http.Request, newRepoURI string) error {
	origVars := mux.Vars(r)
	origVars["Repo"] = routevar.RepoSpecString(sourcegraph.RepoSpec{URI: newRepoURI})

	destURL, err := mux.CurrentRoute(r).URLPath(router_util.MapToArray(origVars)...)
	if err != nil {
		return err
	}

	http.Redirect(w, r, destURL.String(), http.StatusMovedPermanently)
	return nil
}

// ResolveSrclibDataVersion calls Repos.GetSrclibDataVersionForPath on
// the given entry spec. If a srclib data version exists,
// entry.RepoRev.CommitID is set to the version's commit ID.
func ResolveSrclibDataVersion(ctx context.Context, entry sourcegraph.TreeEntrySpec) (sourcegraph.RepoRevSpec, *sourcegraph.SrclibDataVersion, error) {
	cl, err := sourcegraph.NewClientFromContext(ctx)
	if err != nil {
		return sourcegraph.RepoRevSpec{}, nil, err
	}

	dataVer, err := cl.Repos.GetSrclibDataVersionForPath(ctx, &entry)
	if err == nil {
		entry.RepoRev.CommitID = dataVer.CommitID
	}
	return entry.RepoRev, dataVer, err
}

// GetDefCommon returns common information about a definition, based
// on the route vars.  It additionally returns common repository and
// revision information. It may also return custom errors
// URLMovedError, or NoVCSDataError.
//
// dc.Def.DefKey will be set to the def specification based on the
// request when getting actual def fails.
func GetDefCommon(ctx context.Context, vars map[string]string, opt *sourcegraph.DefGetOptions) (dc *sourcegraph.Def, err error) {
	repoRev, err := routevar.ToRepoRevSpec(vars)
	if err != nil {
		return nil, err
	}

	defSpec, err := routevar.ToDefSpec(vars)
	if err != nil {
		return dc, err
	}

	// If we fail to get a def, return the best known information to the caller.
	dc = &sourcegraph.Def{
		Def: graph.Def{
			DefKey: graph.DefKey{
				Repo:     defSpec.Repo,
				Unit:     defSpec.Unit,
				UnitType: defSpec.UnitType,
				Path:     defSpec.Path,
			},
		},
	}

	cl, err := sourcegraph.NewClientFromContext(ctx)
	if err != nil {
		return
	}

	resolvedRev, _, err := ResolveSrclibDataVersion(ctx, sourcegraph.TreeEntrySpec{RepoRev: repoRev})
	if err != nil {
		return dc, err
	}
	defSpec.CommitID = resolvedRev.CommitID

	dc, err = cl.Defs.Get(ctx, &sourcegraph.DefsGetOp{Def: defSpec, Opt: opt})
	if err != nil {
		return dc, err
	}

	// Now that we have the def, we can check if its file has been
	// changed AFTER the resolved srclib-last-version. If so, then we
	// can't actually display this def, because we'd only be able to
	// show it on an older version of the file (which would mean that
	// users would see file data from an older commit when looking at
	// a newer commit's def--that is BAD).
	//
	// Right now, the best course of action is to 404. This is a
	// fairly rare case that should be remedied as soon as the next
	// build completes. The alternative would be to display a warning
	// saying "this file is N commits behind the requested commit,"
	// but that adds a lot of complexity to the code and to the UI (as
	// we have seen in the past). If a user's looking at a file that
	// was changed since the last srclib version, they also see an
	// unannotated file, so this is consistent with that behavior as
	// well.
	defResolvedRev, err := cl.Repos.GetSrclibDataVersionForPath(ctx, &sourcegraph.TreeEntrySpec{
		RepoRev: repoRev, // use originally requested rev, not already resolved last-srclib-version
		Path:    dc.File,
	})
	if err != nil {
		return dc, err
	}
	if defResolvedRev.CommitID != resolvedRev.CommitID {
		return dc, &errcode.HTTPErr{
			Status: http.StatusNotFound,
			Err:    fmt.Errorf("no srclib data for def %v (file %s was modified between last srclib analysis version %s and rev %s)", defSpec, dc.File, resolvedRev.CommitID, repoRev.Rev),
		}
	}

	// this can not be moved to svc/local, because HTML sanitation needs to
	// happen on the local sourcegraph instance, not on an untrusted
	// server
	if len(dc.Docs) > 0 {
		defDoc := dc.Docs[0]
		var docHTML string
		switch defDoc.Format {
		case "text/html":
			docHTML = defDoc.Data
		// TODO "text/x-markdown"
		// TODO "text/x-rst"
		default: // including "text/plain"
			var buf bytes.Buffer
			doc.ToHTML(&buf, defDoc.Data, nil)
			docHTML = buf.String()
		}
		dc.DocHTML = htmlutil.SanitizeForPB(docHTML)
	}
	return dc, nil
}
