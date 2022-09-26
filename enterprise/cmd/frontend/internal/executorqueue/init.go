package executorqueue

import (
	"context"
	"net/http"

	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/observation"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/enterprise"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/executorqueue/handler"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/executorqueue/queues/batches"
	codeintelqueue "github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/executorqueue/queues/codeintel"
)

// Init initializes the executor endpoints required for use with the executor service.
func Init(
	ctx context.Context,
	db database.DB,
	conf conftypes.UnifiedWatchable,
	enterpriseServices *enterprise.Services,
	observationContext *observation.Context,
	codeintelUploadHandler http.Handler,
) error {
	accessToken := func() string { return conf.SiteConfig().ExecutorsAccessToken }

	// Register queues. If this set changes, be sure to also update the list of valid
	// queue names in ./metrics/queue_allocation.go, and register a metrics exporter
	// in the worker.
	//
	// Note: In order register a new queue type please change the validate() check code in enterprise/cmd/executor/config.go
	queueOptions := []handler.QueueOptions{
		codeintelqueue.QueueOptions(db, accessToken, observationContext),
		batches.QueueOptions(db, accessToken, observationContext),
	}

	queueHandler, err := newExecutorQueueHandler(db, queueOptions, accessToken, codeintelUploadHandler)
	if err != nil {
		return err
	}

	enterpriseServices.NewExecutorProxyHandler = queueHandler
	return nil
}
