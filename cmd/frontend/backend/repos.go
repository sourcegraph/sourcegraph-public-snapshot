package backend

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"go.opentelemetry.io/otel/attribute"

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
	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
	"github.com/sourcegraph/sourcegraph/internal/inventory"
	"github.com/sourcegraph/sourcegraph/internal/repoupdater"
	"github.com/sourcegraph/sourcegraph/internal/repoupdater/protocol"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type ReposService interface {
	Get(ctx context.Context, repo api.RepoID) (*types.Repo, error)
	GetByName(ctx context.Context, name api.RepoName) (*types.Repo, error)
	List(ctx context.Context, opt database.ReposListOptions) ([]*types.Repo, error)
	ListIndexable(ctx context.Context) ([]types.MinimalRepo, error)
	GetInventory(ctx context.Context, repo *types.Repo, commitID api.CommitID, forceEnhancedLanguageDetection bool) (*inventory.Inventory, error)
	DeleteRepositoryFromDisk(ctx context.Context, repoID api.RepoID) error
	RequestRepositoryClone(ctx context.Context, repoID api.RepoID) error
	ResolveRev(ctx context.Context, repo *types.Repo, rev string) (api.CommitID, error)
	GetCommit(ctx context.Context, repo *types.Repo, commitID api.CommitID) (*gitdomain.Commit, error)
}

// NewRepos uses the provided `database.DB` to initialize a new RepoService.
//
// NOTE: The underlying cache is reused from Repos global variable to actually
// make cache be useful. This is mostly a workaround for now until we come up a
// more idiomatic solution.
func NewRepos(logger log.Logger, db database.DB, client gitserver.Client) ReposService {
	repoStore := db.Repos()
	logger = logger.Scoped("repos", "provides a repos store for the backend")
	return &repos{
		logger:          logger,
		db:              db,
		gitserverClient: client,
		store:           repoStore,
		cache:           dbcache.NewIndexableReposLister(logger, repoStore),
	}
}

type repos struct {
	logger          log.Logger
	db              database.DB
	gitserverClient gitserver.Client
	store           database.RepoStore
	cache           *dbcache.IndexableReposLister
}

func (s *repos) Get(ctx context.Context, repo api.RepoID) (_ *types.Repo, err error) {
	if Mocks.Repos.Get != nil {
		return Mocks.Repos.Get(ctx, repo)
	}

	ctx, done := startTrace(ctx, "Get", repo, &err)
	defer done()

	return s.store.Get(ctx, repo)
}

