package handlerutil

import (
	"bytes"
	"go/doc"
	"html/template"
	"net/http"
	"strings"
	"time"

	"golang.org/x/net/context"

	"github.com/sourcegraph/mux"
	"sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"sourcegraph.com/sourcegraph/go-vcs/vcs"
	"sourcegraph.com/sourcegraph/srclib/graph"
	"sourcegraph.com/sourcegraph/vcsstore/vcsclient"
	"src.sourcegraph.com/sourcegraph/app/router"
	"src.sourcegraph.com/sourcegraph/sourcecode"
	"src.sourcegraph.com/sourcegraph/ui/payloads"
	"src.sourcegraph.com/sourcegraph/util/htmlutil"
	"src.sourcegraph.com/sourcegraph/util/httputil/httpctx"
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

// GetRepoCommonOpt values configure calls to GetRepoCommon.
type GetRepoCommonOpt struct {
	// AllowNonEnabledRepos causes GetRepoCommon to NOT display the "repo
	// must be enabled" interstitial for repos that are not enabled.
	AllowNonEnabledRepos bool
}

// GetRepoAndRevCommon returns the repository and RepoRevSpec
// based on the request URL. It may also return custom error types URLMovedError,
// RepoNotEnabledError and NoVCSDataError, which callers should ideally check for.
//
// If no repo is found, it is attempted to be fetched automatically
// (e.g., from the GitHub API if it's a "github.com/user/repo"-style
// repo URI).
func GetRepoAndRevCommon(r *http.Request, opts *GetRepoCommonOpt) (rc *RepoCommon, vc *RepoRevCommon, err error) {
	if opts == nil {
		opts = new(GetRepoCommonOpt)
	}

	rc, err = GetRepoCommon(r, nil)
	if err != nil {
		return
	}

	vc = &RepoRevCommon{}
	vc.RepoRevSpec.RepoSpec = rc.Repo.RepoSpec()

	apiclient := APIClient(r)
	ctx := httpctx.FromRequest(r)

	var commit0 *vcs.Commit
	vc.RepoRevSpec, commit0, err = GetRepoRev(r, apiclient.Repos, rc.Repo)
	if IsRepoNoVCSDataError(err) {
		if rc.Repo.Mirror {
			// Trigger cloning/updating this repo from its remote
			// mirror if it has one. Only wait 1 second. That's
			// usually enough to see if it failed immediately with an
			// error, but it lets us avoid blocking on the entire
			// clone process.
			ctx, cancel := context.WithTimeout(ctx, time.Second*1)
			defer cancel()
			if _, err = apiclient.MirrorRepos.RefreshVCS(ctx, &sourcegraph.MirrorReposRefreshVCSOp{Repo: vc.RepoRevSpec.RepoSpec}); err != nil {
				if ctx.Err() == context.DeadlineExceeded {
					// If deadline exceeded, fall through to NoVCSDataError return below.
				} else {
					return
				}
			}
		}
		if err != nil {
			return nil, nil, &NoVCSDataError{rc}
		}
	} else if err != nil {
		return
	}

	if commit0 != nil {
		vc.RepoCommit, err = AugmentCommit(r, rc.Repo.URI, commit0)
		if err != nil {
			return
		}
	}

	return
}

func IsRepoNoVCSDataError(err error) bool {
	return err != nil && (strings.Contains(err.Error(), "vcsstore") || IsHTTPErrorCode(err, http.StatusNotFound) ||
		strings.Contains(err.Error(), "has no default branch"))
}

// GetRepoCommon returns the repository and RepoSpec based on the request URL.
// Callers should ideally handle custom error types URLMovedError or RepoNotEnabledError.
//
// If no repo is found, it is attempted to be fetched automatically (e.g., from
// the GitHub API if it's a "github.com/user/repo"-style repo URI).
func GetRepoCommon(r *http.Request, opts *GetRepoCommonOpt) (rc *RepoCommon, err error) {
	if opts == nil {
		opts = new(GetRepoCommonOpt)
	}

	apiclient := APIClient(r)

	rc = &RepoCommon{}
	rc.Repo, _, err = GetRepo(r, apiclient.Repos)
	if err != nil {
		return
	}

	ctx := httpctx.FromRequest(r)
	repoSpec := rc.Repo.RepoSpec()
	rc.RepoConfig, err = apiclient.Repos.GetConfig(ctx, &repoSpec)
	if err != nil {
		return
	}
	if !opts.AllowNonEnabledRepos && !rc.RepoConfig.Enabled {
		return nil, &RepoNotEnabledError{rc}
	}

	return
}

// GetRepo gets the repo (from the reposSvc) specified in the URL's
// RepoSpec route param. Callers should ideally check for a return error of type
// URLMovedError and handle this scenario by warning or redirecting the user.
func GetRepo(r *http.Request, reposSvc sourcegraph.ReposClient) (repo *sourcegraph.Repo, repoSpec sourcegraph.RepoSpec, err error) {
	origRepoSpec, err := sourcegraph.UnmarshalRepoSpec(mux.Vars(r))
	if err != nil {
		return nil, sourcegraph.RepoSpec{}, err
	}

	repoSpec = origRepoSpec
	repo, err = reposSvc.Get(httpctx.FromRequest(r), &repoSpec)
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

// GetRepoRev resolves the RepoRevSpec and commit (from the reposSvc)
// specified in the URL's RepoRevSpec route param. The provided repo's
// DefaultBranch is used in case no revspec is present in the URL.
func GetRepoRev(r *http.Request, reposSvc sourcegraph.ReposClient, repo *sourcegraph.Repo) (sourcegraph.RepoRevSpec, *vcs.Commit, error) {
	repoRev, err := sourcegraph.UnmarshalRepoRevSpec(mux.Vars(r))
	if err != nil {
		return sourcegraph.RepoRevSpec{RepoSpec: repo.RepoSpec()}, nil, err
	}

	commit, err := reposSvc.GetCommit(httpctx.FromRequest(r), &repoRev)
	if err != nil {
		return repoRev, nil, err
	}
	repoRev.CommitID = string(commit.ID)

	if repoRev.Rev == "" {
		repoRev.Rev = repo.DefaultBranch
	}

	return repoRev, commit, nil
}

// GetRepoAndRev returns the Repo and the RepoRevSpec for a repository. It may
// also return custom error URLMovedError to allow special handling of this case,
// such as for example redirecting the user.
func GetRepoAndRev(r *http.Request, reposSvc sourcegraph.ReposClient) (repo *sourcegraph.Repo, repoRevSpec sourcegraph.RepoRevSpec, commit *vcs.Commit, err error) {
	repo, repoRevSpec.RepoSpec, err = GetRepo(r, reposSvc)
	if err != nil {
		return repo, repoRevSpec, nil, err
	}
	repoRevSpec, commit, err = GetRepoRev(r, reposSvc, repo)
	return repo, repoRevSpec, commit, err
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
	EntrySpec sourcegraph.TreeEntrySpec
	Entry     *sourcegraph.TreeEntry
}

// FlattenName flattens a nested TreeEntry name, joining with slashes.
func FlattenName(e *vcsclient.TreeEntry) string {
	if len(e.Entries) == 1 {
		return e.Name + "/" + FlattenName(e.Entries[0])
	} else {
		return e.Name
	}
}

// FlattenNameHTML flattens a nested TreeEntry name, returning HTML for rendering the slash-separated name
// with all but the last elements grayed out.
func FlattenNameHTML(e *vcsclient.TreeEntry) template.HTML {
	if len(e.Entries) == 1 {
		return template.HTML(`<span class="dim">`+template.HTMLEscapeString(e.Name)+`/</span>`) + FlattenNameHTML(e.Entries[0])
	} else {
		return template.HTML(template.HTMLEscapeString(e.Name))
	}
}

// GetTreeEntryCommon returns common data specific to the UI requirements for
// displaying a tree entry. It additionally returns information about the
// repository, the revision and build based on the request and the passed options.
// It may also return custom errors URLMovedError, RepoNotEnabledError, NoBuildError or
// NoVCSDataError.
func GetTreeEntryCommon(r *http.Request, opt *sourcegraph.RepoTreeGetOptions) (tc *TreeEntryCommon, rc *RepoCommon, vc *RepoRevCommon, bc RepoBuildCommon, err error) {
	if opt == nil {
		opt = new(sourcegraph.RepoTreeGetOptions)
	}
	rc, vc, err = GetRepoAndRevCommon(r, nil)
	if err != nil {
		return tc, rc, vc, bc, err
	}

	apiclient := APIClient(r)

	bc, err = GetRepoBuildCommon(r, rc, vc, &GetRepoBuildCommonOpt{AllowUnbuilt: true})
	if err != nil {
		return tc, rc, vc, bc, err
	}
	v := mux.Vars(r)
	tc = &TreeEntryCommon{
		EntrySpec: sourcegraph.TreeEntrySpec{
			RepoRev: bc.BestRevSpec,
			Path:    v["Path"],
		},
	}

	// Only request code formatting if it's likely to be a code file.
	if err = schemaDecoder.Decode(opt, r.URL.Query()); err != nil {
		return
	}
	if !opt.Formatted && !opt.TokenizedSource {
		opt.Formatted = sourcecode.IsLikelyCodeFile(tc.EntrySpec.Path)
	}
	if !opt.TokenizedSource {
		opt.TokenizedSource = r.Header.Get("Content-Type") == "application/json"
	}

	ctx := httpctx.FromRequest(r)
	tc.Entry, err = apiclient.RepoTree.Get(ctx, &sourcegraph.RepoTreeGetOp{Entry: tc.EntrySpec, Opt: opt})
	if err != nil {
		return
	}

	return
}

// GetDefCommon returns common information about a definition, based on the request.
// It additionally returns common build, repository and revision information. It may
// also return custom errors URLMovedError, RepoNotEnabledError, NoBuildError or NoVCSDataError.
//
// dc.Def.DefKey will be set to the def specification based on the request when getting actual def fails.
func GetDefCommon(r *http.Request, opt *sourcegraph.DefGetOptions) (dc *payloads.DefCommon, bc RepoBuildCommon, rc *RepoCommon, vc *RepoRevCommon, err error) {
	v := mux.Vars(r)
	defSpec := sourcegraph.DefSpec{
		Repo:     v["Repo"],
		Unit:     v["Unit"],
		UnitType: v["UnitType"],
		Path:     router_util.EscapePath(v["Path"]),
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

	rc, vc, err = GetRepoAndRevCommon(r, nil)
	if err != nil {
		return dc, bc, rc, vc, err
	}
	bc, err = GetRepoBuildCommon(r, rc, vc, &GetRepoBuildCommonOpt{
		// If the user didn't specify a revision for the definition, then we use
		// inexact matching for the build data. This means we can use the build data
		// from e.g. just a few commits behind HEAD if HEAD hasn't been built yet or
		// had a build failure.
		Inexact: len(v["Rev"]) == 0,
	})
	if err != nil {
		return dc, bc, rc, vc, err
	}

	cl := APIClient(r)

	vc.RepoRevSpec = bc.BestRevSpec // Remove after getRepo refactor.
	defSpec.CommitID = bc.BestRevSpec.CommitID

	// Insert additional available information into the def.
	dc.Def.Def.DefKey.CommitID = defSpec.CommitID

	def, err := cl.Defs.Get(httpctx.FromRequest(r), &sourcegraph.DefsGetOp{Def: defSpec, Opt: opt})
	if err != nil {
		return dc, bc, rc, vc, err
	}

	// this can not be moved to svc/local, because HTML sanitation needs to
	// happen on the local sourcegraph instance, not on an untrusted
	// federation remote
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
	return dc, bc, rc, vc, nil
}

func GetRepoTreeListCommon(r *http.Request) (*sourcegraph.RepoTreeListResult, error) {
	repoRevSpec, err := sourcegraph.UnmarshalRepoRevSpec(mux.Vars(r))
	if err != nil {
		return nil, err
	}

	cl := APIClient(r)
	return cl.RepoTree.List(httpctx.FromRequest(r), &sourcegraph.RepoTreeListOp{Rev: repoRevSpec})
}
