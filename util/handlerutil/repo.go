package handlerutil

import (
	"bytes"
	"fmt"
	"go/doc"
	"html/template"
	"net/http"
	"strings"
	"time"

	"github.com/sourcegraph/mux"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"gopkg.in/inconshreveable/log15.v2"
	"sourcegraph.com/sourcegraph/srclib/graph"
	"src.sourcegraph.com/sourcegraph/app/router"
	"src.sourcegraph.com/sourcegraph/errcode"
	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/pkg/vcs"
	"src.sourcegraph.com/sourcegraph/sourcecode"
	"src.sourcegraph.com/sourcegraph/ui/payloads"
	"src.sourcegraph.com/sourcegraph/util/htmlutil"
	"src.sourcegraph.com/sourcegraph/util/router_util"
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
	RepoCommit  *payloads.AugmentedCommit
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

	cl, err := sourcegraph.NewClientFromContext(ctx)
	if err != nil {
		return
	}

	vc = &RepoRevCommon{}
	vc.RepoRevSpec.RepoSpec = rc.Repo.RepoSpec()

	var commit0 *vcs.Commit
	vc.RepoRevSpec, commit0, err = getRepoRev(ctx, vars, rc.Repo.DefaultBranch)
	if IsRepoNoVCSDataError(err) {
		if rc.Repo.Mirror {
			// Trigger cloning/updating this repo from its remote
			// mirror if it has one. Only wait 1 second. That's
			// usually enough to see if it failed immediately with an
			// error, but it lets us avoid blocking on the entire
			// clone process.
			ctx, cancel := context.WithTimeout(ctx, time.Second*1)
			defer cancel()
			if _, err = cl.MirrorRepos.RefreshVCS(ctx, &sourcegraph.MirrorReposRefreshVCSOp{Repo: vc.RepoRevSpec.RepoSpec}); err != nil {
				if ctx.Err() == context.DeadlineExceeded {
					// If deadline exceeded, fall through to NoVCSDataError return below.
				} else {
					return
				}
			}
		}
		if err != nil {
			if _, ok := vars["Rev"]; ok {
				return nil, nil, vcs.ErrRevisionNotFound
			}
			return nil, nil, &NoVCSDataError{rc}
		}
	} else if err != nil {
		return
	}

	if commit0 != nil {
		var augCommits []*payloads.AugmentedCommit
		augCommits, err = AugmentCommits(ctx, rc.Repo.URI, []*vcs.Commit{commit0})
		if err != nil {
			return
		}
		vc.RepoCommit = augCommits[0]
	}

	return
}

