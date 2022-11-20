package codeintel

import (
	"context"

	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/executorqueue/handler"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/autoindexing"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/shared/types"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/executor"
	apiclient "github.com/sourcegraph/sourcegraph/internal/executor"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/workerutil"
)

func QueueOptions(db database.DB, accessToken func() string, observationContext *observation.Context) handler.QueueOptions {
	recordTransformer := func(ctx context.Context, record workerutil.Record, resourceMetadata handler.ResourceMetadata) (apiclient.Job, error) {
		return transformRecord(record.(types.Index), resourceMetadata, accessToken())
	}

	store := executor.NewExecutorStore(db.Handle(), autoindexing.IndexWorkerStoreOptions, observationContext)

	return handler.QueueOptions{
		Name:              "codeintel",
		Store:             store,
		RecordTransformer: recordTransformer,
	}
}