// GetByName retrieves the repository with the given name. It will lazy sync a repo
// not yet present in the database under certain conditions. See repos.Syncer.SyncRepo.
func (s *repos) GetByName(ctx context.Context, name api.RepoName) (_ *types.Repo, err error) {
	if Mocks.Repos.GetByName != nil {
		return Mocks.Repos.GetByName(ctx, name)
	}

	ctx, done := startTrace(ctx, "GetByName", name, &err)
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

	newName, err := s.add(ctx, name)
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

// add adds the repository with the given name to the database by calling
// repo-updater when in sourcegraph.com mode. It's possible that the repo has
// been renamed on the code host in which case a different name may be returned.
func (s *repos) add(ctx context.Context, name api.RepoName) (addedName api.RepoName, err error) {
	ctx, done := startTrace(ctx, "Add", name, &err)
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
		if err := s.gitserverClient.IsRepoCloneable(ctx, name); err != nil {
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

	ctx, done := startTrace(ctx, "List", opt, &err)
	defer func() {
		if err == nil {
			if tr := trace.TraceFromContext(ctx); tr != nil {
				tr.SetAttributes(attribute.Int("result.len", len(repos)))
			}
		}
		done()
	}()

	return s.store.List(ctx, opt)
}

// ListIndexable calls database.ListMinimalRepos, with tracing. It lists ALL
// indexable repos. In addition, it only lists cloned repositories.
//
// The intended call site for this is the logic which assigns repositories to
// zoekt shards.
func (s *repos) ListIndexable(ctx context.Context) (repos []types.MinimalRepo, err error) {
	ctx, done := startTrace(ctx, "ListIndexable", nil, &err)
	defer func() {
		if err == nil {
			if tr := trace.TraceFromContext(ctx); tr != nil {
				tr.SetAttributes(attribute.Int("result.len", len(repos)))
			}
		}
		done()
	}()

	if envvar.SourcegraphDotComMode() {
		return s.cache.List(ctx)
	}

	return s.store.ListMinimalRepos(ctx, database.ReposListOptions{
		OnlyCloned: true,
	})
}

func (s *repos) GetInventory(ctx context.Context, repo *types.Repo, commitID api.CommitID, forceEnhancedLanguageDetection bool) (res *inventory.Inventory, err error) {
	if Mocks.Repos.GetInventory != nil {
		return Mocks.Repos.GetInventory(ctx, repo, commitID)
	}

	ctx, done := startTrace(ctx, "GetInventory", map[string]any{"repo": repo.Name, "commitID": commitID}, &err)
	defer done()

	// Cap GetInventory operation to some reasonable time.
	ctx, cancel := context.WithTimeout(ctx, 3*time.Minute)
	defer cancel()

	invCtx, err := InventoryContext(s.logger, repo.Name, s.gitserverClient, commitID, forceEnhancedLanguageDetection)
	if err != nil {
		return nil, err
	}

	root, err := s.gitserverClient.Stat(ctx, authz.DefaultSubRepoPermsChecker, repo.Name, commitID, "")
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

func (s *repos) DeleteRepositoryFromDisk(ctx context.Context, repoID api.RepoID) (err error) {
	if Mocks.Repos.DeleteRepositoryFromDisk != nil {
		return Mocks.Repos.DeleteRepositoryFromDisk(ctx, repoID)
	}

	repo, err := s.Get(ctx, repoID)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("error while fetching repo with ID %d", repoID))
	}

	ctx, done := startTrace(ctx, "DeleteRepositoryFromDisk", repoID, &err)
	defer done()

	err = s.gitserverClient.Remove(ctx, repo.Name)
	return err
}

func (s *repos) RequestRepositoryClone(ctx context.Context, repoID api.RepoID) (err error) {
	repo, err := s.Get(ctx, repoID)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("error while fetching repo with ID %d", repoID))
	}

	ctx, done := startTrace(ctx, "RequestRepositoryClone", repoID, &err)
	defer done()

	resp, err := s.gitserverClient.RequestRepoClone(ctx, repo.Name)
	if err != nil {
		return err
	}
	if resp.Error != "" {
		return errors.Newf("requesting clone for repo ID %d failed: %s", repoID, resp.Error)
	}

	return nil
}

// ResolveRev will return the absolute commit for a commit-ish spec in a repo.
// If no rev is specified, HEAD is used.
// Error cases:
// * Repo does not exist: gitdomain.RepoNotExistError
// * Commit does not exist: gitdomain.RevisionNotFoundError
// * Empty repository: gitdomain.RevisionNotFoundError
// * The user does not have permission: errcode.IsNotFound
// * Other unexpected errors.
func (s *repos) ResolveRev(ctx context.Context, repo *types.Repo, rev string) (commitID api.CommitID, err error) {
	if Mocks.Repos.ResolveRev != nil {
		return Mocks.Repos.ResolveRev(ctx, repo, rev)
	}

	ctx, done := startTrace(ctx, "ResolveRev", map[string]any{"repo": repo.Name, "rev": rev}, &err)
	defer done()

	return s.gitserverClient.ResolveRevision(ctx, repo.Name, rev, gitserver.ResolveRevisionOptions{})
}

func (s *repos) GetCommit(ctx context.Context, repo *types.Repo, commitID api.CommitID) (res *gitdomain.Commit, err error) {
	ctx, done := startTrace(ctx, "GetCommit", map[string]any{"repo": repo.Name, "commitID": commitID}, &err)
	defer done()

	s.logger.Debug("GetCommit", log.String("repo", string(repo.Name)), log.String("commitID", string(commitID)))

	if !gitserver.IsAbsoluteRevision(string(commitID)) {
		return nil, errors.Errorf("non-absolute CommitID for Repos.GetCommit: %v", commitID)
	}

	return s.gitserverClient.GetCommit(ctx, authz.DefaultSubRepoPermsChecker, repo.Name, commitID, gitserver.ResolveRevisionOptions{})
}

// ErrRepoSeeOther indicates that the repo does not exist on this server but might exist on an external Sourcegraph
// server.
type ErrRepoSeeOther struct {
	// RedirectURL is the base URL for the repository at an external location.
	RedirectURL string
}

func (e ErrRepoSeeOther) Error() string {
	return fmt.Sprintf("repo not found at this location, but might exist at %s", e.RedirectURL)
}
