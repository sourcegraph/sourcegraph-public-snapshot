package batches

import (
	"context"
	"database/sql"

	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/executorqueue/handler"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/batches/background"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/batches/store"
	btypes "github.com/sourcegraph/sourcegraph/enterprise/internal/batches/types"
	apiclient "github.com/sourcegraph/sourcegraph/enterprise/internal/executor"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/workerutil"
)

func QueueOptions(db dbutil.DB, config *Config, observationContext *observation.Context) handler.QueueOptions {
	recordTransformer := func(ctx context.Context, record workerutil.Record) (apiclient.Job, error) {
		return transformRecord(ctx, db, record.(*btypes.BatchSpecExecution), config)
	}

	store := background.NewExecutorStore(basestore.NewHandleWithDB(db, sql.TxOptions{}), observationContext)
	return handler.QueueOptions{
		Store:                  store,
		RecordTransformer:      recordTransformer,
		CanceledRecordsFetcher: store.FetchCanceled,
	}
}

func WorkspaceExecutionQueueOptions(db dbutil.DB, config *Config, observationContext *observation.Context) handler.QueueOptions {
	recordTransformer := func(ctx context.Context, record workerutil.Record) (apiclient.Job, error) {
		batchesStore := store.New(db, observationContext, nil)
		return transformBatchSpecWorkspaceExecutionJobRecord(ctx, batchesStore, record.(*btypes.BatchSpecWorkspaceExecutionJob), config)
	}

	store := background.NewBatchSpecWorkspaceExecutionWorkerStore(basestore.NewHandleWithDB(db, sql.TxOptions{}), observationContext)
	return handler.QueueOptions{
		Store:             store,
		RecordTransformer: recordTransformer,
	}
}
