package executorqueue

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/observation"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/enterprise"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/codeintel"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/executorqueue/handler"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/executorqueue/queues/batches"
	codeintelqueue "github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/executorqueue/queues/codeintel"
	"github.com/sourcegraph/sourcegraph/internal/oobmigration"
)

// Init initializes the executor endpoints required for use with the executor service.
func Init(ctx context.Context, db database.DB, conf conftypes.UnifiedWatchable, outOfBandMigrationRunner *oobmigration.Runner, enterpriseServices *enterprise.Services, observationContext *observation.Context, services *codeintel.Services) error {
	accessToken := func() string { return conf.SiteConfig().ExecutorsAccessToken }

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

	queueHandler, err := newExecutorQueueHandler(db.Executors(), queueOptions, accessToken, handler)
	if err != nil {
		return err
	}

	enterpriseServices.NewExecutorProxyHandler = queueHandler
	return nil
}
