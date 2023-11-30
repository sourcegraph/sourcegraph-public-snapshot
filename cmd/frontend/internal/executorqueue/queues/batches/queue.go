package batches

import (
	"context"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/executorqueue/handler"
	bstore "github.com/sourcegraph/sourcegraph/internal/batches/store"
	btypes "github.com/sourcegraph/sourcegraph/internal/batches/types"
	"github.com/sourcegraph/sourcegraph/internal/database"
	apiclient "github.com/sourcegraph/sourcegraph/internal/executor/types"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

func QueueHandler(observationCtx *observation.Context, db database.DB, _ func() string) handler.QueueHandler[*btypes.BatchSpecWorkspaceExecutionJob] {
	logger := log.Scoped("executor-queue.batches")
	recordTransformer := func(ctx context.Context, version string, record *btypes.BatchSpecWorkspaceExecutionJob, _ handler.ResourceMetadata) (apiclient.Job, error) {
		batchesStore := bstore.New(db, observationCtx, nil)
		return transformRecord(ctx, logger, batchesStore, record, version)
	}

	store := bstore.NewBatchSpecWorkspaceExecutionWorkerStore(observationCtx, db.Handle())
	return handler.QueueHandler[*btypes.BatchSpecWorkspaceExecutionJob]{
		Name:              "batches",
		Store:             store,
		RecordTransformer: recordTransformer,
	}
}
