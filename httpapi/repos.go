package httpapi

import (
	"errors"
	"net/http"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"gopkg.in/inconshreveable/log15.v2"

	"github.com/gorilla/mux"

	"sourcegraph.com/sourcegraph/sourcegraph/go-sourcegraph/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/util/githubutil"
	"sourcegraph.com/sourcegraph/sourcegraph/util/handlerutil"
)

func serveRepo(w http.ResponseWriter, r *http.Request) error {
	ctx, _ := handlerutil.Client(r)

	cl, err := sourcegraph.NewClientFromContext(ctx)
	if err != nil {
		return err
	}
	repoSpec, err := sourcegraph.UnmarshalRepoSpec(mux.Vars(r))
	if err != nil {
		return err
	}

	isCloning := false
	res, err := cl.Repos.Resolve(ctx, &sourcegraph.RepoResolveOp{Path: repoSpec.URI})
	if err != nil && grpc.Code(err) != codes.NotFound {
		return err
	}
	if remoteRepo := res.GetRemoteRepo(); remoteRepo != nil {
		if actualURI := githubutil.RepoURI(remoteRepo.Owner, remoteRepo.Name); actualURI != repoSpec.URI {
			return &handlerutil.URLMovedError{actualURI}
		}
		// Automatically create a local mirror.
		log15.Info("Creating a local mirror of remote repo", "cloneURL", remoteRepo.HTTPCloneURL)
		_, err := cl.Repos.Create(ctx, &sourcegraph.ReposCreateOp{
			Op: &sourcegraph.ReposCreateOp_FromGitHubID{FromGitHubID: int32(remoteRepo.GitHubID)},
		})
		if err != nil {
			return err
		}

		// Trigger mirror repo refresh (which kicks off the initial build).
		_, err = cl.MirrorRepos.RefreshVCS(ctx, &sourcegraph.MirrorReposRefreshVCSOp{
			Repo: repoSpec,
		})
		if err != nil {
			return err
		}

		isCloning = true
	}

	repo, _, err := handlerutil.GetRepo(ctx, mux.Vars(r))
	if err != nil {
		return err
	}

	var lastMod time.Time
	if repo.UpdatedAt != nil {
		lastMod = repo.UpdatedAt.Time()
	}

	if clientCached, err := writeCacheHeaders(w, r, lastMod, defaultCacheMaxAge); clientCached || err != nil {
		return err
	}

	response := struct {
		*sourcegraph.Repo
		IsCloning bool
	}{
		Repo:      repo,
		IsCloning: isCloning,
	}

	return writeJSON(w, &response)
}

func serveRepos(w http.ResponseWriter, r *http.Request) error {
	ctx, cl := handlerutil.Client(r)

	var opt sourcegraph.RepoListOptions
	err := schemaDecoder.Decode(&opt, r.URL.Query())
	if err != nil {
		return err
	}

	repos, err := cl.Repos.List(ctx, &opt)
	if err != nil {
		return err
	}

	writePaginationHeader(w, r.URL, opt.ListOptions, 0 /* TODO */)
	if clientCached, err := writeCacheHeaders(w, r, time.Time{}, defaultCacheMaxAge); clientCached || err != nil {
		return err
	}

	return writeJSON(w, repos)
}

func serveRemoteRepos(w http.ResponseWriter, r *http.Request) error {
	ctx, cl := handlerutil.Client(r)

	var err error
	var reposOnPage *sourcegraph.RemoteRepoList
	var remoteRepos = &sourcegraph.RemoteRepoList{}
	for page := 1; ; page++ {
		reposOnPage, err = cl.Repos.ListRemote(ctx, &sourcegraph.ReposListRemoteOptions{
			ListOptions: sourcegraph.ListOptions{PerPage: 100, Page: int32(page)},
		})
		if err != nil {
			break
		}

		if len(reposOnPage.RemoteRepos) == 0 {
			break
		}
		remoteRepos.RemoteRepos = append(remoteRepos.RemoteRepos, reposOnPage.RemoteRepos...)
	}

	// true if the user has not yet linked GitHub
	isAuthError := func(err error) bool {
		return grpc.Code(err) == codes.Unauthenticated || grpc.Code(err) == codes.PermissionDenied
	}
	if err != nil && !isAuthError(err) {
		return err
	}

	response := struct {
		*sourcegraph.RemoteRepoList
		HasLinkedGitHub bool
	}{
		RemoteRepoList:  remoteRepos,
		HasLinkedGitHub: err == nil,
	}

	return writeJSON(w, &response)
}

// getRepoLastBuildTime returns the time of the newest build for the
// specified repository and commitID. For performance reasons, commitID is
// assumed to be canonical (and is not resolved); if not 40 characters, an error is
// returned.
func getRepoLastBuildTime(r *http.Request, repoSpec sourcegraph.RepoSpec, commitID string) (time.Time, error) {
	if len(commitID) != 40 {
		return time.Time{}, errors.New("refusing (for performance reasons) to get the last build time for non-canonical repository commit ID")
	}

	ctx, cl := handlerutil.Client(r)

	builds, err := cl.Builds.List(ctx, &sourcegraph.BuildListOptions{
		Repo:        repoSpec.URI,
		CommitID:    commitID,
		Ended:       true,
		Succeeded:   true,
		ListOptions: sourcegraph.ListOptions{Page: 1, PerPage: 1},
	})
	if err != nil {
		return time.Time{}, err
	}
	if len(builds.Builds) == 1 {
		build := builds.Builds[0]
		if build.EndedAt != nil {
			return build.EndedAt.Time(), nil
		}
	}
	return time.Time{}, nil
}
