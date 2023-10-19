package executorqueue

import (
	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/conf/confdefaults"
	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
	"github.com/sourcegraph/sourcegraph/internal/conf/deploy"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/observation"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/enterprise"
)

// Init initializes the executor endpoints required for use with the executor service.
func Init(
	observationCtx *observation.Context,
	db database.DB,
	conf conftypes.UnifiedWatchable,
	enterpriseServices *enterprise.Services,
) error {
	codeintelUploadHandler := enterpriseServices.NewCodeIntelUploadHandler(false)
	batchesWorkspaceFileGetHandler := enterpriseServices.BatchesChangesFileGetHandler
	batchesWorkspaceFileExistsHandler := enterpriseServices.BatchesChangesFileGetHandler

	accessToken := func() string {
		if deploy.IsSingleBinary() {
			return confdefaults.AppInMemoryExecutorPassword
		}
		return conf.SiteConfig().ExecutorsAccessToken
	}

	logger := log.Scoped("executorqueue")

	queueHandler := newExecutorQueuesHandler(
		observationCtx,
		db,
		logger,
		accessToken,
		codeintelUploadHandler,
		batchesWorkspaceFileGetHandler,
		batchesWorkspaceFileExistsHandler,
	)

	enterpriseServices.NewExecutorProxyHandler = queueHandler
	return nil
}
