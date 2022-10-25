package batches

import (
	"context"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/executorqueue/handler"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/batches/store"
	btypes "github.com/sourcegraph/sourcegraph/enterprise/internal/batches/types"
	apiclient "github.com/sourcegraph/sourcegraph/enterprise/internal/executor"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/workerutil"
)

func QueueOptions(db database.DB, _ func() string, observationContext *observation.Context) handler.QueueOptions {
	logger := log.Scoped("executor-queue.batches", "The executor queue handlers for the batches queue")
	recordTransformer := func(ctx context.Context, record workerutil.Record, _ handler.ResourceMetadata) (apiclient.Job, error) {
		batchesStore := store.New(db, observationContext, nil)
		return transformRecord(ctx, logger, batchesStore, record.(*btypes.BatchSpecWorkspaceExecutionJob))
	}

	store := store.NewBatchSpecWorkspaceExecutionWorkerStore(db.Handle(), observationContext)
	return handler.QueueOptions{
		Name:              "batches",
		Store:             store,
		RecordTransformer: recordTransformer,
	}
}
