package codeintel

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/executorqueue/handler"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/autoindexing"
	uploadsshared "github.com/sourcegraph/sourcegraph/internal/codeintel/uploads/shared"
	"github.com/sourcegraph/sourcegraph/internal/database"
	apiclient "github.com/sourcegraph/sourcegraph/internal/executor/types"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	dbworkerstore "github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker/store"
)

func QueueHandler(observationCtx *observation.Context, db database.DB, accessToken func() string) handler.QueueHandler[uploadsshared.AutoIndexJob] {
	recordTransformer := func(ctx context.Context, _ string, record uploadsshared.AutoIndexJob, resourceMetadata handler.ResourceMetadata) (apiclient.Job, error) {
		return transformRecord(ctx, db, record, resourceMetadata, accessToken())
	}

	store := dbworkerstore.New(observationCtx, db.Handle(), autoindexing.IndexWorkerStoreOptions)

	return handler.QueueHandler[uploadsshared.AutoIndexJob]{
		Name:              "codeintel",
		Store:             store,
		RecordTransformer: recordTransformer,
	}
}
