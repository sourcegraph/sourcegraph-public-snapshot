package backend

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	log15 "gopkg.in/inconshreveable/log15.v2"

	opentracing "github.com/opentracing/opentracing-go"
	otlog "github.com/opentracing/opentracing-go/log"
	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/db"
	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/pkg/types"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/conf"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/env"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/externalservice/github"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/gitserver"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/inventory"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/rcache"
)

// ErrRepoSeeOther indicates that the repo does not exist on this server but might exist on an external Sourcegraph
// server.
type ErrRepoSeeOther struct {
	// RedirectURL is the base URL for the repository at an external location.
	RedirectURL string
}

func (e ErrRepoSeeOther) Error() string {
	return fmt.Sprintf("repo not found at this location, but might exist at %s", e.RedirectURL)
}

var Repos = &repos{}

type repos struct{}

func (s *repos) Get(ctx context.Context, repo api.RepoID) (_ *types.Repo, err error) {
	if Mocks.Repos.Get != nil {
		return Mocks.Repos.Get(ctx, repo)
	}

	ctx, done := trace(ctx, "Repos", "Get", repo, &err)
	defer done()

	return db.Repos.Get(ctx, repo)
}

// GetByURI retrieves the repository with the given URI. If the URI refers to a repository on a known external
// service (such as a code host) that is not yet present in the database, it will automatically look up the
// repository externally and add it to the database before returning it.
func (s *repos) GetByURI(ctx context.Context, uri api.RepoURI) (_ *types.Repo, err error) {
	if Mocks.Repos.GetByURI != nil {
		return Mocks.Repos.GetByURI(ctx, uri)
	}

	ctx, done := trace(ctx, "Repos", "GetByURI", uri, &err)
	defer done()

	repo, err := db.Repos.GetByURI(ctx, uri)
	if err != nil && conf.Get().AutoRepoAdd {
		if strings.HasPrefix(strings.ToLower(string(uri)), "github.com/") {
			if ghRepo, err := s.addFromGitHubAPI(ctx, uri); err == nil {
				return ghRepo, nil
			} else if err == context.DeadlineExceeded || err == context.Canceled {
				return nil, err
			}
		}

		if err := gitserver.DefaultClient.IsRepoCloneable(ctx, uri); err != nil {
			return nil, db.ErrRepoNotFound
		}
		if err := s.TryInsertNew(ctx, uri, "", false, true); err != nil {
			return nil, err
		}
		return db.Repos.GetByURI(ctx, uri)
	} else if err != nil {

		if !conf.Get().DisablePublicRepoRedirects && strings.HasPrefix(strings.ToLower(string(uri)), "github.com/") {
			return nil, ErrRepoSeeOther{RedirectURL: fmt.Sprintf("https://sourcegraph.com/%s", uri)}
		}
		return nil, err
	}

	return repo, nil
}

func (s *repos) addFromGitHubAPI(ctx context.Context, uri api.RepoURI) (*types.Repo, error) {
	// Repo does not exist in DB, create new entry.
	ctx = context.WithValue(ctx, github.GitHubTrackingContextKey, "Repos.GetByURI")
	ghRepo, err := github.GetRepo(ctx, uri)
	if err != nil {
		return nil, err
	}

	if actualURI := api.RepoURI("github.com/" + ghRepo.GetFullName()); actualURI != uri {
		// not canonical name (the GitHub api will redirect from the old name to
		// the results for the new name if the repo got renamed on GitHub)
		if repo, err := db.Repos.GetByURI(ctx, actualURI); err == nil {
			return repo, nil
		}
	}

	if err := s.TryInsertNew(ctx, uri, ghRepo.GetDescription(), ghRepo.GetFork(), true); err != nil {
		return nil, err
	}

	return db.Repos.GetByURI(ctx, uri)
}

func (s *repos) TryInsertNew(ctx context.Context, uri api.RepoURI, description string, fork, enabled bool) error {
	return db.Repos.TryInsertNew(ctx, uri, description, fork, enabled)
}

func (s *repos) List(ctx context.Context, opt db.ReposListOptions) (repos []*types.Repo, err error) {
	if Mocks.Repos.List != nil {
		return Mocks.Repos.List(ctx, opt)
	}

	ctx, done := trace(ctx, "Repos", "List", opt, &err)
	defer func() {
		if err == nil {
			span := opentracing.SpanFromContext(ctx)
			span.LogFields(otlog.Int("result.len", len(repos)))
		}
		done()
	}()

	ctx = context.WithValue(ctx, github.GitHubTrackingContextKey, "Repos.List")

	return db.Repos.List(ctx, opt)
}

var inventoryCache = rcache.New("inv")

func (s *repos) GetInventory(ctx context.Context, repo api.RepoID, commitID api.CommitID) (res *inventory.Inventory, err error) {
	if Mocks.Repos.GetInventory != nil {
		return Mocks.Repos.GetInventory(ctx, repo, commitID)
	}

	ctx, done := trace(ctx, "Repos", "GetInventory", map[string]interface{}{"repo": repo, "commitID": commitID}, &err)
	defer done()

	// Cap GetInventory operation to some reasonable time.
	ctx, cancel := context.WithTimeout(ctx, 3*time.Minute)
	defer cancel()

	if !isAbsCommitID(commitID) {
		return nil, errNotAbsCommitID
	}

	// Try cache first
	cacheKey := fmt.Sprintf("%d:%s", repo, commitID)
	if b, ok := inventoryCache.Get(cacheKey); ok {
		var inv inventory.Inventory
		if err := json.Unmarshal(b, &inv); err == nil {
			return &inv, nil
		}
		log15.Warn("Repos.GetInventory failed to unmarshal cached JSON inventory", "repo", repo, "commitID", commitID, "err", err)
	}

	// Not found in the cache, so compute it.
	inv, err := s.GetInventoryUncached(ctx, repo, commitID)
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

func (s *repos) GetInventoryUncached(ctx context.Context, repo api.RepoID, commitID api.CommitID) (*inventory.Inventory, error) {
	if Mocks.Repos.GetInventoryUncached != nil {
		return Mocks.Repos.GetInventoryUncached(ctx, repo, commitID)
	}

	vcsrepo, err := db.RepoVCS.Open(ctx, repo)
	if err != nil {
		return nil, err
	}

	files, err := vcsrepo.ReadDir(ctx, commitID, "", true)
	if err != nil {
		return nil, err
	}
	return inventory.Get(ctx, files)
}

var indexerAddr = env.Get("SRC_INDEXER", "indexer:3179", "The address of the indexer service.")

func (s *repos) RefreshIndex(ctx context.Context, repo api.RepoURI) (err error) {
	if Mocks.Repos.RefreshIndex != nil {
		return Mocks.Repos.RefreshIndex(ctx, repo)
	}

	go func() {
		resp, err := http.Get("http://" + indexerAddr + "/refresh?repo=" + string(repo))
		if err != nil {
			log15.Error("RefreshIndex failed", "error", err)
			return
		}
		resp.Body.Close()
	}()

	return nil
}
