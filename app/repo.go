package app

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	"code.google.com/p/rog-go/parallel"
	"github.com/sourcegraph/mux"
	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"sourcegraph.com/sourcegraph/vcsstore/vcsclient"
	"src.sourcegraph.com/sourcegraph/app/appconf"
	"src.sourcegraph.com/sourcegraph/app/internal"
	"src.sourcegraph.com/sourcegraph/app/internal/schemautil"
	"src.sourcegraph.com/sourcegraph/app/internal/tmpl"
	"src.sourcegraph.com/sourcegraph/app/router"
	"src.sourcegraph.com/sourcegraph/conf"
	"src.sourcegraph.com/sourcegraph/errcode"
	"src.sourcegraph.com/sourcegraph/ext"
	"src.sourcegraph.com/sourcegraph/util"
	"src.sourcegraph.com/sourcegraph/util/buildutil"
	"src.sourcegraph.com/sourcegraph/util/cacheutil"
	"src.sourcegraph.com/sourcegraph/util/handlerutil"
	"src.sourcegraph.com/sourcegraph/util/httputil/httpctx"
)

func init() {
	internal.RegisterErrorHandlerForType(&handlerutil.URLMovedError{}, func(w http.ResponseWriter, r *http.Request, err error) error {
		return handlerutil.RedirectToNewRepoURI(w, r, err.(*handlerutil.URLMovedError).NewURL)
	})

	internal.RegisterErrorHandlerForType(&handlerutil.RepoNotEnabledError{}, func(w http.ResponseWriter, r *http.Request, err error) error {
		return renderRepoNotEnabledTemplate(w, r, err.(*handlerutil.RepoNotEnabledError).RepoCommon)
	})

	internal.RegisterErrorHandlerForType(&handlerutil.NoVCSDataError{}, func(w http.ResponseWriter, r *http.Request, err error) error {
		return renderRepoNoVCSDataTemplate(w, r, err.(*handlerutil.NoVCSDataError).RepoCommon)
	})

	internal.RegisterErrorHandlerForType(&handlerutil.NoBuildError{}, func(w http.ResponseWriter, r *http.Request, err error) error {
		return serveRepoNoBuildError(w, r, err.(*handlerutil.NoBuildError))
	})
}

func serveRepoRefresh(w http.ResponseWriter, r *http.Request) error {
	ctx := httpctx.FromRequest(r)
	apiclient := handlerutil.APIClient(r)

	rc, err := handlerutil.GetRepoCommon(r, nil)
	if err != nil {
		return err
	}

	op := &sourcegraph.MirrorReposRefreshVCSOp{
		Repo: rc.Repo.RepoSpec(),
	}
	// For private repos, supply auth.
	if rc.Repo.Private {
		tokenStore := ext.AccessTokens{}
		token, err := tokenStore.Get(ctx, util.RepoURIHost(rc.Repo.URI))
		if err != nil {
			return err
		}

		// Setting credentials will perform this operation locally (non-federated).
		op.Credentials = &sourcegraph.VCSCredentials{
			Pass: token,
		}
	}

	if _, err := apiclient.MirrorRepos.RefreshVCS(httpctx.FromRequest(r), op); err != nil {
		return err
	}

	http.Redirect(w, r, router.Rel.URLToRepo(rc.Repo.URI).String(), http.StatusNoContent)
	return nil
}

