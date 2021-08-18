package metrics

import (
	"context"
	"time"

	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/executorqueue/handler"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

func Init(observationContext *observation.Context, queueOptions map[string]handler.QueueOptions, metricsConfig *Config) error {
	// Emit metrics to control alerts
	initPrometheusMetrics(observationContext, queueOptions)

	// Emit metrics to control executor auto-scaling
	if err := initExternalMetricReporters(queueOptions, metricsConfig); err != nil {
		return err
	}

	return nil
}

func initExternalMetricReporters(queueOptions map[string]handler.QueueOptions, metricsConfig *Config) error {
	awsReporter, err := newAWSReporter(metricsConfig)
	if err != nil {
		return err
	}

	gcsReporter, err := newGCPReporter(metricsConfig)
	if err != nil {
		return err
	}

	var reporters []reporter
	if awsReporter != nil {
		reporters = append(reporters, awsReporter)
	}
	if gcsReporter != nil {
		reporters = append(reporters, gcsReporter)
	}

	ctx := context.Background()
	routines := make([]goroutine.BackgroundRoutine, 0, len(queueOptions))
	for queueName, queue := range queueOptions {
		routines = append(routines, goroutine.NewPeriodicGoroutine(ctx, 5*time.Second, &externalEmitter{
			queueName:  queueName,
			store:      queue.Store,
			reporters:  reporters,
			allocation: metricsConfig.Allocations[queueName],
		}))
	}

	go goroutine.MonitorBackgroundRoutines(ctx, routines...)
	return nil
}
