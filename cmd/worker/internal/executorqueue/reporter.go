package executorqueue

import (
	"context"
	"time"

	executorutil "github.com/sourcegraph/sourcegraph/internal/executor/util"
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

func initExternalMetricReporters[T workerutil.Record](queueName string, store_ store.Store[T], metricsConfig *Config) (goroutine.BackgroundRoutine, error) {
	reporters, err := configureReporters(metricsConfig)
	if err != nil {
		return nil, err
	}

	ctx := context.Background()
	return goroutine.NewPeriodicGoroutine(
		ctx,
		&externalEmitter[T]{
			queueName:  queueName,
			countFuncs: []func(context.Context, store.RecordState) (int, error){store_.CountByState},
			reporters:  reporters,
			allocation: metricsConfig.Allocations[queueName],
		},
		goroutine.WithName("executors.autoscaler-metrics"),
		goroutine.WithDescription("emits metrics to GCP/AWS for auto-scaling"),
		goroutine.WithInterval(5*time.Second),
	), nil
}

// NewMultiqueueMetricReporter returns a periodic background routine that reports the sum of the lengths all configured queues.
// This does not reinitialise Prometheus metrics as is done in NewMetricReporter, as this only needs to be done once and is
// already done for the single queue metrics.
func NewMultiqueueMetricReporter(queueNames []string, metricsConfig *Config, countFuncs ...func(_ context.Context, bitset store.RecordState) (int, error)) (goroutine.BackgroundRoutine, error) {
	reporters, err := configureReporters(metricsConfig)
	if err != nil {
		return nil, err
	}

	queueStr := executorutil.FormatQueueNamesForMetrics("", queueNames)
	ctx := context.Background()
	return goroutine.NewPeriodicGoroutine(
		ctx,
		&externalEmitter[workerutil.Record]{
			queueName:  queueStr,
			countFuncs: countFuncs,
			reporters:  reporters,
			// TODO this is a temp fix to get an allocation for both
			allocation: metricsConfig.Allocations[queueNames[0]],
		},
		goroutine.WithName("multiqueue-executors.autoscaler-metrics"),
		goroutine.WithDescription("emits multiqueue metrics to GCP/AWS for auto-scaling"),
		goroutine.WithInterval(30*time.Second),
	), nil
}

func configureReporters(metricsConfig *Config) ([]reporter, error) {
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
	return reporters, nil
}