func serveRepo(w http.ResponseWriter, r *http.Request) error {
	ctx := httpctx.FromRequest(r)
	apiclient := handlerutil.APIClient(r)

	repoSpec, err := sourcegraph.UnmarshalRepoSpec(mux.Vars(r))
	if err != nil {
		return err
	}

	// Special-case: redirect "user/repo" URLs (with no "github.com/") to the
	// path "/github.com/user/repo". This lets you use
	// "sourcegraph.com/user/repo" as your repo's URL.
	if appconf.Flags.EnableGitHubRepoShortURIAliases {
		if parts := strings.Split(repoSpec.URI, "/"); len(parts) == 2 && !strings.Contains(parts[0], ".") {
			http.Redirect(w, r, router.Rel.URLToRepo("github.com/"+repoSpec.URI).String(), http.StatusSeeOther)
			return nil
		}
	}

	// Special-case: redirect "github.com/user" URLs (with only 1 path component
	// after github.com) to the corresponding user profile URL.
	if strings.HasPrefix(repoSpec.URI, "github.com/") && strings.Count(repoSpec.URI, "/") == 1 {
		i := strings.Index(repoSpec.URI, "/")
		login := repoSpec.URI[i+1:]
		if login == "" {
			return &errcode.HTTPErr{Status: http.StatusBadRequest, Err: errors.New("bad repository URI")}
		}
		http.Redirect(w, r, router.Rel.URLToUser(login).String(), http.StatusMovedPermanently)
		return nil
	}

	// Special-case: redirect "github.com/user/repo/..." (old URLs) to
	// "github.com/user/repo".
	if strings.HasPrefix(repoSpec.URI, "github.com/") && strings.Count(repoSpec.URI, "/") > 2 {
		parts := strings.SplitN(repoSpec.URI, "/", 4)
		if len(parts) < 4 {
			return &errcode.HTTPErr{Status: http.StatusNotFound, Err: errors.New("bad github repository url")}
		}
		http.Redirect(w, r, router.Rel.URLToRepo(string(strings.Join(parts[0:3], "/"))).String(), http.StatusMovedPermanently)
		return nil
	}

	// Special-case: Redirect "/xyz" single-path-component repo URIs
	// to "/~xyz" for the live site. That's because people have
	// existing links to https://sourcegraph.com/myname and we don't
	// want to break these.
	if appconf.Flags.EnableGitHubStyleUserPaths {
		if strings.Count(repoSpec.URI, "/") == 0 {
			http.Redirect(w, r, router.Rel.URLToUser(repoSpec.URI).String(), http.StatusSeeOther)
			return nil
		}
	}

	rc, vc, err := handlerutil.GetRepoAndRevCommon(r, nil)
	if err != nil {
		return err
	}

	var bc handlerutil.RepoBuildCommon
	var readme *sourcegraph.Readme
	var tree *sourcegraph.TreeEntry
	var treeEntrySpec sourcegraph.TreeEntrySpec
	var exactBuild *sourcegraph.Build
	if vc.RepoCommit != nil {
		var err error
		bc, err = handlerutil.GetRepoBuildCommon(r, rc, vc, nil)
		if err != nil {
			return err
		}

		// If we have never built this repo, build it.
		if !appconf.Flags.NoUIBuild && !bc.Built {
			form := sourcegraph.BuildCreateOptions{
				BuildConfig: sourcegraph.BuildConfig{
					Import:   true,
					Queue:    true,
					Priority: int32(buildutil.DefaultPriority(rc.Repo.Private, buildutil.Manual)),
				},
			}
			commitID := bc.BestRevSpec.CommitID
			repoRevSpec := sourcegraph.RepoRevSpec{RepoSpec: rc.Repo.RepoSpec(), Rev: commitID, CommitID: commitID}
			_, err := apiclient.Builds.Create(ctx, &sourcegraph.BuildsCreateOp{RepoRev: repoRevSpec, Opt: &form})
			if err != nil {
				if grpc.Code(err) == codes.PermissionDenied || grpc.Code(err) == codes.Unauthenticated {
					log.Printf("cannot create build: %v\n", err)
				} else {
					return err
				}
			}
		}

		treeEntrySpec = sourcegraph.TreeEntrySpec{RepoRev: bc.BestRevSpec, Path: "."}
		run := parallel.NewRun(7)
		run.Do(func() (err error) {
			readme, err = apiclient.Repos.GetReadme(ctx, &bc.BestRevSpec)
			if errcode.IsHTTPErrorCode(err, http.StatusNotFound) {
				// Lack of a readme is not a fatal error.
				err = nil
				readme = nil
			}
			return
		})
		run.Do(func() (err error) {
			opt := sourcegraph.RepoTreeGetOptions{GetFileOptions: vcsclient.GetFileOptions{
				RecurseSingleSubfolderLimit: 200,
			}}
			tree, err = apiclient.RepoTree.Get(ctx, &sourcegraph.RepoTreeGetOp{Entry: treeEntrySpec, Opt: &opt})
			if err == nil {
				tree_ := *tree
				go cacheutil.PrecacheTreeEntry(apiclient, ctx, &tree_, treeEntrySpec)
			}

			return
		})
		if err := run.Wait(); err != nil {
			return err
		}

		build, err := handlerutil.APIClient(r).Builds.GetRepoBuildInfo(ctx, &sourcegraph.BuildsGetRepoBuildInfoOp{
			Repo: vc.RepoRevSpec,
			Opt: &sourcegraph.BuildsGetRepoBuildInfoOptions{
				Exact: true,
			},
		})
		if err != nil && !errcode.IsHTTPErrorCode(err, http.StatusNotFound) {
			return err
		}
		if err == nil {
			exactBuild = build.Exact
		}
	}

	// The canonical URL for the repo's default branch is the URL
	// without an "@revspec" (like "@master").
	var canonicalURL *url.URL
	if vc.RepoRevSpec.Rev == rc.Repo.DefaultBranch {
		canonicalURL = conf.AppURL(ctx).ResolveReference(router.Rel.URLToRepo(rc.Repo.URI))
	}

	if rc.Repo.Mirror {
		RepoUpdater.enqueue(rc.Repo)
	}

	return tmpl.Exec(r, w, "repo/main.html", http.StatusOK, nil, &struct {
		handlerutil.RepoCommon
		handlerutil.RepoRevCommon
		handlerutil.RepoBuildCommon
		ExactBuild *sourcegraph.Build
		Readme     *sourcegraph.Readme
		EntryPath  string
		Entry      *sourcegraph.TreeEntry
		EntrySpec  sourcegraph.TreeEntrySpec

		HasVCSData bool

		RobotsIndex bool
		tmpl.Common
	}{
		RepoCommon:      *rc,
		RepoRevCommon:   *vc,
		RepoBuildCommon: bc,
		ExactBuild:      exactBuild,
		Readme:          readme,
		EntryPath:       ".",
		Entry:           tree,
		EntrySpec:       treeEntrySpec,

		HasVCSData: vc.RepoCommit != nil,

		RobotsIndex: !rc.Repo.Private,

		Common: tmpl.Common{
			CanonicalURL: canonicalURL,
		},
	})
}

