package search

import (
	"context"
	"fmt"

	"github.com/inconshreveable/log15"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/search/streaming"
	streamapi "github.com/sourcegraph/sourcegraph/internal/search/streaming/api"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func getEventRepoMetadata(ctx context.Context, db database.DB, event streaming.SearchEvent) (map[api.RepoID]*types.SearchedRepo, error) {
	ids := repoIDs(event.Results)
	if len(ids) == 0 {
		// Return early if there are no repos in the event
		return nil, nil
	}

	metadataList, err := db.Repos().Metadata(ctx, ids...)
	if err != nil {
		return nil, errors.Wrap(err, "fetch metadata from db")
	}

	repoMetadata := make(map[api.RepoID]*types.SearchedRepo, len(ids))
	for _, repo := range metadataList {
		repoMetadata[repo.ID] = repo
	}
	return repoMetadata, nil
}

// repoNamer returns a best-effort function which translates repository IDs
// into names.
func repoNamer(ctx context.Context, db database.DB) streamapi.RepoNamer {
	cache := map[api.RepoID]api.RepoName{}

	return func(ids []api.RepoID) []api.RepoName {
		// Strategy is to populate from cache. So we first populate the cache
		// with IDs not already in the cache.
		var missing []api.RepoID
		for _, id := range ids {
			if _, ok := cache[id]; !ok {
				missing = append(missing, id)
			}
		}

		if len(missing) > 0 {
			err := db.Repos().StreamMinimalRepos(ctx, database.ReposListOptions{
				IDs: missing,
			}, func(repo *types.MinimalRepo) {
				cache[repo.ID] = repo.Name
			})
			if err != nil {
				// repoNamer is best-effort, so we just log the error.
				log15.Warn("streaming search repoNamer failed to list names", "error", err)
			}
		}

		names := make([]api.RepoName, 0, len(ids))
		for _, id := range ids {
			if name, ok := cache[id]; ok {
				names = append(names, name)
			} else {
				names = append(names, api.RepoName(fmt.Sprintf("UNKNOWN{ID=%d}", id)))
			}
		}

		return names
	}
}
