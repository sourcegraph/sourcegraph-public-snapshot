package store

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/sourcegraph/sourcegraph/internal/memo"
	"github.com/sourcegraph/sourcegraph/internal/metrics"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type operations struct {
	getInferenceScript                     *observation.Operation
	setInferenceScript                     *observation.Operation
	repositoryExceptions                   *observation.Operation
	setRepositoryExceptions                *observation.Operation
	getIndexConfigurationByRepositoryID    *observation.Operation
	updateIndexConfigurationByRepositoryID *observation.Operation
	topRepositoriesToConfigure             *observation.Operation
	repositoryIDsWithConfiguration         *observation.Operation
	getLastIndexScanForRepository          *observation.Operation
	setConfigurationSummary                *observation.Operation
	truncateConfigurationSummary           *observation.Operation
	getRepositoriesForIndexScan            *observation.Operation
	getQueuedRepoRev                       *observation.Operation
	markRepoRevsAsProcessed                *observation.Operation
	isQueued                               *observation.Operation
	isQueuedRootIndexer                    *observation.Operation
	insertAutoIndexJobs                    *observation.Operation
	insertDependencyIndexingJob            *observation.Operation
	queueRepoRev                           *observation.Operation

	indexesInserted prometheus.Counter
}

var (
	indexesInsertedCounterMemo = memo.NewMemoizedConstructorWithArg(func(r prometheus.Registerer) (prometheus.Counter, error) {
		indexesInsertedCounter := prometheus.NewCounter(prometheus.CounterOpts{
			Name: "src_codeintel_dbstore_indexes_inserted",
			Help: "The number of codeintel index records inserted.",
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
			"codeintel_autoindexing_store",
			metrics.WithLabels("op"),
			metrics.WithCountHelp("Total number of method invocations."),
		)
	})

	op := func(name string) *observation.Operation {
		return observationCtx.Operation(observation.Op{
			Name:              fmt.Sprintf("codeintel.autoindexing.store.%s", name),
			MetricLabelValues: []string{name},
			Metrics:           m,
		})
	}

	indexesInsertedCounter, _ := indexesInsertedCounterMemo.Init(observationCtx.Registerer)

	return &operations{
		getInferenceScript:                     op("GetInferenceScript"),
		setInferenceScript:                     op("SetInferenceScript"),
		repositoryExceptions:                   op("RepositoryExceptions"),
		setRepositoryExceptions:                op("SetRepositoryExceptions"),
		getIndexConfigurationByRepositoryID:    op("GetIndexConfigurationByRepositoryID"),
		updateIndexConfigurationByRepositoryID: op("UpdateIndexConfigurationByRepositoryID"),
		topRepositoriesToConfigure:             op("TopRepositoriesToConfigure"),
		repositoryIDsWithConfiguration:         op("RepositoryIDsWithConfiguration"),
		getLastIndexScanForRepository:          op("GetLastIndexScanForRepository"),
		setConfigurationSummary:                op("SetConfigurationSummary"),
		truncateConfigurationSummary:           op("TruncateConfigurationSummary"),
		getRepositoriesForIndexScan:            op("GetRepositoriesForIndexScan"),
		getQueuedRepoRev:                       op("GetQueuedRepoRev"),
		markRepoRevsAsProcessed:                op("MarkRepoRevsAsProcessed"),
		isQueued:                               op("IsQueued"),
		isQueuedRootIndexer:                    op("IsQueuedRootIndexer"),
		insertAutoIndexJobs:                    op("InsertJobs"),
		insertDependencyIndexingJob:            op("InsertDependencyIndexingJob"),
		queueRepoRev:                           op("QueueRepoRev"),

		indexesInserted: indexesInsertedCounter,
	}
}
