package metrics

import (
	"context"
	"sync"
	"time"

	"github.com/cockroachdb/errors"

	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/executorqueue/handler"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/executorqueue/metrics/aws"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/executorqueue/metrics/gcp"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker/store"
)

const metricReportInterval = 5 * time.Second

type ReportCounter interface {
	ReportCount(ctx context.Context, queueName string, store store.Store, count int)
}

type reporters map[string]ReportCounter

func (r reporters) RegisterReporter(name string, init func(string) (ReportCounter, error)) error {
	if reporter, err := init(metricsConfig.EnvironmentLabel); err != nil {
		return err
	} else if reporter != nil {
		r[name] = reporter
	}

	return nil
}

func Init(observationContext *observation.Context, queueOptions map[string]handler.QueueOptions) error {
	// Emit metrics to control alerts
	initPrometheusMetrics(observationContext, queueOptions)

	initCustomMetricReporters(queueOptions)

	return nil
}

func initCustomMetricReporters(queueOptions map[string]handler.QueueOptions) {
	reporters := make(reporters)
	reporters.RegisterReporter("aws", func(s string) (ReportCounter, error) { return aws.InitReportCounter(s) })
	reporters.RegisterReporter("gcp", func(s string) (ReportCounter, error) { return gcp.InitReportCounter(s) })

	routines := make([]goroutine.BackgroundRoutine, 0, len(queueOptions))

	for queueName, queue := range queueOptions {
		reporter := &externalReportCounter{
			store:      queue.Store,
			reporters:  reporters,
			allocation: metricsConfig.Allocations[queueName],
		}
		routines = append(routines, goroutine.NewPeriodicGoroutine(
			context.Background(),
			metricReportInterval,
			reporter,
		))
	}

	// Emit metrics to control executor auto-scaling
	go goroutine.MonitorBackgroundRoutines(
		context.Background(),
		routines...,
	)
}

type externalReportCounter struct {
	queueName  string
	reporters  map[string]ReportCounter
	store      store.Store
	allocation allocationConfig
}

func (r *externalReportCounter) Handle(ctx context.Context) error {
	count, err := r.store.QueuedCount(context.Background(), true, nil)
	if err != nil {
		return errors.Wrap(err, "dbworkerstore.QueuedCount")
	}
	allocationConfigured := r.allocation.IsConfigured()
	var wg sync.WaitGroup
	for reporterName, reporter := range r.reporters {
		wg.Add(1)
		go func(reporterName string, reporter ReportCounter) {
			defer wg.Done()

			count := count
			if allocationConfigured {
				count = int(float64(count) * r.allocation[reporterName])
			}
			reporter.ReportCount(ctx, r.queueName, r.store, count)
		}(reporterName, reporter)
	}
	wg.Wait()
	return nil
}
