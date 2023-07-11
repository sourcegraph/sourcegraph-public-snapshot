package codeintel

import (
	"context"

	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/executorqueue/handler"
	apiclient "github.com/sourcegraph/sourcegraph/enterprise/internal/executor/types"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/autoindexing"
	uploadsshared "github.com/sourcegraph/sourcegraph/internal/codeintel/uploads/shared"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	dbworkerstore "github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker/store"
)

func QueueHandler(observationCtx *observation.Context, db database.DB, accessToken func() string) handler.QueueHandler[uploadsshared.Index] {
	recordTransformer := func(ctx context.Context, _ string, record uploadsshared.Index, resourceMetadata handler.ResourceMetadata) (apiclient.Job, error) {
		return transformRecord(ctx, db, record, resourceMetadata, accessToken())
	}

	store := dbworkerstore.New(observationCtx, db.Handle(), autoindexing.IndexWorkerStoreOptions)

	return handler.QueueHandler[uploadsshared.Index]{
		Name:              "codeintel",
		Store:             store,
		RecordTransformer: recordTransformer,
	}
}
