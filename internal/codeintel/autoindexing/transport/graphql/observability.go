package graphql

import (
	"fmt"

	"github.com/sourcegraph/sourcegraph/internal/metrics"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type operations struct {
	// Indexes
	getIndexByID                  *observation.Operation
	getIndexesByIDs               *observation.Operation
	getRecentIndexesSummary       *observation.Operation
	getLastIndexScanForRepository *observation.Operation
	deleteIndexByID               *observation.Operation
	queueAutoIndexJobsForRepo     *observation.Operation

	// Index Configuration
	getIndexConfiguration                  *observation.Operation
	updateIndexConfigurationByRepositoryID *observation.Operation
	inferedIndexConfiguration              *observation.Operation
	inferedIndexConfigurationHints         *observation.Operation
}

func newOperations(observationContext *observation.Context) *operations {
	m := metrics.NewREDMetrics(
		observationContext.Registerer,
		"codeintel_autoindexing_transport_graphql",
		metrics.WithLabels("op"),
		metrics.WithCountHelp("Total number of method invocations."),
	)

	op := func(name string) *observation.Operation {
		return observationContext.Operation(observation.Op{
			Name:              fmt.Sprintf("codeintel.autoindexing.transport.graphql.%s", name),
			MetricLabelValues: []string{name},
			Metrics:           m,
		})
	}

	return &operations{
		// Indexes
		getIndexByID:                  op("GetIndexByID"),
		getIndexesByIDs:               op("GetIndexesByIDs"),
		getRecentIndexesSummary:       op("GetRecentIndexesSummary"),
		getLastIndexScanForRepository: op("GetLastIndexScanForRepository"),
		deleteIndexByID:               op("DeleteIndexByID"),
		queueAutoIndexJobsForRepo:     op("QueueAutoIndexJobsForRepo"),

		// Index Configuration
		getIndexConfiguration:                  op("IndexConfiguration"),
		updateIndexConfigurationByRepositoryID: op("UpdateIndexConfigurationByRepositoryID"),
		inferedIndexConfiguration:              op("InferedIndexConfiguration"),
		inferedIndexConfigurationHints:         op("InferedIndexConfigurationHints"),
	}
}
