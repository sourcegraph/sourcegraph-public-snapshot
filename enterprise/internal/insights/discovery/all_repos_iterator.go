package discovery

import (
	"context"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type RepoIterator interface {
	ForEach(ctx context.Context, each func(repoName string, id api.RepoID) error) error
}

// IndexableReposLister is a subset of the API exposed by the backend.ListIndexable.
type IndexableReposLister interface {
	List(ctx context.Context) ([]types.MinimalRepo, error)
}

// RepoStore is a subset of the API exposed by the database.Repos() store.
type RepoStore interface {
	List(ctx context.Context, opt database.ReposListOptions) (results []*types.Repo, err error)
}

// AllReposIterator implements an efficient way to iterate over every single repository on
// Sourcegraph that should be considered for code insights.
//
// It caches multiple consecutive uses in order to ensure repository lists (which can be quite
// large, e.g. 500,000+ repositories) are only fetched as frequently as needed.
type AllReposIterator struct {
	IndexableReposLister  IndexableReposLister
	RepoStore             RepoStore
	Clock                 func() time.Time
	SourcegraphDotComMode bool // result of envvar.SourcegraphDotComMode()

	// RepositoryListCacheTime describes how long to cache repository lists for. These API calls
	// can result in hundreds of thousands of repositories, so choose wisely as it can be expensive
	// to pull such large numbers of rows from the DB frequently.
	RepositoryListCacheTime time.Duration

	counter *prometheus.CounterVec

	// Internal fields below.
	cachedPageRequests map[database.LimitOffset]cachedPageRequest
}

func NewAllReposIterator(indexableReposLister IndexableReposLister, repoStore RepoStore, clock func() time.Time, sourcegraphDotComMode bool, repositoryListCacheTime time.Duration, counterOpts *prometheus.CounterOpts) *AllReposIterator {
	return &AllReposIterator{IndexableReposLister: indexableReposLister, RepoStore: repoStore, Clock: clock, SourcegraphDotComMode: sourcegraphDotComMode, RepositoryListCacheTime: repositoryListCacheTime, counter: promauto.NewCounterVec(*counterOpts, []string{"result"})}
}

func (a *AllReposIterator) timeSince(t time.Time) time.Duration {
	return a.Clock().Sub(t)
}

// ForEach invokes the given function for every repository that we should consider gathering data
// for historically.
//
// This takes into account paginating repository names from the database (as there could be e.g.
// 500,000+ of them). It also takes into account Sourcegraph.com, where we only gather historical
// data for the same subset of repos we index for search.
//
// If the forEach function returns an error, pagination is stopped and the error returned.
func (a *AllReposIterator) ForEach(ctx context.Context, forEach func(repoName string, id api.RepoID) error) error {
	// 🚨 SECURITY: this context will ensure that this iterator goes over all repositories
	globalCtx := actor.WithInternalActor(ctx)

	// Regular deployments of Sourcegraph.
	//
	// We paginate 1,000 repositories out of the DB at a time.
	limitOffset := database.LimitOffset{
		Limit:  1000,
		Offset: 0,
	}
	for {
		// Get the next page.
		repos, err := a.cachedRepoStoreList(globalCtx, limitOffset)
		if err != nil {
			return errors.Wrap(err, "RepoStore.List")
		}
		if len(repos) == 0 {
			return nil // done!
		}

		// Call the forEach function on every repository.
		for _, r := range repos {
			if err := forEach(string(r.Name), r.ID); err != nil {
				a.counter.WithLabelValues("error").Inc()
				return errors.Wrap(err, "forEach")
			}
			a.counter.WithLabelValues("success").Inc()

		}

		// Set outselves up to get the next page.
		limitOffset.Offset += len(repos)
	}
}

// cachedRepoStoreList calls a.repoStore.List to do a paginated list of repositories, and caches the
// results in-memory for some time.
func (a *AllReposIterator) cachedRepoStoreList(ctx context.Context, page database.LimitOffset) ([]*types.Repo, error) {
	if a.cachedPageRequests == nil {
		a.cachedPageRequests = map[database.LimitOffset]cachedPageRequest{}
	}
	cacheEntry, ok := a.cachedPageRequests[page]
	if ok && a.timeSince(cacheEntry.age) < a.RepositoryListCacheTime {
		return cacheEntry.results, nil
	}

	trueP := true
	repos, err := a.RepoStore.List(ctx, database.ReposListOptions{
		Index: &trueP,

		LimitOffset: &page,
	})
	if err != nil {
		return nil, err
	}
	a.cachedPageRequests[page] = cachedPageRequest{
		age:     a.Clock(),
		results: repos,
	}
	return repos, nil
}

type cachedPageRequest struct {
	age     time.Time
	results []*types.Repo
}
