package executorqueue

import (
	"context"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/workerutil"
	"github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker/store"
)

func NewMetricReporter[T workerutil.Record](observationCtx *observation.Context, queueName string, store store.Store[T], metricsConfig *Config) (goroutine.BackgroundRoutine, error) {
	// Emit metrics to control alerts.
	initPrometheusMetric(observationCtx, queueName, store)

	// Emit metrics to control executor auto-scaling.
	return initExternalMetricReporters(queueName, store, metricsConfig)
}

func initExternalMetricReporters[T workerutil.Record](queueName string, store store.Store[T], metricsConfig *Config) (goroutine.BackgroundRoutine, error) {
	awsReporter, err := newAWSReporter(metricsConfig)
	if err != nil {
		return nil, err
	}

	gcsReporter, err := newGCPReporter(metricsConfig)
	if err != nil {
		return nil, err
	}

	var reporters []reporter
	if awsReporter != nil {
		reporters = append(reporters, awsReporter)
	}
	if gcsReporter != nil {
		reporters = append(reporters, gcsReporter)
	}

	ctx := context.Background()
	return goroutine.NewPeriodicGoroutine(
		ctx,
		&externalEmitter[T]{
			queueName:  queueName,
			store:      store,
			reporters:  reporters,
			allocation: metricsConfig.Allocations[queueName],
		},
		goroutine.WithName("executors.autoscaler-metrics"),
		goroutine.WithDescription("emits metrics to GCP/AWS for auto-scaling"),
		goroutine.WithInterval(5*time.Second),
	), nil
}
