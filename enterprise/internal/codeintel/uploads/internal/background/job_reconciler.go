package background

import (
	"context"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type reconcilerJob struct {
	uploadSvc  UploadService
	operations *operations
}

func NewReconciler(
	uploadSvc UploadService,
	interval time.Duration,
	batchSize int,
	observationContext *observation.Context,
) goroutine.BackgroundRoutine {
	job := reconcilerJob{
		uploadSvc:  uploadSvc,
		operations: newOperations(observationContext),
	}

	return goroutine.NewPeriodicGoroutine(context.Background(), interval, goroutine.HandlerFunc(func(ctx context.Context) error {
		return job.handleReconcile(ctx, batchSize)
	}))
}

func (j reconcilerJob) handleReconcile(ctx context.Context, batchSize int) (err error) {
	if err := j.handleReconcileFromFrontend(ctx, batchSize); err != nil {
		return err
	}

	if err := j.handleReconcileFromCodeintelDB(ctx, batchSize); err != nil {
		return err
	}

	return nil
}

// handleReconcileFromFrontend marks upload records that has no resolvable data in the codeintel-db.
func (j reconcilerJob) handleReconcileFromFrontend(ctx context.Context, batchSize int) (err error) {
	ids, err := j.uploadSvc.FrontendReconcileCandidates(ctx, batchSize)
	if err != nil {
		return err
	}

	j.operations.numReconcileScansFromFrontend.Add(float64(len(ids)))

	idsWithMeta, err := j.uploadSvc.IDsWithMeta(ctx, ids)
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
	j.operations.numReconcileDeletesFromFrontend.Add(float64(len(abandoned)))
	return nil
}

// handleReconcileFromCodeintelDB removes data from the codeintel-db that has no correlated upload
// in the frontend database.
func (j reconcilerJob) handleReconcileFromCodeintelDB(ctx context.Context, batchSize int) (err error) {
	ids, err := j.uploadSvc.CodeIntelDBReconcileCandidates(ctx, batchSize)
	if err != nil {
		return err
	}

	j.operations.numReconcileScansFromCodeIntelDB.Add(float64(len(ids)))

	dumps, err := j.uploadSvc.GetDumpsByIDs(ctx, ids)
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

	if err := j.uploadSvc.DeleteLsifDataByUploadIds(ctx, abandoned...); err != nil {
		return err
	}

	j.operations.numReconcileDeletesFromCodeIntelDB.Add(float64(len(abandoned)))
	return nil
}