func IsRepoNoVCSDataError(err error) bool {
	return err != nil && (strings.Contains(err.Error(), "vcsstore") || errcode.IsHTTPErrorCode(err, http.StatusNotFound) ||
		strings.Contains(err.Error(), "has no default branch"))
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
	origRepoSpec, err := sourcegraph.UnmarshalRepoSpec(vars)
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
func getRepoRev(ctx context.Context, vars map[string]string, defaultRev string) (sourcegraph.RepoRevSpec, *vcs.Commit, error) {
	repoRev, err := sourcegraph.UnmarshalRepoRevSpec(vars)
	if err != nil {
		return sourcegraph.RepoRevSpec{}, nil, err
	}

	cl, err := sourcegraph.NewClientFromContext(ctx)
	if err != nil {
		return sourcegraph.RepoRevSpec{}, nil, err
	}

	commit, err := cl.Repos.GetCommit(ctx, &repoRev)
	if err != nil {
		return repoRev, nil, err
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

	return repoRev, commit, nil
}

// GetRepoAndRev returns the Repo and the RepoRevSpec for a repository. It may
// also return custom error URLMovedError to allow special handling of this case,
// such as for example redirecting the user.
func GetRepoAndRev(ctx context.Context, vars map[string]string) (repo *sourcegraph.Repo, repoRevSpec sourcegraph.RepoRevSpec, commit *vcs.Commit, err error) {
	repo, repoRevSpec.RepoSpec, err = GetRepo(ctx, vars)
	if err != nil {
		return repo, repoRevSpec, nil, err
	}

	repoRevSpec, commit, err = getRepoRev(ctx, vars, repo.DefaultBranch)
	return repo, repoRevSpec, commit, err
}

// ResolveRepoRev fills in the Rev and CommitID if they are missing.
func ResolveRepoRev(r *http.Request, repoRev *sourcegraph.RepoRevSpec) error {
	if repoRev.Rev != "" && len(repoRev.CommitID) == 40 {
		return nil
	}
	if repoRev.Rev == "" && len(repoRev.CommitID) == 40 {
		repoRev.Rev = repoRev.CommitID
		return nil
	}
	if len := len(repoRev.CommitID); len != 0 && len != 40 {
		return &errcode.HTTPErr{Status: http.StatusBadRequest, Err: fmt.Errorf("invalid commit ID %q (must be absolute, 40-char)", repoRev.CommitID)}
	}

	ctx, cl := Client(r)

	if repoRev.Rev == "" {
		repo, err := cl.Repos.Get(ctx, &repoRev.RepoSpec)
		if err != nil {
			return err
		}
		repoRev.Rev = repo.DefaultBranch
		if repo.DefaultBranch == "" {
			log15.Warn("ResolveRepoRev: no rev specified and repo has no default branch", "repo", repoRev.URI)
		}
	}

	commit, err := cl.Repos.GetCommit(ctx, repoRev)
	if err != nil {
		return err
	}
	repoRev.CommitID = string(commit.ID)
	return nil
}

// RedirectToNewRepoURI writes an HTTP redirect response with a
// Location that matches the request's location except with the
// RepoSpec route var updated to refer to newRepoURI (instead of the
// originally requested repo URI).
func RedirectToNewRepoURI(w http.ResponseWriter, r *http.Request, newRepoURI string) error {
	origVars := mux.Vars(r)
	origVars["Repo"] = (sourcegraph.RepoSpec{URI: newRepoURI}).SpecString()

	destURL, err := mux.CurrentRoute(r).URLPath(router_util.MapToArray(origVars)...)
	if err != nil {
		return err
	}

	http.Redirect(w, r, destURL.String(), http.StatusMovedPermanently)
	return nil
}

// TreeEntryCommon holds all of the tree entry-specific information necessary to
// render a tree entry page template. It is returned by getTreeEntry. It is assumes
// that pages rendered are also provided with repoCommon and
// repoRevCommon template data.
type TreeEntryCommon struct {
	EntrySpec         sourcegraph.TreeEntrySpec
	Entry             *sourcegraph.TreeEntry
	SrclibDataVersion *sourcegraph.SrclibDataVersion
}

// FlattenName flattens a nested TreeEntry name, joining with slashes.
func FlattenName(e *sourcegraph.BasicTreeEntry) string {
	if len(e.Entries) == 1 {
		return e.Name + "/" + FlattenName(e.Entries[0])
	} else {
		return e.Name
	}
}

// FlattenNameHTML flattens a nested TreeEntry name, returning HTML for rendering the slash-separated name
// with all but the last elements grayed out.
func FlattenNameHTML(e *sourcegraph.BasicTreeEntry) template.HTML {
	if len(e.Entries) == 1 {
		return template.HTML(`<span class="dim">`+template.HTMLEscapeString(e.Name)+`/</span>`) + FlattenNameHTML(e.Entries[0])
	} else {
		return template.HTML(template.HTMLEscapeString(e.Name))
	}
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

// GetTreeEntryCommon returns common data specific to the UI
// requirements for displaying a tree entry. It additionally returns
// information about the repository, the revision and build based on
// the request and the passed options.  It may also return custom
// errors URLMovedError, or NoVCSDataError.
func GetTreeEntryCommon(ctx context.Context, vars map[string]string, opt *sourcegraph.RepoTreeGetOptions) (tc *TreeEntryCommon, rc *RepoCommon, vc *RepoRevCommon, err error) {
	if opt == nil {
		opt = new(sourcegraph.RepoTreeGetOptions)
	}
	rc, vc, err = GetRepoAndRevCommon(ctx, vars)
	if err != nil {
		return tc, rc, vc, err
	}

	cl, err := sourcegraph.NewClientFromContext(ctx)
	if err != nil {
		return
	}

	tc = &TreeEntryCommon{}
	tc.EntrySpec = sourcegraph.TreeEntrySpec{
		RepoRev: vc.RepoRevSpec,
		Path:    vars["Path"],
	}

	if resolvedRev, dataVer, err := ResolveSrclibDataVersion(ctx, tc.EntrySpec); err == nil {
		tc.EntrySpec.RepoRev = resolvedRev
		tc.SrclibDataVersion = dataVer
	} else if err != nil && grpc.Code(err) != codes.NotFound {
		// Continue with existing rev and commit ID even if there's no srclib data.
		return tc, rc, vc, err
	}

	if tc.EntrySpec.RepoRev.Rev == "" {
		panic("empty Rev for repo " + tc.EntrySpec.RepoRev.URI)
	}
	if tc.EntrySpec.RepoRev.CommitID == "" {
		panic("empty CommitID for repo " + tc.EntrySpec.RepoRev.URI + " rev " + tc.EntrySpec.RepoRev.Rev)
	}

	tc.Entry, err = cl.RepoTree.Get(ctx, &sourcegraph.RepoTreeGetOp{Entry: tc.EntrySpec, Opt: opt})
	if err != nil {
		return
	}

	return
}

// GetDefCommon returns common information about a definition, based
// on the route vars.  It additionally returns common repository and
// revision information. It may also return custom errors
// URLMovedError, or NoVCSDataError.
//
// dc.Def.DefKey will be set to the def specification based on the
// request when getting actual def fails.
func GetDefCommon(ctx context.Context, vars map[string]string, opt *sourcegraph.DefGetOptions) (dc *payloads.DefCommon, rc *RepoCommon, vc *RepoRevCommon, err error) {
	defSpec := sourcegraph.DefSpec{
		Repo:     vars["Repo"],
		Unit:     vars["Unit"],
		UnitType: vars["UnitType"],
		Path:     router_util.EscapePath(vars["Path"]),
	}
	// If we fail to get a def, return the best known information to the caller.
	dc = &payloads.DefCommon{
		Def: &sourcegraph.Def{
			Def: graph.Def{
				DefKey: graph.DefKey{
					Repo:     defSpec.Repo,
					Unit:     defSpec.Unit,
					UnitType: defSpec.UnitType,
					Path:     defSpec.Path,
				},
			},
		},
	}

	rc, vc, err = GetRepoAndRevCommon(ctx, vars)
	if err != nil {
		return dc, rc, vc, err
	}

	cl, err := sourcegraph.NewClientFromContext(ctx)
	if err != nil {
		return
	}

	resolvedRev, _, err := ResolveSrclibDataVersion(ctx, sourcegraph.TreeEntrySpec{RepoRev: vc.RepoRevSpec})
	if err != nil {
		return dc, rc, vc, err
	}
	vc.RepoRevSpec.CommitID = resolvedRev.CommitID
	defSpec.CommitID = resolvedRev.CommitID

	if vc.RepoRevSpec.Rev == "" {
		panic("empty Rev for repo " + vc.RepoRevSpec.URI)
	}
	if vc.RepoRevSpec.CommitID == "" {
		panic("empty CommitID for repo " + vc.RepoRevSpec.URI + " rev " + vc.RepoRevSpec.Rev)
	}

	// Insert additional available information into the def.
	dc.Def.Def.DefKey.CommitID = defSpec.CommitID

	def, err := cl.Defs.Get(ctx, &sourcegraph.DefsGetOp{Def: defSpec, Opt: opt})
	if err != nil {
		return dc, rc, vc, err
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
		RepoRev: sourcegraph.RepoRevSpec{
			RepoSpec: vc.RepoRevSpec.RepoSpec,
			Rev:      vc.RepoRevSpec.Rev,
			CommitID: vc.RepoRevSpec.Rev, // use originally requested rev, not already resolved last-srclib-version
		},
		Path: def.File,
	})
	if err != nil {
		return dc, rc, vc, err
	}
	if defResolvedRev.CommitID != resolvedRev.CommitID {
		return dc, rc, vc, &errcode.HTTPErr{
			Status: http.StatusNotFound,
			Err:    fmt.Errorf("no srclib data for def %v (file %s was modified between last srclib analysis version %s and rev %s)", defSpec, def.File, resolvedRev.CommitID, vc.RepoRevSpec.Rev),
		}
	}

	// this can not be moved to svc/local, because HTML sanitation needs to
	// happen on the local sourcegraph instance, not on an untrusted
	// server
	if len(def.Docs) > 0 {
		defDoc := def.Docs[0]
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
		def.DocHTML = htmlutil.SanitizeForPB(docHTML)
	}

	qualifiedName := sourcecode.DefQualifiedNameAndType(def, "scope")
	qualifiedName = sourcecode.OverrideStyleViaRegexpFlags(qualifiedName)
	dc = &payloads.DefCommon{
		Def:               def,
		QualifiedName:     htmlutil.SanitizeForPB(string(qualifiedName)),
		URL:               router.Rel.URLToDefAtRev(def.DefKey, vc.RepoRevSpec.Rev).String(),
		File:              sourcegraph.TreeEntrySpec{RepoRev: vc.RepoRevSpec, Path: def.File},
		ByteStartPosition: def.DefStart,
		ByteEndPosition:   def.DefEnd,
		Found:             true,
	}
	return dc, rc, vc, nil
}

func GetRepoTreeListCommon(ctx context.Context, vars map[string]string) (*sourcegraph.RepoTreeListResult, error) {
	repoRevSpec, err := sourcegraph.UnmarshalRepoRevSpec(vars)
	if err != nil {
		return nil, err
	}

	cl, err := sourcegraph.NewClientFromContext(ctx)
	if err != nil {
		return nil, err
	}
	return cl.RepoTree.List(ctx, &sourcegraph.RepoTreeListOp{Rev: repoRevSpec})
}