func serveRepoSearch(w http.ResponseWriter, r *http.Request) error {
	ctx := httpctx.FromRequest(r)
	//TODO remove this and implement proper pagnination for search results.
	const maxResults = 10

	var opt sourcegraph.SearchOptions
	err := schemautil.Decode(&opt, r.URL.Query())
	if err != nil {
		return err
	}

	apiclient := handlerutil.APIClient(r)

	// TODO(sqs): This could be optimized since we're calling it via PJAX and
	// don't need to look all this up each time. #*perf

	rc, vc, err := handlerutil.GetRepoAndRevCommon(r, nil)
	if err != nil {
		return err
	}

	bc, err := handlerutil.GetRepoBuildCommon(r, rc, vc, nil)
	if err != nil {
		return err
	}
	vc.RepoRevSpec = bc.BestRevSpec // Remove after getRepo refactor.

	origOpt := opt
	if explicitRev := mux.Vars(r)["Rev"]; explicitRev != "" {
		opt.Query = fmt.Sprintf("%s :%s %s", vc.RepoRevSpec.URI, vc.RepoRevSpec.Rev, opt.Query)
	} else {
		opt.Query = fmt.Sprintf("%s %s", vc.RepoRevSpec.URI, opt.Query)
	}
	opt.Defs = true

	opt.Tree = !appconf.Flags.DisableRepoTreeSearch
	opt.ListOptions.PerPage = maxResults

	results, err := apiclient.Search.Search(ctx, &opt)
	if err != nil {
		return err
	}
	addPopoversToTextSearchResults(results.Tree)

	return tmpl.Exec(r, w, "repo/search_results.html", http.StatusOK, nil, &struct {
		MaxResults int
		handlerutil.RepoCommon
		handlerutil.RepoRevCommon
		handlerutil.RepoBuildCommon
		SearchOptions *sourcegraph.SearchOptions
		SearchResults *sourcegraph.SearchResults
		tmpl.Common
	}{
		MaxResults:      maxResults,
		RepoCommon:      *rc,
		RepoRevCommon:   *vc,
		RepoBuildCommon: bc,
		SearchOptions:   &origOpt,
		SearchResults:   results,
	})
}

