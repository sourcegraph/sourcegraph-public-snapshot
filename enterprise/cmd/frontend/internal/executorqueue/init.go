package executorqueue

import (
	"context"
	"os"

	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/observation"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/enterprise"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/codeintel"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/executorqueue/handler"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/executorqueue/queues/batches"
	codeintelqueue "github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/executorqueue/queues/codeintel"
	"github.com/sourcegraph/sourcegraph/internal/oobmigration"
)

// Init initializes the executor endpoints required for use with the executor service.
func Init(ctx context.Context, db dbutil.DB, conf conftypes.UnifiedWatchable, outOfBandMigrationRunner *oobmigration.Runner, enterpriseServices *enterprise.Services, observationContext *observation.Context, services *codeintel.Services) error {
	accessToken := func() string {
		if accessToken := conf.SiteConfig().ExecutorsAccessToken; accessToken != "" {
			return accessToken
		}
		// Fallback to old environment variable, for a smooth rollout.
		return os.Getenv("EXECUTOR_FRONTEND_PASSWORD")
	}

	// Register queues. If this set changes, be sure to also update the list of valid
	// queue names in ./metrics/queue_allocation.go, and register a metrics exporter
	// in the worker.
	queueOptions := []handler.QueueOptions{
		codeintelqueue.QueueOptions(db, accessToken, observationContext),
		batches.QueueOptions(db, accessToken, observationContext),
	}

	handler, err := codeintel.NewCodeIntelUploadHandler(ctx, conf, db, true, services)
	if err != nil {
		return err
	}

	queueHandler, err := newExecutorQueueHandler(database.NewDB(db).Executors(), queueOptions, accessToken, handler)
	if err != nil {
		return err
	}

	enterpriseServices.NewExecutorProxyHandler = queueHandler
	return nil
}
