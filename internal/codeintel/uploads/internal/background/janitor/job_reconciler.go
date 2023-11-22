package janitor

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/shared/background"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/uploads/internal/lsifstore"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/uploads/internal/store"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

func NewFrontendDBReconciler(
	store store.Store,
	lsifstore lsifstore.Store,
	config *Config,
	observationCtx *observation.Context,
) goroutine.BackgroundRoutine {
	return newReconciler(
		"codeintel.uploads.reconciler.scip-metadata",
		"Counts SCIP metadata records for which there is no data in the codeintel-db schema.",
		&storeWrapper{store},
		&lsifStoreWrapper{lsifstore},
		config,
		observationCtx,
	)
}

func NewCodeIntelDBReconciler(
	store store.Store,
	lsifstore lsifstore.Store,
	config *Config,
	observationCtx *observation.Context,
) goroutine.BackgroundRoutine {
	return newReconciler(
		"codeintel.uploads.reconciler.scip-data",
		"Removes SCIP data records for which there is no known associated metadata in the frontend schema.",
		&lsifStoreWrapper{lsifstore},
		&storeWrapper{store},
		config,
		observationCtx,
	)
}

//
//

type sourceStore interface {
	Candidates(ctx context.Context, batchSize int) ([]int, error)
	Prune(ctx context.Context, ids []int) error
}

type reconcileStore interface {
	FilterExists(ctx context.Context, ids []int) ([]int, error)
}

func newReconciler(
	name string,
	description string,
	sourceStore sourceStore,
	reconcileStore reconcileStore,
	config *Config,
	observationCtx *observation.Context,
) goroutine.BackgroundRoutine {
	return background.NewJanitorJob(context.Background(), background.JanitorOptions{
		Name:        name,
		Description: description,
		Interval:    config.Interval,
		Metrics:     background.NewJanitorMetrics(observationCtx, name),
		CleanupFunc: func(ctx context.Context) (numRecordsScanned, numRecordsAltered int, _ error) {
			candidateIDs, err := sourceStore.Candidates(ctx, config.ReconcilerBatchSize)
			if err != nil {
				return 0, 0, err
			}

			existingIDs, err := reconcileStore.FilterExists(ctx, candidateIDs)
			if err != nil {
				return 0, 0, err
			}

			found := map[int]struct{}{}
			for _, id := range existingIDs {
				found[id] = struct{}{}
			}

			missingIDs := candidateIDs[:0]
			for _, id := range candidateIDs {
				if _, ok := found[id]; ok {
					continue
				}

				missingIDs = append(missingIDs, id)
			}

			if err := sourceStore.Prune(ctx, missingIDs); err != nil {
				return 0, 0, err
			}

			return len(candidateIDs), len(missingIDs), nil
		},
	})
}

//
//

type storeWrapper struct {
	store store.Store
}

func (s *storeWrapper) Candidates(ctx context.Context, batchSize int) ([]int, error) {
	return s.store.ReconcileCandidates(ctx, batchSize)
}

func (s *storeWrapper) Prune(ctx context.Context, ids []int) error {
	// In the future we'll also want to explicitly mark these uploads as missing precise data so that
	// they can be re-indexed or removed by an automatic janitor process. For now we just want to know
	// *IF* this condition happens, so a Prometheus metric is sufficient.
	return nil
}

func (s *storeWrapper) FilterExists(ctx context.Context, candidateIDs []int) ([]int, error) {
	uploads, err := s.store.GetUploadsByIDsAllowDeleted(ctx, candidateIDs...)
	if err != nil {
		return nil, err
	}

	ids := make([]int, 0, len(uploads))
	for _, upload := range uploads {
		ids = append(ids, upload.ID)
	}

	return ids, nil
}

type lsifStoreWrapper struct {
	lsifstore lsifstore.Store
}

func (s *lsifStoreWrapper) Candidates(ctx context.Context, batchSize int) ([]int, error) {
	return s.lsifstore.ReconcileCandidates(ctx, batchSize)
}

func (s *lsifStoreWrapper) Prune(ctx context.Context, ids []int) error {
	return s.lsifstore.DeleteLsifDataByUploadIds(ctx, ids...)
}

func (s *lsifStoreWrapper) FilterExists(ctx context.Context, candidateIDs []int) ([]int, error) {
	return s.lsifstore.IDsWithMeta(ctx, candidateIDs)
}
