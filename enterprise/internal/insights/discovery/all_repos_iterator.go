package discovery

import (
	"context"
	"time"

	"github.com/pkg/errors"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/envvar"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

// AllReposIterator implements an efficient way to iterate over every single repository on
// Sourcegraph that should be considered for code insights.
//
// It caches multiple consecutive uses in order to ensure repository lists (which can be quite
// large, e.g. 500,000+ repositories) are only fetched as frequently as needed.
type AllReposIterator struct {
	DefaultRepoStore *database.DefaultRepoStore
	RepoStore        *database.RepoStore

	// RepositoryListCacheTime describes how long to cache repository lists for. These API calls
	// can result in hundreds of thousands of repositories, so choose wisely.
	RepositoryListCacheTime time.Duration

	// Internal fields below.
	cachedRepoNamesAge time.Time
	cachedRepoNames    []string
	cachedPageRequests map[database.LimitOffset]cachedPageRequest
}

// ForEach invokes the given function for every repository that we should consider gathering data
// for historically.
//
// This takes into account paginating repository names from the database (as there could be e.g.
// 500,000+ of them). It also takes into account Sourcegraph.com, where we only gather historical
// data for the same subset of repos we index for search.
//
// If the forEach function returns an error, pagination is stopped and the error returned.
func (a *AllReposIterator) ForEach(ctx context.Context, forEach func(repoName string) error) error {
	if envvar.SourcegraphDotComMode() {
		// Use cached results if we have them and it is reasonable to do so.
		if time.Since(a.cachedRepoNamesAge) < a.RepositoryListCacheTime {
			for _, repo := range a.cachedRepoNames {
				if err := forEach(repo); err != nil {
					return err
				}
			}
			return nil
		}
		a.cachedRepoNames = nil

		// We shouldn't try to fill historical data for ALL repos on Sourcegraph.com, it would take
		// forever. Instead, we use the same list of default repositories used when you do a global
		// search on Sourcegraph.com.
		res, err := a.DefaultRepoStore.List(ctx)
		if err != nil {
			return errors.Wrap(err, "DefaultRepoStore.List")
		}
		names := make([]string, len(res))
		for i, r := range res {
			names[i] = string(r.Name)
		}
		a.cachedRepoNamesAge = time.Now()
		a.cachedRepoNames = names
		for _, repo := range a.cachedRepoNames {
			if err := forEach(repo); err != nil {
				return errors.Wrap(err, "forEach")
			}
		}
		return nil
	}

	// Regular deployments of Sourcegraph.
	//
	// We paginate 1,000 repositories out of the DB at a time.
	limitOffset := database.LimitOffset{
		Limit:  1000,
		Offset: 0,
	}
	for {
		// Get the next page.
		repos, err := a.cachedRepoStoreList(ctx, limitOffset)
		if err != nil {
			return errors.Wrap(err, "RepoStore.List")
		}
		if len(repos) == 0 {
			return nil // done!
		}

		// Call the forEach function on every repository.
		for _, r := range repos {
			if err := forEach(string(r.Name)); err != nil {
				return errors.Wrap(err, "forEach")
			}
		}

		// Set outselves up to get the next page.
		limitOffset.Offset += len(repos)
	}
}

// cachedRepoStoreList calls a.repoStore.List to do a paginated list of repositories, and caches the
// results in-memory for some time.
//
// This is primarily useful because we call this function e.g. 1 time per 365 days.
func (a *AllReposIterator) cachedRepoStoreList(ctx context.Context, page database.LimitOffset) ([]*types.Repo, error) {
	if a.cachedPageRequests == nil {
		a.cachedPageRequests = map[database.LimitOffset]cachedPageRequest{}
	}
	cacheEntry, ok := a.cachedPageRequests[page]
	if ok && time.Since(cacheEntry.age) < a.RepositoryListCacheTime {
		return cacheEntry.results, nil
	}

	repos, err := a.RepoStore.List(ctx, database.ReposListOptions{
		// No point in trying to search uncloned repositories.
		OnlyCloned: true,

		// Order by repository name.
		OrderBy: database.RepoListOrderBy{{Field: database.RepoListName}},

		LimitOffset: &page,
	})
	if err != nil {
		return nil, err
	}
	a.cachedPageRequests[page] = cachedPageRequest{
		age:     time.Now(),
		results: repos,
	}
	return repos, nil
}

type cachedPageRequest struct {
	age     time.Time
	results []*types.Repo
}
