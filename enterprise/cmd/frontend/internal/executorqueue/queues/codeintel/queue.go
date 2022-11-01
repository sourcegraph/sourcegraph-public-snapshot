package codeintel

import (
	"context"

	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/executorqueue/handler"
	apiclient "github.com/sourcegraph/sourcegraph/enterprise/internal/executor"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/autoindexing"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/shared/types"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/workerutil"
)

func QueueOptions(autoIndexingSvc *autoindexing.Service, accessToken func() string, observationContext *observation.Context) handler.QueueOptions {
	recordTransformer := func(ctx context.Context, record workerutil.Record, resourceMetadata handler.ResourceMetadata) (apiclient.Job, error) {
		return transformRecord(record.(types.Index), resourceMetadata, accessToken())
	}

	return handler.QueueOptions{
		Name:              "codeintel",
		Store:             autoindexing.GetWorkerutilStore(autoIndexingSvc),
		RecordTransformer: recordTransformer,
	}
}
