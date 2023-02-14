package graphql

import (
	"fmt"

	"github.com/sourcegraph/sourcegraph/internal/metrics"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type operations struct {
	// Indexes
	getRecentIndexesSummary       *observation.Operation
	getLastIndexScanForRepository *observation.Operation
	deleteLsifIndex               *observation.Operation
	deleteLsifIndexes             *observation.Operation
	reindexLsifIndex              *observation.Operation
	reindexLsifIndexes            *observation.Operation
	queueAutoIndexJobsForRepo     *observation.Operation
	lsifIndexByID                 *observation.Operation
	lsifIndexes                   *observation.Operation
	lsifIndexesByRepo             *observation.Operation
	indexConfiguration            *observation.Operation
	updateIndexConfiguration      *observation.Operation

	// Index Configuration
	inferedIndexConfiguration      *observation.Operation
	inferedIndexConfigurationHints *observation.Operation

	// Language Support
	requestLanguageSupport    *observation.Operation
	requestedLanguageSupport  *observation.Operation
	setRequestLanguageSupport *observation.Operation

	// Misc
	summary              *observation.Operation
	repositorySummary    *observation.Operation
	getSupportedByCtags  *observation.Operation
	gitBlobCodeIntelInfo *observation.Operation

	preciseIndexes        *observation.Operation
	preciseIndexByID      *observation.Operation
	deletePreciseIndex    *observation.Operation
	deletePreciseIndexes  *observation.Operation
	reindexPreciseIndex   *observation.Operation
	reindexPreciseIndexes *observation.Operation
}

func newOperations(observationCtx *observation.Context) *operations {
	m := metrics.NewREDMetrics(
		observationCtx.Registerer,
		"codeintel_autoindexing_transport_graphql",
		metrics.WithLabels("op"),
		metrics.WithCountHelp("Total number of method invocations."),
	)

	op := func(name string) *observation.Operation {
		return observationCtx.Operation(observation.Op{
			Name:              fmt.Sprintf("codeintel.autoindexing.transport.graphql.%s", name),
			MetricLabelValues: []string{name},
			Metrics:           m,
		})
	}

	return &operations{
		// Indexes
		getRecentIndexesSummary:       op("GetRecentIndexesSummary"),
		getLastIndexScanForRepository: op("GetLastIndexScanForRepository"),
		queueAutoIndexJobsForRepo:     op("QueueAutoIndexJobsForRepo"),
		deleteLsifIndex:               op("DeleteLsifIndex"),
		deleteLsifIndexes:             op("DeleteLsifIndexes"),
		reindexLsifIndex:              op("ReindexLsifIndex"),
		reindexLsifIndexes:            op("ReindexLsifIndexes"),
		lsifIndexByID:                 op("LsifIndexByID"),
		lsifIndexes:                   op("LsifIndexes"),
		lsifIndexesByRepo:             op("LsifIndexesByRepo"),
		indexConfiguration:            op("IndexConfiguration"),
		updateIndexConfiguration:      op("UpdateIndexConfiguration"),

		// Index Configuration
		inferedIndexConfiguration:      op("InferedIndexConfiguration"),
		inferedIndexConfigurationHints: op("InferedIndexConfigurationHints"),

		// Language Support
		requestLanguageSupport:    op("RequestLanguageSupport"),
		requestedLanguageSupport:  op("RequestedLanguageSupport"),
		setRequestLanguageSupport: op("SetRequestLanguageSupport"),

		// Misc
		summary:              op("Summary"),
		repositorySummary:    op("RepositorySummary"),
		getSupportedByCtags:  op("GetSupportedByCtags"),
		gitBlobCodeIntelInfo: op("GitBlobCodeIntelInfo"),

		preciseIndexes:        op("PreciseIndexes"),
		preciseIndexByID:      op("PreciseIndexByID"),
		deletePreciseIndex:    op("DeletePreciseIndex"),
		deletePreciseIndexes:  op("DeletePreciseIndexes"),
		reindexPreciseIndex:   op("ReindexPreciseIndex"),
		reindexPreciseIndexes: op("ReindexPreciseIndexes"),
	}
}
