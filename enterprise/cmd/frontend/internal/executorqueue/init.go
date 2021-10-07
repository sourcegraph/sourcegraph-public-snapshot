package executorqueue

import (
	"context"
	"log"

	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/observation"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/enterprise"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/codeintel"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/executorqueue/config"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/executorqueue/handler"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/executorqueue/queues/batches"
	codeintelqueue "github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/executorqueue/queues/codeintel"
	"github.com/sourcegraph/sourcegraph/internal/oobmigration"
)

type configuration interface {
	Load()
	Validate() error
}

var (
	sharedConfig    *config.SharedConfig
	codeintelConfig *codeintelqueue.Config
	batchesConfig   *batches.Config
)

// Load configs at startup. We cannot use env.Get after the application started.
func init() {
	sharedConfig = &config.SharedConfig{}
	codeintelConfig = &codeintelqueue.Config{Shared: sharedConfig}
	batchesConfig = &batches.Config{Shared: sharedConfig}
	configs := []configuration{sharedConfig, codeintelConfig, batchesConfig}

	for _, config := range configs {
		config.Load()
	}
}

// Init initializes the executor endpoints required for use with the executor service.
func Init(ctx context.Context, db dbutil.DB, outOfBandMigrationRunner *oobmigration.Runner, enterpriseServices *enterprise.Services, observationContext *observation.Context) error {
	for _, config := range []configuration{sharedConfig, codeintelConfig, batchesConfig} {
		if err := config.Validate(); err != nil {
			log.Fatalf("failed to load config: %s", err)
		}
	}

	// Register queues. If this set changes, be sure to also update the list of valid
	// queue names in ./metrics/queue_allocation.go, and register a metrics exporter
	// in the worker.
	queueOptions := map[string]handler.QueueOptions{
		"codeintel": codeintelqueue.QueueOptions(db, codeintelConfig, observationContext),
		"batches":   batches.QueueOptions(db, batchesConfig, observationContext),
	}

	handler, err := codeintel.NewCodeIntelUploadHandler(ctx, db, true)
	if err != nil {
		return err
	}

	queueHandler, err := newExecutorQueueHandler(queueOptions, handler)
	if err != nil {
		return err
	}

	enterpriseServices.NewExecutorProxyHandler = queueHandler
	return nil
}
