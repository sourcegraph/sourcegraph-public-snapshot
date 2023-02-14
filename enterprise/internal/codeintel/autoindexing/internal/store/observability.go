package store

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/sourcegraph/sourcegraph/internal/memo"
	"github.com/sourcegraph/sourcegraph/internal/metrics"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type operations struct {
	// Commits
	processStaleSourcedCommits *observation.Operation

	// Indexes
	insertIndex                    *observation.Operation
	indexesInserted                prometheus.Counter
	getIndexes                     *observation.Operation
	getIndexByID                   *observation.Operation
	getIndexesByIDs                *observation.Operation
	getRecentIndexesSummary        *observation.Operation
	getLastIndexScanForRepository  *observation.Operation
	deleteIndexByID                *observation.Operation
	deleteIndexes                  *observation.Operation
	reindexIndexByID               *observation.Operation
	reindexIndexes                 *observation.Operation
	deleteIndexesWithoutRepository *observation.Operation
	isQueued                       *observation.Operation
	queueRepoRev                   *observation.Operation
	getQueuedRepoRev               *observation.Operation
	markRepoRevsAsProcessed        *observation.Operation

	// Index Configuration
	getIndexConfigurationByRepositoryID    *observation.Operation
	updateIndexConfigurationByRepositoryID *observation.Operation
	setInferenceScript                     *observation.Operation
	getInferenceScript                     *observation.Operation
	// Language Support
	getLanguagesRequestedBy   *observation.Operation
	setRequestLanguageSupport *observation.Operation

	insertDependencyIndexingJob *observation.Operation
	expireFailedRecords         *observation.Operation

	getRepoName                *observation.Operation
	topRepositoriesToConfigure *observation.Operation
	setConfigurationSummary    *observation.Operation
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
		// Commits
		processStaleSourcedCommits: op("ProcessStaleSourcedCommits"),

		// Indexes
		insertIndex:                    op("InsertIndex"),
		indexesInserted:                indexesInsertedCounter,
		getIndexes:                     op("GetIndexes"),
		getIndexByID:                   op("GetIndexByID"),
		getIndexesByIDs:                op("GetIndexesByIDs"),
		getRecentIndexesSummary:        op("GetRecentIndexesSummary"),
		getLastIndexScanForRepository:  op("GetLastIndexScanForRepository"),
		deleteIndexByID:                op("DeleteIndexByID"),
		deleteIndexes:                  op("DeleteIndexes"),
		reindexIndexByID:               op("ReindexIndexByID"),
		reindexIndexes:                 op("ReindexIndexes"),
		deleteIndexesWithoutRepository: op("DeleteIndexesWithoutRepository"),
		isQueued:                       op("IsQueued"),
		queueRepoRev:                   op("QueueRepoRev"),
		getQueuedRepoRev:               op("GetQueuedRepoRev"),
		markRepoRevsAsProcessed:        op("MarkRepoRevsAsProcessed"),

		// Index Configuration
		getIndexConfigurationByRepositoryID:    op("GetIndexConfigurationByRepositoryID"),
		updateIndexConfigurationByRepositoryID: op("UpdateIndexConfigurationByRepositoryID"),
		getInferenceScript:                     op("GetInferenceScript"),
		setInferenceScript:                     op("SetInferenceScript"),

		// Language Support
		getLanguagesRequestedBy:   op("GetLanguagesRequestedBy"),
		setRequestLanguageSupport: op("SetRequestLanguageSupport"),

		insertDependencyIndexingJob: op("InsertDependencyIndexingJob"),
		expireFailedRecords:         op("ExpireFailedRecords"),

		getRepoName:                op("GetRepoName"),
		topRepositoriesToConfigure: op("TopRepositoriesToConfigure"),
		setConfigurationSummary:    op("SetConfigurationSummary"),
	}
}
