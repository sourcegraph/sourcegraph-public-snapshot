package search

import (
	"context"

	"github.com/inconshreveable/log15"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/search/streaming"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

func getEventRepoMetadata(ctx context.Context, db dbutil.DB, event streaming.SearchEvent) map[api.RepoID]types.RepoMetadata {
	ids := repoIDs(event.Results)
	if len(ids) == 0 {
		// Return early if there are no repos in the event
		return nil
	}

	metadataList, err := database.Repos(db).Metadata(ctx, ids...)
	if err != nil {
		log15.Error("streaming: failed to retrieve repo metadata", "error", err)
		return nil
	}

	repoMetadata := make(map[api.RepoID]types.RepoMetadata, len(ids))
	for _, repo := range metadataList {
		repoMetadata[repo.ID] = repo
	}
	return repoMetadata
}
