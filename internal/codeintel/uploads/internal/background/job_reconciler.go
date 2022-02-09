package background

import (
	"context"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/goroutine"
)

func (b backgroundJob) NewReconciler(
	interval time.Duration,
	batchSize int,
) goroutine.BackgroundRoutine {
	return goroutine.NewPeriodicGoroutine(context.Background(), interval, goroutine.HandlerFunc(func(ctx context.Context) error {
		return b.handleReconcile(ctx, batchSize)
	}))
}

func (b backgroundJob) handleReconcile(ctx context.Context, batchSize int) (err error) {
	ids, err := b.uploadSvc.ReconcileCandidates(ctx, batchSize)
	if err != nil {
		return err
	}

	b.operations.numReconcileScans.Add(float64(len(ids)))

	dumps, err := b.uploadSvc.GetDumpsByIDs(ctx, ids)
	if err != nil {
		return err
	}

	found := map[int]struct{}{}
	for _, dump := range dumps {
		found[dump.ID] = struct{}{}
	}

	abandoned := ids[:0]
	for _, id := range ids {
		if _, ok := found[id]; ok {
			continue
		}

		abandoned = append(abandoned, id)
	}

	if err := b.uploadSvc.DeleteLsifDataByUploadIds(ctx, abandoned...); err != nil {
		return err
	}

	b.operations.numReconcileDeletes.Add(float64(len(abandoned)))
	return nil
}
