package summary

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/autoindexing/internal/inference"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/autoindexing/internal/jobselector"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/autoindexing/internal/store"
	uploadsshared "github.com/sourcegraph/sourcegraph/internal/codeintel/uploads/shared"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/lib/errors"

	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

// For mocking in tests
var autoIndexingEnabled = conf.CodeIntelAutoIndexingEnabled

func NewSummaryBuilder(
	observationCtx *observation.Context,
	store store.Store,
	jobSelector *jobselector.JobSelector,
	uploadSvc UploadService,
	config *Config,
) goroutine.BackgroundRoutine {
	return goroutine.NewPeriodicGoroutine(
		// We should use an internal actor when doing cross service calls.
		actor.WithInternalActor(context.Background()),
		goroutine.HandlerFunc(func(ctx context.Context) error {
			repositoryWithCounts, err := store.TopRepositoriesToConfigure(ctx, config.NumRepositoriesToConfigure)
			if err != nil {
				return err
			}

			for _, repositoryWithCount := range repositoryWithCounts {
				recentUploads, err := uploadSvc.GetRecentUploadsSummary(ctx, repositoryWithCount.RepositoryID)
				if err != nil {
					return err
				}
				recentIndexes, err := uploadSvc.GetRecentIndexesSummary(ctx, repositoryWithCount.RepositoryID)
				if err != nil {
					return err
				}

				inferredAvailableIndexers := map[string]uploadsshared.AvailableIndexer{}

				if autoIndexingEnabled() {
					commit := "HEAD"
					result, err := jobSelector.InferIndexJobsFromRepositoryStructure(ctx, repositoryWithCount.RepositoryID, commit, "", false)
					if err != nil {
						if errors.As(err, &inference.LimitError{}) {
							continue
						}

						return err
					}

					// Create blocklist for indexes that have already been uploaded.
					blocklist := map[string]struct{}{}
					for _, u := range recentUploads {
						key := uploadsshared.GetKeyForLookup(u.Indexer, u.Root)
						blocklist[key] = struct{}{}
					}
					for _, u := range recentIndexes {
						key := uploadsshared.GetKeyForLookup(u.Indexer, u.Root)
						blocklist[key] = struct{}{}
					}

					inferredAvailableIndexers = uploadsshared.PopulateInferredAvailableIndexers(result.IndexJobs, blocklist, inferredAvailableIndexers)
					// inferredAvailableIndexers = uploadsshared.PopulateInferredAvailableIndexers(indexJobHints, blocklist, inferredAvailableIndexers)
				}

				if err := store.SetConfigurationSummary(ctx, repositoryWithCount.RepositoryID, repositoryWithCount.Count, inferredAvailableIndexers); err != nil {
					return err
				}
			}

			if err := store.TruncateConfigurationSummary(ctx, config.NumRepositoriesToConfigure); err != nil {
				return err
			}

			return nil
		}),
		goroutine.WithName("codeintel.autoindexing-summary-builder"),
		goroutine.WithDescription("build an auto-indexing summary over repositories with high search activity"),
		goroutine.WithInterval(config.Interval),
	)
}
