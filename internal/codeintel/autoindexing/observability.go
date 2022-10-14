package autoindexing

import (
	"fmt"

	"github.com/sourcegraph/sourcegraph/internal/metrics"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type operations struct {
	// Indexes
	getIndexes                     *observation.Operation
	getIndexByID                   *observation.Operation
	getIndexesByIDs                *observation.Operation
	getRecentIndexesSummary        *observation.Operation
	getLastIndexScanForRepository  *observation.Operation
	deleteIndexByID                *observation.Operation
	deleteIndexes                  *observation.Operation
	deleteIndexesWithoutRepository *observation.Operation
	expireFailedRecords            *observation.Operation
	queueRepoRev                   *observation.Operation
	queueIndex                     *observation.Operation
	queueIndexForPackage           *observation.Operation

	// Commits
	getStaleSourcedCommits *observation.Operation
	updateSourcedCommits   *observation.Operation
	deleteSourcedCommits   *observation.Operation

	// Index Configuration
	getIndexConfigurationByRepositoryID    *observation.Operation
	updateIndexConfigurationByRepositoryID *observation.Operation
	inferIndexConfiguration                *observation.Operation
	setInferenceScript                     *observation.Operation
	getInferenceScript                     *observation.Operation

	// Tags
	getListTags *observation.Operation

	// Language support
	getLanguagesRequestedBy   *observation.Operation
	setRequestLanguageSupport *observation.Operation
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
		// Indexes
		getIndexes:                     op("GetIndexes"),
		getIndexByID:                   op("GetIndexByID"),
		getIndexesByIDs:                op("GetIndexesByIDs"),
		getRecentIndexesSummary:        op("GetRecentIndexesSummary"),
		getLastIndexScanForRepository:  op("GetLastIndexScanForRepository"),
		deleteIndexByID:                op("DeleteIndexByID"),
		deleteIndexes:                  op("DeleteIndexes"),
		deleteIndexesWithoutRepository: op("DeleteIndexesWithoutRepository"),
		expireFailedRecords:            op("ExpireFailedRecords"),
		queueRepoRev:                   op("QueueRepoRev"),
		queueIndex:                     op("QueueIndex"),
		queueIndexForPackage:           op("QueueIndexForPackage"),

		// Commits
		getStaleSourcedCommits: op("GetStaleSourcedCommits"),
		updateSourcedCommits:   op("UpdateSourcedCommits"),
		deleteSourcedCommits:   op("DeleteSourcedCommits"),

		// Index Configuration
		getIndexConfigurationByRepositoryID:    op("GetIndexConfigurationByRepositoryID"),
		updateIndexConfigurationByRepositoryID: op("UpdateIndexConfigurationByRepositoryID"),
		inferIndexConfiguration:                op("InferIndexConfiguration"),
		getInferenceScript:                     op("GetInferenceScript"),
		setInferenceScript:                     op("SetInferenceScript"),

		// Tags
		getListTags: op("GetListTags"),

		// Language support
		getLanguagesRequestedBy:   op("GetLanguagesRequestedBy"),
		setRequestLanguageSupport: op("SetRequestLanguageSupport"),
	}
}
