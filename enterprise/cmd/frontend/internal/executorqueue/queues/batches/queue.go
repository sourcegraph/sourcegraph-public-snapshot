package batches

import (
	"context"
	"database/sql"

	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/executorqueue/handler"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/batches/background"
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

	return handler.QueueOptions{
		Store:             background.NewExecutorStore(basestore.NewWithDB(db, sql.TxOptions{}), observationContext),
		RecordTransformer: recordTransformer,
	}
}
