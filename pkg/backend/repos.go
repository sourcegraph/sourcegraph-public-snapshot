package backend

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	log15 "gopkg.in/inconshreveable/log15.v2"

	gogithub "github.com/sourcegraph/go-github/github"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/accesscontrol"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/actor"
	sourcegraph "sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/api/legacyerr"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/env"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/github"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/inventory"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/localstore"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/rcache"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/vcs"
)

var Repos = &repos{}

type repos struct{}

// e2eUserPrefix is prefixed to all e2etest user account logins to
// ensure they can be filtered out of different systems easily and do
// not conflict with real user accounts.
const e2eUserPrefix = "e2etestuserx4FF3"

func (s *repos) Get(ctx context.Context, repoSpec *sourcegraph.RepoSpec) (res *sourcegraph.Repo, err error) {
	if Mocks.Repos.Get != nil {
		return Mocks.Repos.Get(ctx, repoSpec)
	}

	ctx, done := trace(ctx, "Repos", "Get", repoSpec, &err)
	defer done()

	repo, err := localstore.Repos.Get(ctx, repoSpec.ID)
	if err != nil {
		return nil, err
	}

	// SECURITY: calling setRepoFieldsFromRemote ensures we keep repository metadata up to date
	// (most importantly the "Private" field) and also adds redundancy to our security. However, we
	// don't call it if there are no GitHub creds. Do not remove this setRepoFieldsFromRemote call
	// without first checking with Richard and Beyang.
	if !github.PreferRawGit {
		if err := s.setRepoFieldsFromRemote(ctx, repo); err != nil {
			return nil, err
		}
	}

	if repo.Blocked {
		return nil, legacyerr.Errorf(legacyerr.FailedPrecondition, "repo %s is blocked", repo.URI)
	}

	return repo, nil
}

func (s *repos) GetByURI(ctx context.Context, uri string) (res *sourcegraph.Repo, err error) {
	if Mocks.Repos.GetByURI != nil {
		return Mocks.Repos.GetByURI(ctx, uri)
	}

	ctx, done := trace(ctx, "Repos", "GetByURI", uri, &err)
	defer done()

	repo, err := localstore.Repos.GetByURI(ctx, uri)
	if err != nil {
		return nil, err
	}

	if repo.Blocked {
		return nil, legacyerr.Errorf(legacyerr.FailedPrecondition, "repo %s is blocked", repo.URI)
	}

	return repo, nil
}

// ghRepoQueryMatcher matches search queries that look like they refer
// to GitHub repositories. Examples include "github.com/gorilla/mux", "gorilla/mux", "gorilla mux",
// "gorilla / mux"
var ghRepoQueryMatcher = regexp.MustCompile(`^(?:github.com/)?([^/\s]+)[/\s]+([^/\s]+)$`)

func (s *repos) List(ctx context.Context, opt *sourcegraph.RepoListOptions) (res *sourcegraph.RepoList, err error) {
	if Mocks.Repos.List != nil {
		return Mocks.Repos.List(ctx, opt)
	}

	ctx, done := trace(ctx, "Repos", "List", opt, &err)
	defer done()

	ctx = context.WithValue(ctx, github.GitHubTrackingContextKey, "Repos.List")
	if opt == nil {
		opt = &sourcegraph.RepoListOptions{}
	}

	if opt.RemoteOnly {
		// List all of the repos we can access that are associated with the current
		// user. This includes that user's public repos and all the repos accessible
		// via their installations.
		repos, err := github.ListAccessibleRepos(ctx)
		if err != nil {
			return nil, err
		}

		// If we didn't list any repos above (e.g. if they don't have the
		// GitHub app installed), then list their public repos. Otherwise we
		// would end up with their public repos being in the list twice.
		if len(repos) == 0 {
			public, err := github.ListPublicReposForUser(ctx, actor.FromContext(ctx).Login)
			if err != nil {
				return nil, err
			}
			repos = append(repos, public...)
		}
		return &sourcegraph.RepoList{Repos: repos}, nil
	}

	repos, err := localstore.Repos.List(ctx, &localstore.RepoListOp{
		Query:       opt.Query,
		ListOptions: opt.ListOptions,
	})
	if err != nil {
		return nil, err
	}

	// Augment with external results if user is authenticated,
	// RemoteSearch is true, and Query is non-empty.
	if opt.RemoteSearch {
		if !actor.FromContext(ctx).IsAuthenticated() {
			// GitHub repo search API calls are subject to a strict
			// rate limit shared by all unauthenticated users. We
			// would quickly exceed it if we allowed this.
			return nil, errors.New("refusing to perform remote search for unauthenticated user")
		}

		ghquery := opt.Query
		if matches := ghRepoQueryMatcher.FindStringSubmatch(opt.Query); matches != nil {
			// Apply query transformation to make GitHub results better.
			ghquery = fmt.Sprintf("user:%s in:name %s", matches[1], matches[2])
		}

		var ghrepos []*sourcegraph.Repo
		var err error
		if ghquery == "" {
			ghrepos, err = github.ListAllGitHubRepos(ctx, &gogithub.RepositoryListOptions{})
			ghrepos, repos = repos, ghrepos
		} else {
			ghrepos, err = github.SearchRepo(ctx, ghquery, nil)
		}
		if err == nil {
			existingRepos := make(map[string]struct{}, len(repos))
			for _, repo := range repos {
				existingRepos[repo.URI] = struct{}{}
			}
			for _, ghrepo := range ghrepos {
				if _, in := existingRepos[ghrepo.URI]; !in {
					repos = append(repos, ghrepo)
				}
			}
		} else {
			// Fetching results from GitHub is best-effort, as we
			// might hit the rate limit and don't want this to kill
			// the search experience entirely.
			log15.Warn("Unable to fetch repo search results from GitHub", "query", opt.Query, "error", err)
		}
	}

	return &sourcegraph.RepoList{Repos: repos}, nil
}

