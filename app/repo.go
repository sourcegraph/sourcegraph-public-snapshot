package app

import (
	"time"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"gopkg.in/inconshreveable/log15.v2"

	"net/http"
	"net/url"
	"strings"

	"github.com/rogpeppe/rog-go/parallel"
	"github.com/sourcegraph/mux"
	"sourcegraph.com/sourcegraph/sourcegraph/app/internal"
	"sourcegraph.com/sourcegraph/sourcegraph/app/internal/tmpl"
	"sourcegraph.com/sourcegraph/sourcegraph/app/router"
	"sourcegraph.com/sourcegraph/sourcegraph/conf"
	"sourcegraph.com/sourcegraph/sourcegraph/errcode"
	"sourcegraph.com/sourcegraph/sourcegraph/go-sourcegraph/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/repoupdater"
	"sourcegraph.com/sourcegraph/sourcegraph/util/githubutil"
	"sourcegraph.com/sourcegraph/sourcegraph/util/handlerutil"
)

func init() {
	internal.RegisterErrorHandlerForType(&handlerutil.URLMovedError{}, func(w http.ResponseWriter, r *http.Request, err error) error {
		return handlerutil.RedirectToNewRepoURI(w, r, err.(*handlerutil.URLMovedError).NewURL)
	})

	internal.RegisterErrorHandlerForType(&handlerutil.NoVCSDataError{}, func(w http.ResponseWriter, r *http.Request, err error) error {
		return renderRepoNoVCSDataTemplate(w, r, err.(*handlerutil.NoVCSDataError).RepoCommon)
	})
}

func serveRepo(w http.ResponseWriter, r *http.Request) error {
	repoSpec, err := sourcegraph.UnmarshalRepoSpec(mux.Vars(r))
	if err != nil {
		return err
	}

	ctx, cl := handlerutil.Client(r)

	// Resolve repo path, and create local mirror for remote repo if
	// needed.
	res, err := cl.Repos.Resolve(ctx, &sourcegraph.RepoResolveOp{Path: repoSpec.URI})
	if err != nil && grpc.Code(err) != codes.NotFound {
		return err
	}
	if remoteRepo := res.GetRemoteRepo(); remoteRepo != nil {
		if actualURI := githubutil.RepoURI(remoteRepo.Owner, remoteRepo.Name); actualURI != repoSpec.URI {
			http.Redirect(w, r, router.Rel.URLToRepo(actualURI).String(), http.StatusMovedPermanently)
			return nil
		}

		// Automatically create a local mirror.
		log15.Info("Creating a local mirror of remote repo", "cloneURL", remoteRepo.HTTPCloneURL)
		_, err := cl.Repos.Create(ctx, &sourcegraph.ReposCreateOp{
			Op: &sourcegraph.ReposCreateOp_FromGitHubID{FromGitHubID: int32(remoteRepo.GitHubID)},
		})
		if err != nil {
			return err
		}
	}

	rc, vc, err := handlerutil.GetRepoAndRevCommon(ctx, mux.Vars(r))
	if err != nil {
		if noVCSDataErr, ok := err.(*handlerutil.NoVCSDataError); ok && noVCSDataErr.RepoCommon.Repo.Mirror {
			// Trigger cloning/updating this repo from its remote
			// mirror if it has one. Only wait a short time. That's
			// usually enough to see if it failed immediately with an
			// error, but it lets us avoid blocking on the entire
			// clone process.
			ctx, cancel := context.WithTimeout(ctx, time.Second*1)
			defer cancel()
			if _, err := cl.MirrorRepos.RefreshVCS(ctx, &sourcegraph.MirrorReposRefreshVCSOp{Repo: vc.RepoRevSpec.RepoSpec}); err != nil {
				if ctx.Err() == context.DeadlineExceeded {
					return noVCSDataErr
				}
				return err
			}
		}

		// Even if the RefreshVCS call above succeeded within the
		// timeout, still return the no-VCS-data error and display the
		// interstitial. This avoids having multiple complex code
		// paths and having to retry VCS operations (which is tricky
		// to do right).
		return err
	}

	var readme *sourcegraph.Readme
	var tree *sourcegraph.TreeEntry
	var treeEntrySpec sourcegraph.TreeEntrySpec
	if vc.RepoCommit != nil {
		treeEntrySpec = sourcegraph.TreeEntrySpec{RepoRev: vc.RepoRevSpec, Path: "."}
		run := parallel.NewRun(2)
		run.Do(func() (err error) {
			readme, err = cl.Repos.GetReadme(ctx, &vc.RepoRevSpec)
			if errcode.IsHTTPErrorCode(err, http.StatusNotFound) {
				// Lack of a readme is not a fatal error.
				err = nil
				readme = nil
			}
			return
		})
		run.Do(func() (err error) {
			opt := sourcegraph.RepoTreeGetOptions{GetFileOptions: sourcegraph.GetFileOptions{
				RecurseSingleSubfolderLimit: 200,
			}}
			tree, err = cl.RepoTree.Get(ctx, &sourcegraph.RepoTreeGetOp{Entry: treeEntrySpec, Opt: &opt})
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
		repoupdater.Enqueue(rc.Repo, handlerutil.UserFromContext(ctx))
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

func repoMetaDescription(rp *sourcegraph.Repo) string {
	desc := "Docs and usage examples for " + rp.Name
	if rp.Description != "" {
		desc += ": " + rp.Description
	}
	return desc
}
