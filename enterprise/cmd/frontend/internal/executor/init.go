package executor

import (
	"context"
	"log"

	"github.com/inconshreveable/log15"
	"github.com/opentracing/opentracing-go"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/enterprise"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/executor-queue/config"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/executor-queue/queues/batches"
	codeintel2 "github.com/sourcegraph/sourcegraph/enterprise/cmd/executor-queue/queues/codeintel"
	apiserver "github.com/sourcegraph/sourcegraph/enterprise/cmd/executor-queue/server"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/codeintel"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/oobmigration"
	"github.com/sourcegraph/sourcegraph/internal/trace"
)

type configuration interface {
	Load()
	Validate() error
}

var (
	sharedConfig    = &config.SharedConfig{}
	codeintelConfig = &codeintel2.Config{Shared: sharedConfig}
	batchesConfig   = &batches.Config{Shared: sharedConfig}
	configs         = []configuration{sharedConfig, codeintelConfig, batchesConfig}
)

func init() {
	for _, config := range configs {
		config.Load()
	}
}

func Init(ctx context.Context, db dbutil.DB, outOfBandMigrationRunner *oobmigration.Runner, enterpriseServices *enterprise.Services) error {
	for _, config := range configs {
		if err := config.Validate(); err != nil {
			log.Fatalf("failed to load config: %s", err)
		}
	}

	handler, err := codeintel.NewCodeIntelUploadHandler(ctx, db, true)
	if err != nil {
		return err
	}

	// Initialize tracing/metrics
	observationContext := &observation.Context{
		Logger:     log15.Root(),
		Tracer:     &trace.Tracer{Tracer: opentracing.GlobalTracer()},
		Registerer: prometheus.DefaultRegisterer,
	}

	// Initialize queues
	queueOptions := map[string]apiserver.QueueOptions{
		"codeintel": codeintel2.QueueOptions(db, codeintelConfig, observationContext),
		"batches":   batches.QueueOptions(db, batchesConfig, observationContext),
	}

	for queueName, options := range queueOptions {
		// Make local copy of queue name for capture below
		queueName, store := queueName, options.Store

		prometheus.DefaultRegisterer.MustRegister(prometheus.NewGaugeFunc(prometheus.GaugeOpts{
			Name:        "src_executor_total",
			Help:        "Total number of jobs in the queued state.",
			ConstLabels: map[string]string{"queue": queueName},
		}, func() float64 {
			// TODO(efritz) - do not count soft-deleted code intel index records
			count, err := store.QueuedCount(context.Background(), nil)
			if err != nil {
				log15.Error("Failed to get queued job count", "queue", queueName, "error", err)
			}

			return float64(count)
		}))
	}

	proxyHandler, err := newInternalProxyHandler(handler, queueOptions)
	if err != nil {
		return err
	}

	enterpriseServices.NewExecutorProxyHandler = proxyHandler
	return nil
}
