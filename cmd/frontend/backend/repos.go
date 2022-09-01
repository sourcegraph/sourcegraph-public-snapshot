package backend

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/opentracing/opentracing-go"
	otlog "github.com/opentracing/opentracing-go/log"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/envvar"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/conf/deploy"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbcache"
	"github.com/sourcegraph/sourcegraph/internal/errcode"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/inventory"
	"github.com/sourcegraph/sourcegraph/internal/repoupdater"
	"github.com/sourcegraph/sourcegraph/internal/repoupdater/protocol"
	"github.com/sourcegraph/sourcegraph/internal/types"
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

// NewRepos uses the provided `database.RepoStore` to initialize a new repos
// store for the backend.
//
// NOTE: The underlying cache is reused from Repos global variable to actually
// make cache be useful. This is mostly a workaround for now until we come up a
// more idiomatic solution.
func NewRepos(logger log.Logger, db database.DB) *repos {
	repoStore := db.Repos()
	return &repos{
		db:    db,
		store: repoStore,
		cache: dbcache.NewIndexableReposLister(logger, repoStore),
	}
}

type repos struct {
	db    database.DB
	store database.RepoStore
	cache *dbcache.IndexableReposLister
}

func (s *repos) Get(ctx context.Context, repo api.RepoID) (_ *types.Repo, err error) {
	if Mocks.Repos.Get != nil {
		return Mocks.Repos.Get(ctx, repo)
	}

	ctx, done := trace(ctx, "Repos", "Get", repo, &err)
	defer done()

	return s.store.Get(ctx, repo)
}

// GetByName retrieves the repository with the given name. It will lazy sync a repo
// not yet present in the database under certain conditions. See repos.Syncer.SyncRepo.
func (s *repos) GetByName(ctx context.Context, name api.RepoName) (_ *types.Repo, err error) {
	if Mocks.Repos.GetByName != nil {
		return Mocks.Repos.GetByName(ctx, name)
	}

	ctx, done := trace(ctx, "Repos", "GetByName", name, &err)
	defer done()

	repo, err := s.store.GetByName(ctx, name)
	if err == nil {
		return repo, nil
	}

	if !errcode.IsNotFound(err) {
		return nil, err
	}

	if errcode.IsNotFound(err) && !envvar.SourcegraphDotComMode() {
		// The repo doesn't exist and we're not on sourcegraph.com, we should not lazy
		// clone it.
		return nil, err
	}

	newName, err := s.Add(ctx, name)
	if err == nil {
		return s.store.GetByName(ctx, newName)
	}

	if errcode.IsNotFound(err) && shouldRedirect(name) {
		return nil, ErrRepoSeeOther{RedirectURL: (&url.URL{
			Scheme:   "https",
			Host:     "sourcegraph.com",
			Path:     string(name),
			RawQuery: url.Values{"utm_source": []string{deploy.Type()}}.Encode(),
		}).String()}
	}

	return nil, err
}

func shouldRedirect(name api.RepoName) bool {
	return !conf.Get().DisablePublicRepoRedirects &&
		extsvc.CodeHostOf(name, extsvc.PublicCodeHosts...) != nil
}

var metricIsRepoCloneable = promauto.NewCounterVec(prometheus.CounterOpts{
	Name: "src_frontend_repo_add_is_cloneable",
	Help: "temporary metric to measure if this codepath is valuable on sourcegraph.com",
}, []string{"status"})

// Add adds the repository with the given name to the database by calling
// repo-updater when in sourcegraph.com mode. It's possible that the repo has
// been renamed on the code host in which case a different name may be returned.
func (s *repos) Add(ctx context.Context, name api.RepoName) (addedName api.RepoName, err error) {
	ctx, done := trace(ctx, "Repos", "Add", name, &err)
	defer done()

	// Avoid hitting repo-updater (and incurring a hit against our GitHub/etc. API rate
	// limit) for repositories that don't exist or private repositories that people attempt to
	// access.
	codehost := extsvc.CodeHostOf(name, extsvc.PublicCodeHosts...)
	if codehost == nil {
		return "", &database.RepoNotFoundErr{Name: name}
	}

	status := "unknown"
	defer func() {
		metricIsRepoCloneable.WithLabelValues(status).Inc()
	}()

	if !codehost.IsPackageHost() {
		if err := gitserver.NewClient(s.db).IsRepoCloneable(ctx, name); err != nil {
			if ctx.Err() != nil {
				status = "timeout"
			} else {
				status = "fail"
			}
			return "", err
		}
	}

	status = "success"

	// Looking up the repo in repo-updater makes it sync that repo to the
	// database on sourcegraph.com if that repo is from github.com or gitlab.com
	lookupResult, err := repoupdater.DefaultClient.RepoLookup(ctx, protocol.RepoLookupArgs{Repo: name})
	if lookupResult != nil && lookupResult.Repo != nil {
		return lookupResult.Repo.Name, err
	}
	return "", err
}

func (s *repos) List(ctx context.Context, opt database.ReposListOptions) (repos []*types.Repo, err error) {
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

	return s.store.List(ctx, opt)
}

// ListIndexable calls database.ListMinimalRepos, with tracing. It lists
// ALL indexable repos which could include private user added repos.
// In addition, it only lists cloned repositories.
//
// The intended call site for this is the logic which assigns repositories to
// zoekt shards.
func (s *repos) ListIndexable(ctx context.Context) (repos []types.MinimalRepo, err error) {
	ctx, done := trace(ctx, "Repos", "ListIndexable", nil, &err)
	defer func() {
		if err == nil {
			span := opentracing.SpanFromContext(ctx)
			span.LogFields(otlog.Int("result.len", len(repos)))
		}
		done()
	}()

	if envvar.SourcegraphDotComMode() {
		return s.cache.List(ctx)
	}

	trueP := true
	return s.store.ListMinimalRepos(ctx, database.ReposListOptions{
		Index:      &trueP,
		OnlyCloned: true,
	})
}

func (s *repos) GetInventory(ctx context.Context, repo *types.Repo, commitID api.CommitID, forceEnhancedLanguageDetection bool) (res *inventory.Inventory, err error) {
	if Mocks.Repos.GetInventory != nil {
		return Mocks.Repos.GetInventory(ctx, repo, commitID)
	}

	ctx, done := trace(ctx, "Repos", "GetInventory", map[string]any{"repo": repo.Name, "commitID": commitID}, &err)
	defer done()

	// Cap GetInventory operation to some reasonable time.
	ctx, cancel := context.WithTimeout(ctx, 3*time.Minute)
	defer cancel()

	invCtx, err := InventoryContext(repo.Name, s.db, commitID, forceEnhancedLanguageDetection)
	if err != nil {
		return nil, err
	}

	root, err := gitserver.NewClient(s.db).Stat(ctx, authz.DefaultSubRepoPermsChecker, repo.Name, commitID, "")
	if err != nil {
		return nil, err
	}

	// In computing the inventory, sub-tree inventories are cached based on the OID of the Git
	// tree. Compared to per-blob caching, this creates many fewer cache entries, which means fewer
	// stores, fewer lookups, and less cache storage overhead. Compared to per-commit caching, this
	// yields a higher cache hit rate because most trees are unchanged across commits.
	inv, err := invCtx.Entries(ctx, root)
	if err != nil {
		return nil, err
	}
	return &inv, nil
}
