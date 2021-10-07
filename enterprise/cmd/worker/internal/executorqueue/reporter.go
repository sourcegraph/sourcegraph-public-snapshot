package executorqueue

import (
	"context"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker/store"
)

func NewMetricReporter(observationContext *observation.Context, queueName string, store store.Store, metricsConfig *Config) (goroutine.BackgroundRoutine, error) {
	// Emit metrics to control alerts.
	initPrometheusMetric(observationContext, queueName, store)

	// Emit metrics to control executor auto-scaling.
	return initExternalMetricReporters(queueName, store, metricsConfig)
}

func initExternalMetricReporters(queueName string, store store.Store, metricsConfig *Config) (goroutine.BackgroundRoutine, error) {
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
	return goroutine.NewPeriodicGoroutine(ctx, 5*time.Second, &externalEmitter{
		queueName:  queueName,
		store:      store,
		reporters:  reporters,
		allocation: metricsConfig.Allocations[queueName],
	}), nil
}
