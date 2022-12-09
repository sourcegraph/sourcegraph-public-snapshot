package executorqueue

import (
	"context"
	"sync"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/shared/types"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/workerutil"
	"github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker/store"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type externalEmitter[T workerutil.Record] struct {
	queueName  string
	store      store.Store[T]
	reporters  []reporter
	allocation QueueAllocation
}

var _ goroutine.Handler = &externalEmitter[types.Index]{}

type reporter interface {
	ReportCount(ctx context.Context, queueName string, count int)
	GetAllocation(queueAllocation QueueAllocation) float64
}

func (r *externalEmitter[T]) Handle(ctx context.Context) error {
	count, err := r.store.QueuedCount(context.Background(), true)
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
		go func(fn func()) {
			defer wg.Done()
			fn()
		}(fn)
	}

	wg.Wait()
}
