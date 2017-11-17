package backend

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"time"

	log15 "gopkg.in/inconshreveable/log15.v2"

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

func (s *repos) TryInsertNew(ctx context.Context, uri string, description string, fork bool, private bool) error {
	return localstore.Repos.TryInsertNew(ctx, uri, description, fork, private)
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

	repos, err := localstore.Repos.List(ctx, &localstore.RepoListOp{
		Query:           opt.Query,
		IncludePatterns: opt.IncludePatterns,
		ExcludePattern:  opt.ExcludePattern,
		ListOptions:     opt.ListOptions,
	})
	if err != nil {
		return nil, err
	}
	return &sourcegraph.RepoList{Repos: repos}, nil
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
