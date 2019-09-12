package backend

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/opentracing/opentracing-go"
	otlog "github.com/opentracing/opentracing-go/log"
	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/envvar"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/inventory"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/pkg/conf"
	"github.com/sourcegraph/sourcegraph/pkg/gitserver"
	"github.com/sourcegraph/sourcegraph/pkg/rcache"
	"github.com/sourcegraph/sourcegraph/pkg/repoupdater"
	"github.com/sourcegraph/sourcegraph/pkg/repoupdater/protocol"
	"github.com/sourcegraph/sourcegraph/pkg/vcs/git"
	"gopkg.in/inconshreveable/log15.v2"
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
	if err != nil && envvar.SourcegraphDotComMode() {
		// Automatically add repositories on Sourcegraph.com.
		if err := s.AddGitHubDotComRepository(ctx, name); err != nil {
			return nil, err
		}
		return db.Repos.GetByName(ctx, name)
	} else if err != nil {
		if !conf.Get().DisablePublicRepoRedirects && strings.HasPrefix(strings.ToLower(string(name)), "github.com/") {
			return nil, ErrRepoSeeOther{RedirectURL: (&url.URL{
				Scheme:   "https",
				Host:     "sourcegraph.com",
				Path:     string(name),
				RawQuery: url.Values{"utm_source": []string{conf.DeployType()}}.Encode(),
			}).String()}
		}
		return nil, err
	}

	return repo, nil
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

// ListDefault calls db.DefaultRepos.List, with tracing.
func (s *repos) ListDefault(ctx context.Context) (repos []*types.Repo, err error) {
	ctx, done := trace(ctx, "Repos", "ListDefault", nil, &err)
	defer func() {
		if err == nil {
			span := opentracing.SpanFromContext(ctx)
			span.LogFields(otlog.Int("result.len", len(repos)))
		}
		done()
	}()
	return db.DefaultRepos.List(ctx)
}

var inventoryCache = rcache.New("inv")

// Feature flag for enhanced (but much slower) language detection that uses file contents, not just
// filenames.
var useEnhancedLanguageDetection, _ = strconv.ParseBool(os.Getenv("USE_ENHANCED_LANGUAGE_DETECTION"))

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

	cachedRepo, err := CachedGitRepo(ctx, repo)
	if err != nil {
		return nil, err
	}
	root, err := git.Stat(ctx, *cachedRepo, commitID, "")
	if err != nil {
		return nil, err
	}

	cacheKey := func(tree os.FileInfo) string {
		// Cache based on the OID of the Git tree. Compared to per-blob caching, this creates many
		// fewer cache entries, which means fewer stores, fewer lookups, and less cache storage
		// overhead. Compared to per-commit caching, this yields a higher cache hit rate because
		// most trees are unchanged across commits.
		return tree.Sys().(git.ObjectInfo).OID().String()
	}
	invCtx := inventory.Context{
		ReadTree: func(ctx context.Context, path string) ([]os.FileInfo, error) {
			// TODO: As a perf optimization, we could read multiple levels of the Git tree at once
			// to avoid sequential tree traversal calls.
			return git.ReadDir(ctx, *cachedRepo, commitID, path, false)
		},
		ReadFile: func(ctx context.Context, path string, minBytes int64) ([]byte, error) {
			return git.ReadFile(ctx, *cachedRepo, commitID, path, minBytes)
		},
		CacheGet: func(tree os.FileInfo) (inventory.Inventory, bool) {
			if b, ok := inventoryCache.Get(cacheKey(tree)); ok {
				var inv inventory.Inventory
				if err := json.Unmarshal(b, &inv); err != nil {
					log15.Warn("Repos.GetInventory failed to unmarshal cached JSON inventory", "repo", repo.Name, "commitID", commitID, "tree", tree.Name(), "err", err)
					return inventory.Inventory{}, false
				}
				return inv, true
			}
			return inventory.Inventory{}, false
		},
		CacheSet: func(tree os.FileInfo, inv inventory.Inventory) {
			b, err := json.Marshal(&inv)
			if err != nil {
				log15.Warn("Repos.GetInventory failed to marshal JSON inventory for cache", "repo", repo.Name, "commitID", commitID, "tree", tree.Name(), "err", err)
				return
			}
			inventoryCache.Set(cacheKey(tree), b)
		},
	}

	if !useEnhancedLanguageDetection {
		// If USE_ENHANCED_LANGUAGE_DETECTION is disabled, do not read file contents to determine
		// the language.
		invCtx.ReadFile = func(ctx context.Context, path string, minBytes int64) ([]byte, error) {
			return nil, nil
		}
	}

	inv, err := invCtx.Tree(ctx, root)
	if err != nil {
		return nil, err
	}
	return &inv, nil
}
