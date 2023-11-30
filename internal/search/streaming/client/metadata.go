package client

import (
	"context"
	"fmt"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database"
	streamapi "github.com/sourcegraph/sourcegraph/internal/search/streaming/api"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

// RepoNamer returns a best-effort function which translates repository IDs into names.
func RepoNamer(ctx context.Context, db database.DB) streamapi.RepoNamer {
	logger := log.Scoped("RepoNamer")
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
				// RepoNamer is best-effort, so we just log the error.
				logger.Warn("streaming search RepoNamer failed to list names", log.Error(err))
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
