package batches

import (
	"database/sql"

	apiserver "github.com/sourcegraph/sourcegraph/enterprise/cmd/executor-queue/internal/server"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/batches/background"
	btypes "github.com/sourcegraph/sourcegraph/enterprise/internal/batches/types"
	apiclient "github.com/sourcegraph/sourcegraph/enterprise/internal/executor"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/workerutil"
)

func QueueOptions(db dbutil.DB, config *Config, observationContext *observation.Context) apiserver.QueueOptions {
	recordTransformer := func(record workerutil.Record) (apiclient.Job, error) {
		return transformRecord(record.(*btypes.BatchSpecExecution), config)
	}

	return apiserver.QueueOptions{
		Store:             background.NewExecutorStore(basestore.NewWithDB(db, sql.TxOptions{}), observationContext),
		RecordTransformer: recordTransformer,
	}
}
