package httpapi

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"time"

	"golang.org/x/net/context"

	"gopkg.in/inconshreveable/log15.v2"

	authpkg "sourcegraph.com/sourcegraph/sourcegraph/pkg/auth"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/inventory"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/routevar"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	"github.com/gorilla/mux"

	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/errcode"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/handlerutil"
)

func serveRepo(w http.ResponseWriter, r *http.Request) error {
	ctx, cl := handlerutil.Client(r)

	repoPath := routevar.ToRepo(mux.Vars(r))
	repo, err := cl.Repos.Get(ctx, &sourcegraph.RepoSpec{URI: repoPath})
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

	return writeJSON(w, repo)
}

type repoResolution struct {
	Data         sourcegraph.RepoResolution
	IncludedRepo *sourcegraph.Repo // optimistically included repo; see serveRepoResolve
}

func serveRepoResolve(w http.ResponseWriter, r *http.Request) error {
	ctx, cl := handlerutil.Client(r)

	var op sourcegraph.RepoResolveOp
	if err := schemaDecoder.Decode(&op, r.URL.Query()); err != nil {
		return err
	}
	op.Path = routevar.ToRepo(mux.Vars(r))

	res0, err := cl.Repos.Resolve(ctx, &op)
	if err != nil {
		return err
	}

	res := repoResolution{Data: *res0}

	// As an optimization, optimistically include the full local repo
	// if the operation resolved to a local repo. Clients will almost
	// always need the local repo in this case, so including it saves
	// a round-trip.
	if res0.Repo != "" {
		repo, err := cl.Repos.Get(ctx, &sourcegraph.RepoSpec{URI: res0.Repo})
		if err == nil {
			res.IncludedRepo = repo
		} else {
			log15.Warn("Error optimistically including repo in serveRepoResolve", "repo", res0.Repo, "err", err)
		}
	}

	return writeJSON(w, res)
}

func serveRepoInventory(w http.ResponseWriter, r *http.Request) error {
	ctx, cl := handlerutil.Client(r)

	repoRev, err := resolveRepoRev(ctx, routevar.ToRepoRev(mux.Vars(r)))
	if err != nil {
		return err
	}

	res, err := cl.Repos.GetInventory(ctx, repoRev)
	if err != nil {
		return err
	}

	resp := struct {
		Languages                  []*inventory.Lang
		PrimaryProgrammingLanguage string
	}{
		Languages:                  res.Languages,
		PrimaryProgrammingLanguage: res.PrimaryProgrammingLanguage(),
	}

	return writeJSON(w, &resp)
}

func serveRepos(w http.ResponseWriter, r *http.Request) error {
	ctx, cl := handlerutil.Client(r)

	var opt sourcegraph.RepoListOptions
	err := schemaDecoder.Decode(&opt, r.URL.Query())
	if err != nil {
		return err
	}

	// The only locally hosted repos are sourcegraph repos. We want
	// to prevent these repos showing up on a users homepage, unless they
	// are Sourcegraph staff. Only Sourcegraph staff have write
	// access. This means that only we will see these repos on our
	// dashboard, which is the purpose of this if-statement. When we have
	// a fuller security model or user-selectable repo lists, we can
	// remove this.
	if !authpkg.ActorFromContext(ctx).HasWriteAccess() {
		return writeJSON(w, &sourcegraph.RepoList{})
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

func serveRepoCreate(w http.ResponseWriter, r *http.Request) error {
	ctx, cl := handlerutil.Client(r)

	var op sourcegraph.ReposCreateOp
	if err := json.NewDecoder(r.Body).Decode(&op); err != nil {
		if err == io.EOF {
			return &errcode.HTTPErr{Status: http.StatusBadRequest}
		}
		return err
	}

	repo, err := cl.Repos.Create(ctx, &op)
	if err != nil {
		return err
	}
	return writeJSON(w, &repo)
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

	return writeJSON(w, remoteRepos)
}

// getRepoLastBuildTime returns the time of the newest build for the
// specified repository and commitID. For performance reasons, commitID is
// assumed to be canonical (and is not resolved); if not 40 characters, an error is
// returned.
func getRepoLastBuildTime(r *http.Request, repo, commitID string) (time.Time, error) {
	if len(commitID) != 40 {
		return time.Time{}, errors.New("refusing (for performance reasons) to get the last build time for non-canonical repository commit ID")
	}

	ctx, cl := handlerutil.Client(r)

	builds, err := cl.Builds.List(ctx, &sourcegraph.BuildListOptions{
		Repo:        repo,
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

func resolveRepoRev(ctx context.Context, repoRev routevar.RepoRev) (*sourcegraph.RepoRevSpec, error) {
	cl, err := sourcegraph.NewClientFromContext(ctx)
	if err != nil {
		return nil, err
	}
	res, err := cl.Repos.ResolveRev(ctx, &sourcegraph.ReposResolveRevOp{Repo: repoRev.Repo, Rev: repoRev.Rev})
	if err != nil {
		return nil, err
	}
	return &sourcegraph.RepoRevSpec{
		Repo:     repoRev.Repo,
		CommitID: res.CommitID,
	}, nil
}
