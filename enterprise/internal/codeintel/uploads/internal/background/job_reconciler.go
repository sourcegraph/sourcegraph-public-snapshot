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
	if err := b.handleReconcileFromFrontend(ctx, batchSize); err != nil {
		return err
	}

	if err := b.handleReconcileFromCodeintelDB(ctx, batchSize); err != nil {
		return err
	}

	return nil
}

// handleReconcileFromFrontend marks upload records that has no resolvable data in the codeintel-db.
func (b backgroundJob) handleReconcileFromFrontend(ctx context.Context, batchSize int) (err error) {
	ids, err := b.uploadSvc.FrontendReconcileCandidates(ctx, batchSize)
	if err != nil {
		return err
	}

	b.operations.numReconcileScansFromFrontend.Add(float64(len(ids)))

	idsWithMeta, err := b.uploadSvc.IDsWithMeta(ctx, ids)
	if err != nil {
		return err
	}

	found := map[int]struct{}{}
	for _, id := range idsWithMeta {
		found[id] = struct{}{}
	}

	abandoned := ids[:0]
	for _, id := range ids {
		if _, ok := found[id]; ok {
			continue
		}

		abandoned = append(abandoned, id)
	}

	// In the future we'll also want to explicitly mark these uploads as missing precise data so that
	// they can be re-indexed or removed by an automatic janitor process. For now we just want to know
	// *IF* this condition happens, so a Prometheus metric is sufficient.
	b.operations.numReconcileDeletesFromFrontend.Add(float64(len(abandoned)))
	return nil
}

// handleReconcileFromCodeintelDB removes data from the codeintel-db that has no correlated upload
// in the frontend database.
func (b backgroundJob) handleReconcileFromCodeintelDB(ctx context.Context, batchSize int) (err error) {
	ids, err := b.uploadSvc.CodeIntelDBReconcileCandidates(ctx, batchSize)
	if err != nil {
		return err
	}

	b.operations.numReconcileScansFromCodeIntelDB.Add(float64(len(ids)))

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

	b.operations.numReconcileDeletesFromCodeIntelDB.Add(float64(len(abandoned)))
	return nil
}
