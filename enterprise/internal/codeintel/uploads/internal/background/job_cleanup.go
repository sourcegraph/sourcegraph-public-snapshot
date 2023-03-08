package background

import (
	"context"
	"sort"
	"time"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/shared/background"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/shared/types"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/uploads/internal/lsifstore"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/uploads/internal/store"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/uploads/shared"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

const recordTypeName = "SCIP metadata"

func NewDeletedRepositoryJanitor(
	store store.Store,
	interval time.Duration,
	observationCtx *observation.Context,
) goroutine.BackgroundRoutine {
	name := "codeintel.uploads.janitor.unknown-repository"

	return background.NewJanitorJob(context.Background(), background.JanitorOptions{
		Name:        name,
		Description: "Removes upload records associated with an unknown repository.",
		Interval:    interval,
		Metrics:     background.NewJanitorMetrics(observationCtx, name, recordTypeName),
		CleanupFunc: func(ctx context.Context) (numRecordsScanned, numRecordsAltered int, _ error) {
			return store.DeleteUploadsWithoutRepository(ctx, time.Now())
		},
	})
}

//
//

func NewUnknownCommitJanitor(
	store store.Store,
	gitserverClient GitserverClient,
	interval time.Duration,
	commitResolverBatchSize int,
	minimumTimeSinceLastCheck time.Duration,
	commitResolverMaximumCommitLag time.Duration,
	observationCtx *observation.Context,
) goroutine.BackgroundRoutine {
	name := "codeintel.uploads.janitor.unknown-commit"

	return background.NewJanitorJob(context.Background(), background.JanitorOptions{
		Name:        name,
		Description: "Removes upload records associated with an unknown commit.",
		Interval:    interval,
		Metrics:     background.NewJanitorMetrics(observationCtx, name, recordTypeName),
		CleanupFunc: func(ctx context.Context) (numRecordsScanned, numRecordsAltered int, _ error) {
			return store.ProcessSourcedCommits(
				ctx,
				minimumTimeSinceLastCheck,
				commitResolverMaximumCommitLag,
				commitResolverBatchSize,
				func(ctx context.Context, repositoryID int, commit string) (bool, error) {
					return shouldDeleteRecordsForCommit(ctx, gitserverClient, repositoryID, commit)
				},
				time.Now(),
			)
		},
	})
}

func shouldDeleteRecordsForCommit(ctx context.Context, gitserverClient GitserverClient, repositoryID int, commit string) (bool, error) {
	if _, err := gitserverClient.ResolveRevision(ctx, repositoryID, commit); err != nil {
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

func NewAbandonedUploadJanitor(
	store store.Store,
	interval time.Duration,
	uploadTimeout time.Duration,
	observationCtx *observation.Context,
) goroutine.BackgroundRoutine {
	name := "codeintel.uploads.janitor.abandoned"

	return background.NewJanitorJob(context.Background(), background.JanitorOptions{
		Name:        name,
		Description: "Removes upload records that did did not receive a full payload from the user.",
		Interval:    interval,
		Metrics:     background.NewJanitorMetrics(observationCtx, name, recordTypeName),
		CleanupFunc: func(ctx context.Context) (numRecordsScanned, numRecordsAltered int, _ error) {
			return store.DeleteUploadsStuckUploading(ctx, time.Now().UTC().Add(-uploadTimeout))
		},
	})
}

//
//

const (
	expiredUploadsBatchSize    = 1000
	expiredUploadsMaxTraversal = 100
)

func NewExpiredUploadJanitor(
	store store.Store,
	interval time.Duration,
	observationCtx *observation.Context,
) goroutine.BackgroundRoutine {
	name := "codeintel.uploads.expirer.unreferenced"

	return background.NewJanitorJob(context.Background(), background.JanitorOptions{
		Name:        name,
		Description: "Soft-deletes unreferenced upload records that are not protected by any data retention policy.",
		Interval:    interval,
		Metrics:     background.NewJanitorMetrics(observationCtx, name, recordTypeName),
		CleanupFunc: func(ctx context.Context) (numRecordsScanned, numRecordsAltered int, _ error) {
			return store.SoftDeleteExpiredUploads(ctx, expiredUploadsBatchSize)
		},
	})
}

func NewExpiredUploadTraversalJanitor(
	store store.Store,
	interval time.Duration,
	observationCtx *observation.Context,
) goroutine.BackgroundRoutine {
	name := "codeintel.uploads.expirer.unreferenced-graph"

	return background.NewJanitorJob(context.Background(), background.JanitorOptions{
		Name:        name,
		Description: "Soft-deletes a tree of externally unreferenced upload records that are not protected by any data retention policy.",
		Interval:    interval,
		Metrics:     background.NewJanitorMetrics(observationCtx, name, recordTypeName),
		CleanupFunc: func(ctx context.Context) (numRecordsScanned, numRecordsAltered int, _ error) {
			return store.SoftDeleteExpiredUploadsViaTraversal(ctx, expiredUploadsMaxTraversal)
		},
	})
}

//
//

func NewHardDeleter(
	store store.Store,
	lsifStore lsifstore.LsifStore,
	interval time.Duration,
	observationCtx *observation.Context,
) goroutine.BackgroundRoutine {
	name := "codeintel.uploads.hard-deleter"

	return background.NewJanitorJob(context.Background(), background.JanitorOptions{
		Name:        name,
		Description: "Deleted data associated with soft-deleted upload records.",
		Interval:    interval,
		Metrics:     background.NewJanitorMetrics(observationCtx, name, recordTypeName),
		CleanupFunc: func(ctx context.Context) (numRecordsScanned, numRecordsAltered int, _ error) {
			const uploadsBatchSize = 100
			options := shared.GetUploadsOptions{
				State:            "deleted",
				Limit:            uploadsBatchSize,
				AllowExpired:     true,
				AllowDeletedRepo: true,
			}

			count := 0
			for {
				// Always request the first page of deleted uploads. If this is not
				// the first iteration of the loop, then the previous iteration has
				// deleted the records that composed the previous page, and the
				// previous "second" page is now the first page.
				uploads, totalCount, err := store.GetUploads(ctx, options)
				if err != nil {
					return 0, 0, err
				}

				ids := uploadIDs(uploads)
				if err := lsifStore.DeleteLsifDataByUploadIds(ctx, ids...); err != nil {
					return 0, 0, err
				}

				if err := store.HardDeleteUploadsByIDs(ctx, ids...); err != nil {
					return 0, 0, err
				}

				count += len(uploads)
				if count >= totalCount {
					break
				}
			}

			return count, count, nil
		},
	})
}

func uploadIDs(uploads []types.Upload) []int {
	ids := make([]int, 0, len(uploads))
	for i := range uploads {
		ids = append(ids, uploads[i].ID)
	}
	sort.Ints(ids)

	return ids
}

//
//

func NewAuditLogJanitor(
	store store.Store,
	interval time.Duration,
	auditLogMaxAge time.Duration,
	observationCtx *observation.Context,
) goroutine.BackgroundRoutine {
	name := "codeintel.uploads.janitor.audit-logs"

	return background.NewJanitorJob(context.Background(), background.JanitorOptions{
		Name:        name,
		Description: "Deletes sufficiently old upload audit log records.",
		Interval:    interval,
		Metrics:     background.NewJanitorMetrics(observationCtx, name, recordTypeName),
		CleanupFunc: func(ctx context.Context) (numRecordsScanned, numRecordsAltered int, _ error) {
			return store.DeleteOldAuditLogs(ctx, auditLogMaxAge, time.Now())
		},
	})
}

//
//

func NewSCIPExpirationTask(
	lsifStore lsifstore.LsifStore,
	interval time.Duration,
	unreferencedDocumentBatchSize int,
	unreferencedDocumentMaxAge time.Duration,
	observationCtx *observation.Context,
) goroutine.BackgroundRoutine {
	name := "codeintel.uploads.janitor.scip-documents"

	return background.NewJanitorJob(context.Background(), background.JanitorOptions{
		Name:        name,
		Description: "Deletes SCIP document payloads that are not referenced by any index.",
		Interval:    interval,
		Metrics:     background.NewJanitorMetrics(observationCtx, name, recordTypeName),
		CleanupFunc: func(ctx context.Context) (numRecordsScanned, numRecordsAltered int, _ error) {
			return lsifStore.DeleteUnreferencedDocuments(ctx, unreferencedDocumentBatchSize, unreferencedDocumentMaxAge, time.Now())
		},
	})
}
