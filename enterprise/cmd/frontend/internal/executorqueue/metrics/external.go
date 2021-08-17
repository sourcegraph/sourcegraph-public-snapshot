package metrics

import (
	"context"
	"sync"

	"github.com/cockroachdb/errors"

	metricsconfig "github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/executorqueue/metrics/config"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker/store"
)

type externalMetricsEmitter struct {
	queueName  string
	store      store.Store
	reporters  []reporter
	allocation metricsconfig.QueueAllocation
}

var _ goroutine.Handler = &externalMetricsEmitter{}

type reporter interface {
	ReportCount(ctx context.Context, queueName string, count int)
	GetAllocation(queueAllocation metricsconfig.QueueAllocation) float64
}

func (r *externalMetricsEmitter) Handle(ctx context.Context) error {
	count, err := r.store.QueuedCount(context.Background(), true, nil)
	if err != nil {
		return errors.Wrap(err, "dbworkerstore.QueuedCount")
	}

	fns := make([]func(), 0, len(r.reporters))
	for _, reporter := range r.reporters {
		reportCount := reporter.ReportCount
		count := int(float64(count) * reporter.GetAllocation(r.allocation))
		fns = append(fns, func() { reportCount(ctx, r.queueName, count) })
	}

	runParallel(fns)
	return nil
}

func runParallel(fns []func()) {
	var wg sync.WaitGroup
	wg.Add(len(fns))

	for _, fn := range fns {
		go fn()
	}

	wg.Wait()
}
