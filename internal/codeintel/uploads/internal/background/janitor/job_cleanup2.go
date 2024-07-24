package janitor

import (
	"context"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/shared/background"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/uploads/internal/store"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

func NewUnknownRepositoryJanitor(
	store store.Store,
	config *Config,
	observationCtx *observation.Context,
) goroutine.BackgroundRoutine {
	name := "codeintel.autoindexing.janitor.unknown-repository"

	return background.NewJanitorJob(context.Background(), background.JanitorOptions{
		Name:        name,
		Description: "Removes index records associated with an unknown repository.",
		Interval:    config.Interval,
		Metrics:     background.NewJanitorMetrics(observationCtx, name),
		CleanupFunc: func(ctx context.Context) (numRecordsScanned, numRecordsAltered int, _ error) {
			return store.DeleteAutoIndexJobsWithoutRepository(ctx, time.Now())
		},
	})
}

//
//

func NewUnknownCommitJanitor2(
	store store.Store,
	gitserverClient gitserver.Client,
	config *Config,
	observationCtx *observation.Context,
) goroutine.BackgroundRoutine {
	name := "codeintel.autoindexing.janitor.unknown-commit"

	return background.NewJanitorJob(context.Background(), background.JanitorOptions{
		Name:        name,
		Description: "Removes index records associated with an unknown commit.",
		Interval:    config.Interval,
		Metrics:     background.NewJanitorMetrics(observationCtx, name),
		CleanupFunc: func(ctx context.Context) (numRecordsScanned, numRecordsAltered int, _ error) {
			return store.ProcessStaleSourcedCommits(
				ctx,
				config.MinimumTimeSinceLastCheck,
				config.CommitResolverBatchSize,
				config.CommitResolverMaximumCommitLag,
				func(ctx context.Context, repositoryID int, repositoryName, commit string) (bool, error) {
					return shouldDeleteRecordsForCommit(ctx, gitserverClient, repositoryName, commit)
				},
			)
		},
	})
}

//
//

func NewExpiredRecordJanitor(
	store store.Store,
	config *Config,
	observationCtx *observation.Context,
) goroutine.BackgroundRoutine {
	name := "codeintel.autoindexing.janitor.expired"

	return background.NewJanitorJob(context.Background(), background.JanitorOptions{
		Name:        name,
		Description: "Removes old index records",
		Interval:    config.Interval,
		Metrics:     background.NewJanitorMetrics(observationCtx, name),
		CleanupFunc: func(ctx context.Context) (numRecordsScanned, numRecordsAltered int, _ error) {
			return store.ExpireFailedRecords(ctx, config.FailedIndexBatchSize, config.FailedIndexMaxAge, time.Now())
		},
	})
}
