package httpapi

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"reflect"
	"strconv"
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
	ctx, _ := handlerutil.Client(r)

	repo, err := handlerutil.GetRepo(ctx, mux.Vars(r))
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
	if res0.Repo != 0 {
		repo, err := cl.Repos.Get(ctx, &sourcegraph.RepoSpec{ID: res0.Repo})
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

	repoRev, err := resolveLocalRepoRev(ctx, routevar.ToRepoRev(mux.Vars(r)))
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
func getRepoLastBuildTime(r *http.Request, repoPath string, commitID string) (time.Time, error) {
	if len(commitID) != 40 {
		return time.Time{}, errors.New("refusing (for performance reasons) to get the last build time for non-canonical repository commit ID")
	}

	ctx, cl := handlerutil.Client(r)

	res, err := cl.Repos.Resolve(ctx, &sourcegraph.RepoResolveOp{Path: repoPath})
	if err != nil {
		return time.Time{}, err
	}

	builds, err := cl.Builds.List(ctx, &sourcegraph.BuildListOptions{
		Repo:        res.Repo,
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

func resolveLocalRepo(ctx context.Context, repoPath string) (int32, error) {
	return handlerutil.GetRepoID(ctx, map[string]string{"Repo": repoPath})
}

func resolveLocalRepoRev(ctx context.Context, repoRev routevar.RepoRev) (*sourcegraph.RepoRevSpec, error) {
	cl, err := sourcegraph.NewClientFromContext(ctx)
	if err != nil {
		return nil, err
	}
	repo, err := resolveLocalRepo(ctx, repoRev.Repo)
	if err != nil {
		return nil, err
	}
	res, err := cl.Repos.ResolveRev(ctx, &sourcegraph.ReposResolveRevOp{Repo: repo, Rev: repoRev.Rev})
	if err != nil {
		return nil, err
	}
	return &sourcegraph.RepoRevSpec{
		Repo:     repo,
		CommitID: res.CommitID,
	}, nil
}

func resolveLocalRepos(ctx context.Context, repoPaths []string, ignoreErrors bool) ([]int32, error) {
	repoIDs := make([]int32, 0, len(repoPaths))
	for _, repoPath := range repoPaths {
		repoID, err := resolveLocalRepo(ctx, repoPath)
		if err != nil {
			if !ignoreErrors {
				return nil, err
			} else {
				log15.Warn("resolve local repo", "err", err, "repo", repoPath)
			}
		} else {
			repoIDs = append(repoIDs, repoID)
		}
	}
	return repoIDs, nil
}

// repoIDOrPath is a type used purely for documentation purposes to
// indicate that this URL query parameter can be either a string (repo
// path) or number (repo ID).
type repoIDOrPath string

func init() {
	schemaDecoder.RegisterConverter(repoIDOrPath(""), func(s string) reflect.Value { return reflect.ValueOf(s) })
}

// getRepoID gets the repo ID from an interface{} type that can be
// either an int32 or a string. Typically callers decode the URL query
// string's "Repo" or "repo" field into an interface{} value and then
// pass it to getRepoID. This way, they can accept either the numeric
// repo ID or the repo path, which presents a nicer API to consumers.
func getRepoID(ctx context.Context, v repoIDOrPath) (int32, error) {
	if n, err := strconv.Atoi(string(v)); err == nil {
		return int32(n), nil
	}
	return handlerutil.GetRepoID(ctx, map[string]string{"Repo": string(v)})
}
