package background

import (
	"context"
	"time"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/shared/background"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/uploads/internal/store"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

const recordTypeName2 = "autoindexing"

func NewUnknownRepositoryJanitor(
	store store.Store,
	interval time.Duration,
	observationCtx *observation.Context,
) goroutine.BackgroundRoutine {
	name := "codeintel.autoindexing.janitor.unknown-repository"

	return background.NewJanitorJob(context.Background(), background.JanitorOptions{
		Name:        name,
		Description: "Removes index records associated with an unknown repository.",
		Interval:    interval,
		Metrics:     background.NewJanitorMetrics(observationCtx, name, recordTypeName2),
		CleanupFunc: func(ctx context.Context) (numRecordsScanned, numRecordsAltered int, _ error) {
			return store.DeleteIndexesWithoutRepository(ctx, time.Now())
		},
	})
}

//
//

func NewUnknownCommitJanitor2(
	store store.Store,
	gitserverClient gitserver.Client,
	interval time.Duration,
	commitResolverBatchSize int,
	minimumTimeSinceLastCheck time.Duration,
	commitResolverMaximumCommitLag time.Duration,
	observationCtx *observation.Context,
) goroutine.BackgroundRoutine {
	name := "codeintel.autoindexing.janitor.unknown-commit"

	return background.NewJanitorJob(context.Background(), background.JanitorOptions{
		Name:        name,
		Description: "Removes index records associated with an unknown commit.",
		Interval:    interval,
		Metrics:     background.NewJanitorMetrics(observationCtx, name, recordTypeName2),
		CleanupFunc: func(ctx context.Context) (numRecordsScanned, numRecordsAltered int, _ error) {
			return store.ProcessStaleSourcedCommits(
				ctx,
				minimumTimeSinceLastCheck,
				commitResolverBatchSize,
				commitResolverMaximumCommitLag,
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
	interval time.Duration,
	batchSize int,
	maxAge time.Duration,
	observationCtx *observation.Context,
) goroutine.BackgroundRoutine {
	name := "codeintel.autoindexing.janitor.expired"

	return background.NewJanitorJob(context.Background(), background.JanitorOptions{
		Name:        name,
		Description: "Removes old index records",
		Interval:    interval,
		Metrics:     background.NewJanitorMetrics(observationCtx, name, recordTypeName2),
		CleanupFunc: func(ctx context.Context) (numRecordsScanned, numRecordsAltered int, _ error) {
			return store.ExpireFailedRecords(ctx, batchSize, maxAge, time.Now())
		},
	})
}
