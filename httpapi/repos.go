package httpapi

import (
	"errors"
	"net/http"
	"time"

	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/util/handlerutil"
	"src.sourcegraph.com/sourcegraph/util/httputil/httpctx"
)

func serveRepo(w http.ResponseWriter, r *http.Request) error {
	s := handlerutil.APIClient(r)

	repo, _, err := handlerutil.GetRepo(r, s.Repos)
	if err != nil {
		return err
	}

	if clientCached, err := writeCacheHeaders(w, r, repo.UpdatedAt.Time(), defaultCacheMaxAge); clientCached || err != nil {
		return err
	}

	return writeJSON(w, repo)
}

func serveRepos(w http.ResponseWriter, r *http.Request) error {
	ctx := httpctx.FromRequest(r)
	s := handlerutil.APIClient(r)

	var opt sourcegraph.RepoListOptions
	err := schemaDecoder.Decode(&opt, r.URL.Query())
	if err != nil {
		return err
	}

	repos, err := s.Repos.List(ctx, &opt)
	if err != nil {
		return err
	}

	writePaginationHeader(w, r.URL, opt.ListOptions, 0 /* TODO */)
	if clientCached, err := writeCacheHeaders(w, r, time.Time{}, defaultCacheMaxAge); clientCached || err != nil {
		return err
	}

	return writeJSON(w, repos)
}

// getRepoLastBuildTime returns the time of the newest build for the
// specified repository and commitID. For performance reasons, commitID is
// assumed to be canonical (and is not resolved); if not 40 characters, an error is
// returned.
func getRepoLastBuildTime(r *http.Request, repoSpec sourcegraph.RepoSpec, commitID string) (time.Time, error) {
	if len(commitID) != 40 {
		return time.Time{}, errors.New("refusing (for performance reasons) to get the last build time for non-canonical repository commit ID")
	}

	ctx := httpctx.FromRequest(r)
	cl := handlerutil.APIClient(r)

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
