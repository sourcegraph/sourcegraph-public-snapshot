package background

import (
	"context"
	"time"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/autoindexing/internal/inference"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/autoindexing/internal/jobselector"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/autoindexing/internal/store"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/autoindexing/shared"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/lib/errors"

	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

func NewSummaryBuilder(
	observationCtx *observation.Context,
	store store.Store,
	jobSelector *jobselector.JobSelector,
	uploadSvc UploadService,
	interval time.Duration,
	numRepositoriesToConfigure int,
) goroutine.BackgroundRoutine {
	return goroutine.NewPeriodicGoroutine(
		// We should use an internal actor when doing cross service calls.
		actor.WithInternalActor(context.Background()),
		"codeintel.autoindexing-summary-builder", "build an auto-indexing summary over repositories with high search activity",
		interval,
		goroutine.HandlerFunc(func(ctx context.Context) error {
			repositoryWithCounts, err := store.TopRepositoriesToConfigure(ctx, numRepositoriesToConfigure)
			if err != nil {
				return err
			}

			for _, repositoryWithCount := range repositoryWithCounts {
				recentUploads, err := uploadSvc.GetRecentUploadsSummary(ctx, repositoryWithCount.RepositoryID)
				if err != nil {
					return err
				}
				recentIndexes, err := store.GetRecentIndexesSummary(ctx, repositoryWithCount.RepositoryID)
				if err != nil {
					return err
				}

				// Create blocklist for indexes that have already been uploaded.
				blocklist := map[string]struct{}{}
				for _, u := range recentUploads {
					key := shared.GetKeyForLookup(u.Indexer, u.Root)
					blocklist[key] = struct{}{}
				}
				for _, u := range recentIndexes {
					key := shared.GetKeyForLookup(u.Indexer, u.Root)
					blocklist[key] = struct{}{}
				}

				commit := "HEAD"
				indexJobs, err := jobSelector.InferIndexJobsFromRepositoryStructure(ctx, repositoryWithCount.RepositoryID, commit, "", false)
				if err != nil {
					if errors.As(err, &inference.LimitError{}) {
						continue
					}

					return err
				}
				// indexJobHints, err := jobSelector.InferIndexJobHintsFromRepositoryStructure(ctx, repositoryWithCount.RepositoryID, commit)
				// if err != nil {
				// 	if errors.As(err, &inference.LimitError{}) {
				// 		continue
				// 	}

				// 	return err
				// }

				inferredAvailableIndexers := map[string]shared.AvailableIndexer{}
				inferredAvailableIndexers = shared.PopulateInferredAvailableIndexers(indexJobs, blocklist, inferredAvailableIndexers)
				// inferredAvailableIndexers = shared.PopulateInferredAvailableIndexers(indexJobHints, blocklist, inferredAvailableIndexers)

				if err := store.SetConfigurationSummary(ctx, repositoryWithCount.RepositoryID, repositoryWithCount.Count, inferredAvailableIndexers); err != nil {
					return err
				}
			}

			if err := store.TruncateConfigurationSummary(ctx, numRepositoriesToConfigure); err != nil {
				return err
			}

			return nil
		}),
	)
}
