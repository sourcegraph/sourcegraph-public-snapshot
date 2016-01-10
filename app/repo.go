package app

import (
	"errors"

	"fmt"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"

	"gopkg.in/inconshreveable/log15.v2"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	"github.com/rogpeppe/rog-go/parallel"
	"github.com/sourcegraph/mux"
	"sourcegraph.com/sourcegraph/vcsstore/vcsclient"
	"src.sourcegraph.com/sourcegraph/app/appconf"
	"src.sourcegraph.com/sourcegraph/app/internal"
	"src.sourcegraph.com/sourcegraph/app/internal/tmpl"
	"src.sourcegraph.com/sourcegraph/app/router"
	"src.sourcegraph.com/sourcegraph/conf"
	"src.sourcegraph.com/sourcegraph/errcode"
	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/platform"
	"src.sourcegraph.com/sourcegraph/repoupdater"
	"src.sourcegraph.com/sourcegraph/util/cacheutil"
	"src.sourcegraph.com/sourcegraph/util/handlerutil"
	"src.sourcegraph.com/sourcegraph/util/httputil/httpctx"
)

func init() {
	internal.RegisterErrorHandlerForType(&handlerutil.URLMovedError{}, func(w http.ResponseWriter, r *http.Request, err error) error {
		return handlerutil.RedirectToNewRepoURI(w, r, err.(*handlerutil.URLMovedError).NewURL)
	})

	internal.RegisterErrorHandlerForType(&handlerutil.NoVCSDataError{}, func(w http.ResponseWriter, r *http.Request, err error) error {
		return renderRepoNoVCSDataTemplate(w, r, err.(*handlerutil.NoVCSDataError).RepoCommon)
	})
}

func serveRepoCreate(w http.ResponseWriter, r *http.Request) error {
	repoURI := r.PostFormValue("repo-name")
	if repoURI == "" {
		log15.Warn("No repository URI provided with repo create request")
		return errors.New("Must provide a repository name")
	}

	ctx := httpctx.FromRequest(r)

	apiclient := handlerutil.APIClient(r)

	if _, err := apiclient.Repos.Get(ctx, &sourcegraph.RepoSpec{URI: repoURI}); grpc.Code(err) != codes.NotFound {
		switch err {
		case nil:
			log15.Warn("repo already exists", "repoURI", repoURI)
			return fmt.Errorf("Repo %s already exists", repoURI)
		default:
			log15.Warn("problem fetching repository", "error", err)
			return fmt.Errorf("Problem fetching repository: %s", err)
		}
	}

	repo, err := apiclient.Repos.Create(ctx, &sourcegraph.ReposCreateOp{
		URI: repoURI,
		VCS: "git",
	})
	if err != nil {
		log15.Error("failed to create repo", "error", err)
		return err
	}

	http.Redirect(w, r, router.Rel.URLToRepo(repo.URI).String(), http.StatusSeeOther)
	return nil
}

func serveRepoRefresh(w http.ResponseWriter, r *http.Request) error {
	ctx := httpctx.FromRequest(r)
	apiclient := handlerutil.APIClient(r)

	rc, err := handlerutil.GetRepoCommon(r)
	if err != nil {
		return err
	}

	op := &sourcegraph.MirrorReposRefreshVCSOp{
		Repo: rc.Repo.RepoSpec(),
	}

	if _, err := apiclient.MirrorRepos.RefreshVCS(ctx, op); err != nil {
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

	rc, vc, err := handlerutil.GetRepoAndRevCommon(r)
	if err != nil {
		return err
	}

	var readme *sourcegraph.Readme
	var tree *sourcegraph.TreeEntry
	var treeEntrySpec sourcegraph.TreeEntrySpec
	if vc.RepoCommit != nil {
		treeEntrySpec = sourcegraph.TreeEntrySpec{RepoRev: vc.RepoRevSpec, Path: "."}
		run := parallel.NewRun(2)
		run.Do(func() (err error) {
			readme, err = apiclient.Repos.GetReadme(ctx, &vc.RepoRevSpec)
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
	}

	// The canonical URL for the repo's default branch is the URL
	// without an "@revspec" (like "@master").
	var canonicalURL *url.URL
	if vc.RepoRevSpec.Rev == rc.Repo.DefaultBranch {
		canonicalURL = conf.AppURL(ctx).ResolveReference(router.Rel.URLToRepo(rc.Repo.URI))
	}

	if rc.Repo.Mirror {
		repoupdater.Enqueue(rc.Repo)
	}

	return tmpl.Exec(r, w, "repo/main.html", http.StatusOK, nil, &struct {
		handlerutil.RepoCommon
		handlerutil.RepoRevCommon
		Readme    *sourcegraph.Readme
		EntryPath string
		Entry     *sourcegraph.TreeEntry
		EntrySpec sourcegraph.TreeEntrySpec

		HasVCSData bool

		RobotsIndex bool
		tmpl.Common
	}{
		RepoCommon:    *rc,
		RepoRevCommon: *vc,
		Readme:        readme,
		EntryPath:     ".",
		Entry:         tree,
		EntrySpec:     treeEntrySpec,

		HasVCSData: vc.RepoCommit != nil,

		RobotsIndex: !rc.Repo.Private,

		Common: tmpl.Common{
			CanonicalURL: canonicalURL,
		},
	})
}

func serveRepoSearch(w http.ResponseWriter, r *http.Request) error {
	rc, vc, err := handlerutil.GetRepoAndRevCommon(r)
	if err != nil {
		return err
	}

	return tmpl.Exec(r, w, "repo/search.html", http.StatusOK, nil, &struct {
		handlerutil.RepoCommon
		handlerutil.RepoRevCommon
		SearchFrames map[string]platform.SearchFrame
		tmpl.Common
	}{
		RepoCommon:    *rc,
		RepoRevCommon: *vc,
		SearchFrames:  platform.SearchFrames(),
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

type repoLinkInfo struct {
	LeadingParts []string
	NamePart     string
	URL          *url.URL
	Title        string
}

// absRepoLink produces a formatted link to a repo, and links to the
// absolute URL to the repository on the current server (using
// conf.AppURL).
func absRepoLink(appURL *url.URL, repoURI string) *repoLinkInfo {
	parts := strings.Split(repoURI, "/")

	if maybeHost := strings.ToLower(parts[0]); (maybeHost == "github.com" || maybeHost == "sourcegraph.com") && len(parts) == 3 {
		// Chop off "github.com" or "sourcegraph.com" prefix.
		parts = parts[1:]
	}
	return &repoLinkInfo{
		LeadingParts: parts[:len(parts)-1],
		NamePart:     parts[len(parts)-1],
		URL:          appURL.ResolveReference(router.Rel.URLToRepo(repoURI)),
		Title:        repoURI,
	}
}

func repoLink(repoURI string) *repoLinkInfo {
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
	case router.Repo, router.RepoBadges:
		return true
	}
	return false
}
