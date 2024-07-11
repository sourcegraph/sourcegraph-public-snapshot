package jobstore

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/sourcegraph/sourcegraph/internal/memo"
	"github.com/sourcegraph/sourcegraph/internal/metrics"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type operations struct {
	isQueued             *observation.Operation
	insertIndexingJobs   *observation.Operation
	indexingJobsInserted prometheus.Counter
}

var (
	indexesInsertedCounterMemo = memo.NewMemoizedConstructorWithArg(func(r prometheus.Registerer) (prometheus.Counter, error) {
		indexesInsertedCounter := prometheus.NewCounter(prometheus.CounterOpts{
			Name: "src_codeintel_dbstore_syntactic_indexing_jobs_inserted",
			Help: "The number of codeintel syntactic indexing jobs inserted.",
		})
		r.MustRegister(indexesInsertedCounter)
		return indexesInsertedCounter, nil
	})
	m = new(metrics.SingletonREDMetrics)
)

func newOperations(observationCtx *observation.Context) *operations {
	m := m.Get(func() *metrics.REDMetrics {
		return metrics.NewREDMetrics(
			observationCtx.Registerer,
			"codeintel_syntactic_indexing_jobs_store",
			metrics.WithLabels("op"),
			metrics.WithCountHelp("Total number of method invocations."),
		)
	})

	op := func(name string) *observation.Operation {
		return observationCtx.Operation(observation.Op{
			Name:              fmt.Sprintf("codeintel.syntacticindexing.store.%s", name),
			MetricLabelValues: []string{name},
			Metrics:           m,
		})
	}

	indexesInsertedCounter, _ := indexesInsertedCounterMemo.Init(observationCtx.Registerer)
	return &operations{
		isQueued:             op("IsQueued"),
		insertIndexingJobs:   op("InsertIndexingJobs"),
		indexingJobsInserted: indexesInsertedCounter,
	}
}
