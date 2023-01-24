package executorqueue

import (
	"context"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
	"github.com/sourcegraph/sourcegraph/internal/database"
	metricsstore "github.com/sourcegraph/sourcegraph/internal/metrics/store"
	"github.com/sourcegraph/sourcegraph/internal/observation"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/enterprise"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/executorqueue/handler"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/executorqueue/queues/batches"
	codeintelqueue "github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/executorqueue/queues/codeintel"
)

// Init initializes the executor endpoints required for use with the executor service.
func Init(
	ctx context.Context,
	observationCtx *observation.Context,
	db database.DB,
	conf conftypes.UnifiedWatchable,
	enterpriseServices *enterprise.Services,
) error {
	codeintelUploadHandler := enterpriseServices.NewCodeIntelUploadHandler(false)
	batchesWorkspaceFileGetHandler := enterpriseServices.BatchesChangesFileGetHandler
	batchesWorkspaceFileExistsHandler := enterpriseServices.BatchesChangesFileGetHandler
	accessToken := func() string { return conf.SiteConfig().ExecutorsAccessToken }
	logger := log.Scoped("executorqueue", "")

	metricsStore := metricsstore.NewDistributedStore("executors:")
	executorStore := db.Executors()

	// Register queues. If this set changes, be sure to also update the list of valid
	// queue names in ./metrics/queue_allocation.go, and register a metrics exporter
	// in the worker.
	//
	// Note: In order register a new queue type please change the validate() check code in enterprise/cmd/executor/config.go
	codeintelHandler := handler.NewHandler(executorStore, metricsStore, codeintelqueue.QueueOptions(observationCtx, db, accessToken))
	batchesHandler := handler.NewHandler(executorStore, metricsStore, batches.QueueOptions(observationCtx, db, accessToken))
	queueOptions := []handler.ExecutorHandler{codeintelHandler, batchesHandler}

	queueHandler := newExecutorQueueHandler(
		logger,
		db,
		queueOptions,
		accessToken,
		codeintelUploadHandler,
		batchesWorkspaceFileGetHandler,
		batchesWorkspaceFileExistsHandler,
	)

	enterpriseServices.NewExecutorProxyHandler = queueHandler
	return nil
}
