package metrics

import (
	"context"
	"time"

	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/executorqueue/handler"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/executorqueue/metrics/aws"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/executorqueue/metrics/config"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/executorqueue/metrics/gcp"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

var metricsConfig = &config.Config{}

func init() {
	metricsConfig.Load()
}

func Init(observationContext *observation.Context, queueOptions map[string]handler.QueueOptions) error {
	if err := metricsConfig.Validate(); err != nil {
		return err
	}

	// Emit metrics to control alerts
	initPrometheusMetrics(observationContext, queueOptions)

	// Emit metrics to control executor auto-scaling
	return initExternalMetricReporters(queueOptions)
}

func initExternalMetricReporters(queueOptions map[string]handler.QueueOptions) error {
	awsReporter, err := aws.NewReporter(metricsConfig.EnvironmentLabel)
	if err != nil {
		return err
	}

	gcsReporter, err := gcp.NewReporter(metricsConfig.EnvironmentLabel)
	if err != nil {
		return err
	}

	ctx := context.Background()
	routines := make([]goroutine.BackgroundRoutine, 0, len(queueOptions))
	for queueName, queue := range queueOptions {
		routines = append(routines, goroutine.NewPeriodicGoroutine(ctx, 5*time.Second, &externalMetricsEmitter{
			queueName:  queueName,
			store:      queue.Store,
			reporters:  []reporter{awsReporter, gcsReporter},
			allocation: metricsConfig.Allocations[queueName],
		}))
	}

	go goroutine.MonitorBackgroundRoutines(ctx, routines...)
	return nil
}
