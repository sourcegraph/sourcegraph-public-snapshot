package background

import (
	"context"
	"time"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/autoindexing/internal/store"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/shared/background"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

const recordTypeName = "autoindexing"

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
		Metrics:     background.NewJanitorMetrics(observationCtx, name, recordTypeName),
		CleanupFunc: func(ctx context.Context) (numRecordsScanned, numRecordsAltered int, _ error) {
			return store.DeleteIndexesWithoutRepository(ctx, time.Now())
		},
	})
}

//
//

func NewUnknownCommitJanitor(
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
		Metrics:     background.NewJanitorMetrics(observationCtx, name, recordTypeName),
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

func shouldDeleteRecordsForCommit(ctx context.Context, gitserverClient gitserver.Client, repositoryName, commit string) (bool, error) {
	if _, err := gitserverClient.ResolveRevision(ctx, api.RepoName(repositoryName), commit, gitserver.ResolveRevisionOptions{}); err != nil {
		if gitdomain.IsRepoNotExist(err) {
			// Repository not found; we'll delete these in a separate process
			return false, nil
		}

		if errors.HasType(err, &gitdomain.RevisionNotFoundError{}) {
			// Repository is resolvable but commit is not - remove it
			return true, nil
		}

		// Unexpected error
		return false, err
	}

	// Commit is resolvable, don't touch it
	return false, nil
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
		Metrics:     background.NewJanitorMetrics(observationCtx, name, recordTypeName),
		CleanupFunc: func(ctx context.Context) (numRecordsScanned, numRecordsAltered int, _ error) {
			return store.ExpireFailedRecords(ctx, batchSize, maxAge, time.Now())
		},
	})
}
