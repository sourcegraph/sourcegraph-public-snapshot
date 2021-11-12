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

func QueueOptions(db dbutil.DB, accessToken func() string, observationContext *observation.Context) handler.QueueOptions {
	recordTransformer := func(ctx context.Context, record workerutil.Record) (apiclient.Job, error) {
		batchesStore := store.New(db, observationContext, nil)
		return transformRecord(ctx, batchesStore, record.(*btypes.BatchSpecWorkspaceExecutionJob), accessToken())
	}

	store := background.NewBatchSpecWorkspaceExecutionWorkerStore(basestore.NewHandleWithDB(db, sql.TxOptions{}), observationContext)
	return handler.QueueOptions{
		Name:                   "batches",
		Store:                  store,
		RecordTransformer:      recordTransformer,
		CanceledRecordsFetcher: store.FetchCanceled,
	}
}
