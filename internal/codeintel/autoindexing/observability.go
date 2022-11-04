package autoindexing

import (
	"fmt"

	"github.com/sourcegraph/sourcegraph/internal/metrics"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type operations struct {
	// Commits
	getStaleSourcedCommits *observation.Operation
	updateSourcedCommits   *observation.Operation
	deleteSourcedCommits   *observation.Operation

	// Indexes
	getIndexes                     *observation.Operation
	getIndexByID                   *observation.Operation
	getIndexesByIDs                *observation.Operation
	getRecentIndexesSummary        *observation.Operation
	getLastIndexScanForRepository  *observation.Operation
	deleteIndexByID                *observation.Operation
	deleteIndexesWithoutRepository *observation.Operation
	queueIndex                     *observation.Operation
	queueIndexForPackage           *observation.Operation

	// Index Configuration
	getIndexConfigurationByRepositoryID    *observation.Operation
	updateIndexConfigurationByRepositoryID *observation.Operation
	inferIndexConfiguration                *observation.Operation
}

func newOperations(observationContext *observation.Context) *operations {
	m := metrics.NewREDMetrics(
		observationContext.Registerer,
		"codeintel_autoindexing",
		metrics.WithLabels("op"),
		metrics.WithCountHelp("Total number of method invocations."),
	)

	op := func(name string) *observation.Operation {
		return observationContext.Operation(observation.Op{
			Name:              fmt.Sprintf("codeintel.autoindexing.%s", name),
			MetricLabelValues: []string{name},
			Metrics:           m,
		})
	}

	return &operations{
		// Commits
		getStaleSourcedCommits: op("GetStaleSourcedCommits"),
		updateSourcedCommits:   op("UpdateSourcedCommits"),
		deleteSourcedCommits:   op("DeleteSourcedCommits"),

		// Indexes
		getIndexes:                     op("GetIndexes"),
		getIndexByID:                   op("GetIndexByID"),
		getIndexesByIDs:                op("GetIndexesByIDs"),
		getRecentIndexesSummary:        op("GetRecentIndexesSummary"),
		getLastIndexScanForRepository:  op("GetLastIndexScanForRepository"),
		deleteIndexByID:                op("DeleteIndexByID"),
		deleteIndexesWithoutRepository: op("DeleteIndexesWithoutRepository"),
		queueIndex:                     op("QueueIndex"),
		queueIndexForPackage:           op("QueueIndexForPackage"),

		// Index Configuration
		getIndexConfigurationByRepositoryID:    op("GetIndexConfigurationByRepositoryID"),
		updateIndexConfigurationByRepositoryID: op("UpdateIndexConfigurationByRepositoryID"),
		inferIndexConfiguration:                op("InferIndexConfiguration"),
	}
}
