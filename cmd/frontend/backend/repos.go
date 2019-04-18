package backend

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"time"

	opentracing "github.com/opentracing/opentracing-go"
	otlog "github.com/opentracing/opentracing-go/log"
	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/envvar"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/globals"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/inventory"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/pkg/conf"
	"github.com/sourcegraph/sourcegraph/pkg/gitserver"
	"github.com/sourcegraph/sourcegraph/pkg/rcache"
	"github.com/sourcegraph/sourcegraph/pkg/repoupdater"
	"github.com/sourcegraph/sourcegraph/pkg/repoupdater/protocol"
	"github.com/sourcegraph/sourcegraph/pkg/vcs/git"
	log15 "gopkg.in/inconshreveable/log15.v2"
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

// GetByName retrieves the repository with the given name. If the name refers to a repository on a known external
// service (such as a code host) that is not yet present in the database, it will automatically look up the
// repository externally and add it to the database before returning it.
func (s *repos) GetByName(ctx context.Context, name api.RepoName) (_ *types.Repo, err error) {
	if Mocks.Repos.GetByName != nil {
		return Mocks.Repos.GetByName(ctx, name)
	}

	ctx, done := trace(ctx, "Repos", "GetByName", name, &err)
	defer done()

	repo, err := db.Repos.GetByName(ctx, name)
	if err == nil {
		return repo, nil
	}

	if envvar.SourcegraphDotComMode() {
		// Automatically add repositories on Sourcegraph.com.
		if err := s.AddGitHubDotComRepository(ctx, name); err != nil {
			return nil, err
		}
		return db.Repos.GetByName(ctx, name)
	}

	if !conf.Get().DisablePublicRepoRedirects && strings.HasPrefix(strings.ToLower(string(name)), "github.com/") {
		return nil, ErrRepoSeeOther{RedirectURL: (&url.URL{
			Scheme:   "https",
			Host:     "sourcegraph.com",
			Path:     string(name),
			RawQuery: url.Values{"utm_source": []string{conf.DeployType()}}.Encode(),
		}).String()}
	}

	if repo, _ := db.Repos.GetByURI(ctx, string(name)); repo != nil {
		u := globals.ExternalURL.ResolveReference(&url.URL{Path: string(repo.Name)})
		return nil, ErrRepoSeeOther{RedirectURL: u.String()}
	}

	return nil, err
}

// AddGitHubDotComRepository adds the repository with the given name. The name is mapped to a repository by consulting the
// repo-updater, which contains information about all configured code hosts and the names that they
// handle.
func (s *repos) AddGitHubDotComRepository(ctx context.Context, name api.RepoName) (err error) {
	if Mocks.Repos.AddGitHubDotComRepository != nil {
		return Mocks.Repos.AddGitHubDotComRepository(name)
	}

	ctx, done := trace(ctx, "Repos", "AddGitHubDotComRepository", name, &err)
	defer done()

	// Avoid hitting the repoupdater (and incurring a hit against our GitHub/etc. API rate
	// limit) for repositories that don't exist or private repositories that people attempt to
	// access.
	gitserverRepo, err := quickGitserverRepo(ctx, name, "github.com")
	if err != nil {
		return err
	}
	if gitserverRepo != nil {
		if err := gitserver.DefaultClient.IsRepoCloneable(ctx, *gitserverRepo); err != nil {
			return err
		}
	}

	// Try to look up and add the repo.
	result, err := repoupdater.DefaultClient.RepoLookup(ctx, protocol.RepoLookupArgs{Repo: name})
	if err != nil {
		return err
	}
	if result.Repo != nil {
		// Allow anonymous users on Sourcegraph.com to enable repositories just by visiting them, but
		// everywhere else, require server admins to explicitly enable repositories.
		enableAutoAddedRepos := envvar.SourcegraphDotComMode()
		if err := s.Upsert(ctx, api.InsertRepoOp{
			Name:         result.Repo.Name,
			Description:  result.Repo.Description,
			Fork:         result.Repo.Fork,
			Archived:     result.Repo.Archived,
			Enabled:      enableAutoAddedRepos,
			ExternalRepo: result.Repo.ExternalRepo,
		}); err != nil {
			return err
		}
	}
	return nil
}

func (s *repos) Upsert(ctx context.Context, op api.InsertRepoOp) error {
	return db.Repos.Upsert(ctx, op)
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

	return db.Repos.List(ctx, opt)
}

var inventoryCache = rcache.New("inv")

func (s *repos) GetInventory(ctx context.Context, repo *types.Repo, commitID api.CommitID) (res *inventory.Inventory, err error) {
	if Mocks.Repos.GetInventory != nil {
		return Mocks.Repos.GetInventory(ctx, repo, commitID)
	}

	ctx, done := trace(ctx, "Repos", "GetInventory", map[string]interface{}{"repo": repo.Name, "commitID": commitID}, &err)
	defer done()

	// Cap GetInventory operation to some reasonable time.
	ctx, cancel := context.WithTimeout(ctx, 3*time.Minute)
	defer cancel()

	if !git.IsAbsoluteRevision(string(commitID)) {
		return nil, errors.Errorf("non-absolute CommitID for Repos.GetInventory: %v", commitID)
	}

	// Try cache first
	cacheKey := fmt.Sprintf("%s:%s", repo.Name, commitID)
	if b, ok := inventoryCache.Get(cacheKey); ok {
		var inv inventory.Inventory
		if err := json.Unmarshal(b, &inv); err == nil {
			return &inv, nil
		}
		log15.Warn("Repos.GetInventory failed to unmarshal cached JSON inventory", "repo", repo.Name, "commitID", commitID, "err", err)
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

func (s *repos) GetInventoryUncached(ctx context.Context, repo *types.Repo, commitID api.CommitID) (res *inventory.Inventory, err error) {
	if Mocks.Repos.GetInventoryUncached != nil {
		return Mocks.Repos.GetInventoryUncached(ctx, repo, commitID)
	}

	ctx, done := trace(ctx, "Repos", "GetInventoryUncached", map[string]interface{}{"repo": repo.Name, "commitID": commitID}, &err)
	defer done()

	cachedRepo, err := CachedGitRepo(ctx, repo)
	if err != nil {
		return nil, err
	}

	files, err := git.ReadDir(ctx, *cachedRepo, commitID, "", true)
	if err != nil {
		return nil, err
	}
	return inventory.Get(ctx, files)
}
