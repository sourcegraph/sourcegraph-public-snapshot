package store

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/sourcegraph/sourcegraph/internal/metrics"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type operations struct {
	// Commits
	getStaleSourcedCommits *observation.Operation
	deleteSourcedCommits   *observation.Operation
	updateSourcedCommits   *observation.Operation

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
}

func newOperations(observationContext *observation.Context) *operations {
	m := metrics.NewREDMetrics(
		observationContext.Registerer,
		"codeintel_autoindexing_store",
		metrics.WithLabels("op"),
		metrics.WithCountHelp("Total number of method invocations."),
	)

	op := func(name string) *observation.Operation {
		return observationContext.Operation(observation.Op{
			Name:              fmt.Sprintf("codeintel.autoindexing.store.%s", name),
			MetricLabelValues: []string{name},
			Metrics:           m,
		})
	}

	indexesInsertedCounter := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "src_codeintel_dbstore_indexes_inserted",
		Help: "The number of codeintel index records inserted.",
	})
	observationContext.Registerer.MustRegister(indexesInsertedCounter)

	return &operations{
		// Commits
		getStaleSourcedCommits: op("StaleSourcedCommits"),
		deleteSourcedCommits:   op("DeleteSourcedCommits"),
		updateSourcedCommits:   op("UpdateSourcedCommits"),

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
	}
}