// setRepoFieldsFromRemote sets the fields of the repository from the
// remote (e.g., GitHub) and updates the repository in the store layer.
func (s *repos) setRepoFieldsFromRemote(ctx context.Context, repo *sourcegraph.Repo) error {
	if strings.HasPrefix(strings.ToLower(repo.URI), "github.com/") {
		// Fetch latest metadata from GitHub
		ghrepo, err := github.GetRepo(ctx, repo.URI)
		if err != nil {
			return err
		}
		if update := repoSetFromRemote(repo, ghrepo); update != nil {
			log15.Debug("Updating repo metadata from remote", "repo", repo.URI)
			// setRepoFieldsFromRemote is used in read requests, including
			// unauthed ones. However, this write isn't as the user, but
			// rather an optimization for us to save the data from
			// github. As such we use an elevated context to allow the
			// write.
			if err := localstore.Repos.Update(accesscontrol.WithInsecureSkip(ctx, true), *update); err != nil {
				log15.Error("Failed updating repo metadata from remote", "repo", repo.URI, "error", err)
			}
		}
	}
	return nil
}

func timestampEqual(a, b *time.Time) bool {
	if a == b {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return a.Equal(*b)
}

// repoSetFromRemote updates repo with fields from ghrepo that are
// different. If any fields are changed a non-nil store.RepoUpdate is returned
// representing the update.
func repoSetFromRemote(repo *sourcegraph.Repo, ghrepo *sourcegraph.Repo) *localstore.RepoUpdate {
	updated := false
	updateOp := &localstore.RepoUpdate{
		ReposUpdateOp: &sourcegraph.ReposUpdateOp{
			Repo: repo.ID,
		},
	}

	if ghrepo.URI != repo.URI {
		repo.URI = ghrepo.URI
		updateOp.URI = ghrepo.URI
		updated = true
	}
	if ghrepo.Description != repo.Description {
		repo.Description = ghrepo.Description
		updateOp.Description = ghrepo.Description
		updated = true
	}
	if ghrepo.HomepageURL != repo.HomepageURL {
		repo.HomepageURL = ghrepo.HomepageURL
		updateOp.HomepageURL = ghrepo.HomepageURL
		updated = true
	}
	if ghrepo.DefaultBranch != repo.DefaultBranch {
		repo.DefaultBranch = ghrepo.DefaultBranch
		updateOp.DefaultBranch = ghrepo.DefaultBranch
		updated = true
	}
	if ghrepo.Language != repo.Language {
		repo.Language = ghrepo.Language
		updateOp.Language = ghrepo.Language
		updated = true
	}
	if ghrepo.Blocked != repo.Blocked {
		repo.Blocked = ghrepo.Blocked
		if ghrepo.Blocked {
			updateOp.Blocked = sourcegraph.ReposUpdateOp_TRUE
		} else {
			updateOp.Blocked = sourcegraph.ReposUpdateOp_FALSE
		}
		updated = true
	}
	if ghrepo.Fork != repo.Fork {
		repo.Fork = ghrepo.Fork
		if ghrepo.Fork {
			updateOp.Fork = sourcegraph.ReposUpdateOp_TRUE
		} else {
			updateOp.Fork = sourcegraph.ReposUpdateOp_FALSE
		}
		updated = true
	}
	if ghrepo.Private != repo.Private {
		repo.Private = ghrepo.Private
		if ghrepo.Private {
			updateOp.Private = sourcegraph.ReposUpdateOp_TRUE
		} else {
			updateOp.Private = sourcegraph.ReposUpdateOp_FALSE
		}
		updated = true
	}

	if !timestampEqual(repo.UpdatedAt, ghrepo.UpdatedAt) {
		repo.UpdatedAt = ghrepo.UpdatedAt
		updateOp.UpdatedAt = ghrepo.UpdatedAt
		updated = true
	}
	if !timestampEqual(repo.PushedAt, ghrepo.PushedAt) {
		repo.PushedAt = ghrepo.PushedAt
		updateOp.PushedAt = ghrepo.PushedAt
		updated = true
	}

	if updated {
		return updateOp
	}
	return nil
}

func (s *repos) Update(ctx context.Context, op *sourcegraph.ReposUpdateOp) (err error) {
	if Mocks.Repos.Update != nil {
		return Mocks.Repos.Update(ctx, op)
	}

	ctx, done := trace(ctx, "Repos", "Update", op, &err)
	defer done()

	ts := time.Now()
	update := localstore.RepoUpdate{ReposUpdateOp: op, UpdatedAt: &ts}
	if err := localstore.Repos.Update(ctx, update); err != nil {
		return err
	}

	return nil
}

var inventoryCache = rcache.New("inv")

func (s *repos) GetInventory(ctx context.Context, repoRev *sourcegraph.RepoRevSpec) (res *inventory.Inventory, err error) {
	if Mocks.Repos.GetInventory != nil {
		return Mocks.Repos.GetInventory(ctx, repoRev)
	}

	ctx, done := trace(ctx, "Repos", "GetInventory", repoRev, &err)
	defer done()

	// Cap GetInventory operation to some reasonable time.
	ctx, cancel := context.WithTimeout(ctx, 3*time.Minute)
	defer cancel()

	if !isAbsCommitID(repoRev.CommitID) {
		return nil, errNotAbsCommitID
	}

	// Try cache first
	cacheKey := fmt.Sprintf("%d:%s", repoRev.Repo, repoRev.CommitID)
	if b, ok := inventoryCache.Get(cacheKey); ok {
		var inv inventory.Inventory
		if err := json.Unmarshal(b, &inv); err == nil {
			return &inv, nil
		}
		log15.Warn("Repos.GetInventory failed to unmarshal cached JSON inventory", "repoRev", repoRev, "err", err)
	}

	// Not found in the cache, so compute it.
	inv, err := s.GetInventoryUncached(ctx, repoRev)
	if err != nil {
		return nil, err
	}

	// Store inventory in cache.
	b, err := json.Marshal(inv)
	if err != nil {
		return nil, err
	}
	inventoryCache.Set(cacheKey, b)

	return inv, nil
}

func (s *repos) GetInventoryUncached(ctx context.Context, repoRev *sourcegraph.RepoRevSpec) (*inventory.Inventory, error) {
	if Mocks.Repos.GetInventoryUncached != nil {
		return Mocks.Repos.GetInventoryUncached(ctx, repoRev)
	}

	vcsrepo, err := localstore.RepoVCS.Open(ctx, repoRev.Repo)
	if err != nil {
		return nil, err
	}

	files, err := vcsrepo.ReadDir(ctx, vcs.CommitID(repoRev.CommitID), "", true)
	if err != nil {
		return nil, err
	}
	return inventory.Get(ctx, files)
}

var indexerAddr = env.Get("SRC_INDEXER", "indexer:3179", "The address of the indexer service.")

func (s *repos) RefreshIndex(ctx context.Context, repo string) (err error) {
	if Mocks.Repos.RefreshIndex != nil {
		return Mocks.Repos.RefreshIndex(ctx, repo)
	}

	go func() {
		resp, err := http.Get("http://" + indexerAddr + "/refresh?repo=" + repo)
		if err != nil {
			log15.Error("RefreshIndex failed", "error", err)
			return
		}
		resp.Body.Close()
	}()

	return nil
}