func serveRepoSearchNext(w http.ResponseWriter, r *http.Request) error {
	var search RepoSearch
	err := schemautil.Decode(&search, r.URL.Query())
	if err != nil {
		return err
	}

	rc, vc, err := handlerutil.GetRepoAndRevCommon(r, nil)
	if err != nil {
		return err
	}

	return tmpl.Exec(r, w, "repo/search.html", http.StatusOK, nil, &struct {
		RepoSearch *RepoSearch
		handlerutil.RepoCommon
		handlerutil.RepoRevCommon
		tmpl.Common
	}{
		RepoSearch:    &search,
		RepoCommon:    *rc,
		RepoRevCommon: *vc,
	})
}

func renderRepoNoVCSDataTemplate(w http.ResponseWriter, r *http.Request, rc *handlerutil.RepoCommon) error {
	return tmpl.Exec(r, w, "repo/no_vcs_data.html", http.StatusOK, nil, &struct {
		handlerutil.RepoCommon
		tmpl.Common
	}{
		RepoCommon: *rc,
	})
}

func renderRepoNotEnabledTemplate(w http.ResponseWriter, r *http.Request, rc *handlerutil.RepoCommon) error {
	return tmpl.Exec(r, w, "repo/not_enabled.html", http.StatusOK, nil, &struct {
		handlerutil.RepoCommon
		tmpl.Common
	}{
		RepoCommon: *rc,
	})
}

type RepoSearch struct {
	Query string `url:"q" schema:"q"`
}

type RepoLink struct {
	LeadingParts []string
	NamePart     string
	URL          *url.URL
	Title        string
}

// absRepoLink produces a formatted link to a repo, and links to the
// absolute URL to the repository on the current server (using
// conf.AppURL).
func absRepoLink(appURL *url.URL, repoURI string) *RepoLink {
	parts := strings.Split(repoURI, "/")
	parts[0] = strings.ToLower(parts[0])
	if (parts[0] == "github.com" || parts[0] == "sourcegraph.com") && len(parts) == 3 {
		parts = parts[1:]
	}
	return &RepoLink{
		LeadingParts: parts[:len(parts)-1],
		NamePart:     parts[len(parts)-1],
		URL:          appURL.ResolveReference(router.Rel.URLToRepo(repoURI)),
		Title:        repoURI,
	}
}

func repoLink(repoURI string) *RepoLink {
	return absRepoLink(&url.URL{}, repoURI)
}

func repoLabelForOwner(owner string, repoURI string) []string {
	if ownerPrefix := "github.com/" + owner + "/"; strings.HasPrefix(repoURI, ownerPrefix) {
		repoURI = strings.TrimPrefix(repoURI, ownerPrefix)
	} else if strings.HasPrefix(repoURI, "github.com/") {
		repoURI = strings.TrimPrefix(repoURI, "github.com/")
	} else if strings.HasPrefix(repoURI, "sourcegraph.com/") {
		repoURI = strings.TrimPrefix(repoURI, "sourcegraph.com/")
	}
	return strings.Split(repoURI, "/")
}

func repoMetaDescription(rp *sourcegraph.Repo) string {
	desc := "Docs and usage examples for " + rp.Name
	if rp.Description != "" {
		desc += ": " + rp.Description
	}
	return desc
}

func repoBasename(repoURI string) string {
	return filepath.Base(repoURI)
}

func repoStat(rp *sourcegraph.Repo, statType string) int {
	// TODO(sqs): this is a stub to make templates and go code
	// compile, it does not actually work - we need to reimplement repo
	// stats for this to work.
	return 0
}

// showRepoRevSwitcher returns whether the repo switcher (that lets you
// choose branches/tags) should be displayed on pages generated for
// this route. We only want to show it where it makes sense, when the
// data on the page is dependent on the revision.
//
// The list of routes should be all routes that let you specify a repo
// with a revision, as in "repoURI@revspec".
func showRepoRevSwitcher(routeName string) bool {
	if strings.HasPrefix(routeName, "def") || strings.HasPrefix(routeName, "repo.tree") {
		return true
	}
	switch routeName {
	case router.Repo, router.RepoBadges, router.RepoSearch:
		return true
	}
	return false
}

func repoMaybeUnsupported(repo *sourcegraph.Repo) bool {
	if !repo.IsGitHubRepo() {
		// assume all non-GitHub repos are supported, since we manually added them
		return false
	}
	switch repo.Language {
	case "Go", "JavaScript", "Python", "Ruby", "Java":
		return false
	default:
		return true
	}
}
