package codeintel

import (
	"context"
	"database/sql"

	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/executorqueue/handler"
	apiclient "github.com/sourcegraph/sourcegraph/enterprise/internal/executor"
	store "github.com/sourcegraph/sourcegraph/internal/codeintel/stores/dbstore"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/workerutil"
)

func QueueOptions(db database.DB, accessToken func() string, observationContext *observation.Context) handler.QueueOptions {
	recordTransformer := func(ctx context.Context, record workerutil.Record) (apiclient.Job, error) {
		return transformRecord(record.(store.Index), accessToken())
	}

	return handler.QueueOptions{
		Name:              "codeintel",
		Store:             store.WorkerutilIndexStore(basestore.NewWithDB(db, sql.TxOptions{}), observationContext),
		RecordTransformer: recordTransformer,
	}
}
