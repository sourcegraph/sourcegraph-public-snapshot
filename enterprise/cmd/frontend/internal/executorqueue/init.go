package executorqueue

import (
	"context"
	"strconv"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
	"github.com/sourcegraph/sourcegraph/internal/conf/deploy"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/env"
	metricsstore "github.com/sourcegraph/sourcegraph/internal/metrics/store"
	"github.com/sourcegraph/sourcegraph/internal/observation"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/enterprise"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/executorqueue/handler"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/executorqueue/queues/batches"
	codeintelqueue "github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/executorqueue/queues/codeintel"
)

func queueDisableAccessTokenDefault() string {
	isSingleProgram := deploy.IsDeployTypeSingleProgram(deploy.Type())
	if isSingleProgram {
		return "true"
	}
	return "false"
}

var queueDisableAccessToken = env.Get("EXECUTOR_QUEUE_DISABLE_ACCESS_TOKEN_INSECURE", queueDisableAccessTokenDefault(), "Disable usage of an access token between executors and Sourcegraph (DANGEROUS")

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

	logger := log.Scoped("executorqueue", "")

	accessToken := func() (token string, accessTokenEnabled bool) {
		token = conf.SiteConfig().ExecutorsAccessToken
		wantDisableAccessToken, _ := strconv.ParseBool(queueDisableAccessToken)

		if wantDisableAccessToken {
			isSingleProgram := deploy.IsDeployTypeSingleProgram(deploy.Type())
			isSingleDockerContainer := deploy.IsDeployTypeSingleDockerContainer(deploy.Type())
			allowedDeployType := isSingleProgram || isSingleDockerContainer || env.InsecureDev
			if allowedDeployType && token == "" {
				// Disable the access token.
				return "", false
			}
			// Respect the access token.
			logger.Warn("access token may only be disabled if executors.accessToken is empty in site config AND the deployment type is single-program, single-docker-container, or dev")
			return token, true
		}

		// Respect the access token.
		return token, true
	}

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
