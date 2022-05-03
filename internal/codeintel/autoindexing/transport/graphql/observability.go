package graphql

import (
	"fmt"

	"github.com/sourcegraph/sourcegraph/internal/metrics"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type operations struct {
	deleteLSIFIndex                    *observation.Operation
	indexConfiguration                 *observation.Operation
	lsifIndexByID                      *observation.Operation
	lsifIndexes                        *observation.Operation
	lsifIndexesByRepo                  *observation.Operation
	queueAutoIndexJobsForRepo          *observation.Operation
	updateRepositoryIndexConfiguration *observation.Operation
}

func newOperations(observationContext *observation.Context) *operations {
	metrics := metrics.NewREDMetrics(
		observationContext.Registerer,
		"codeintel_autoindexing_transport_graphql",
		metrics.WithLabels("op"),
		metrics.WithCountHelp("Total number of method invocations."),
	)

	op := func(name string) *observation.Operation {
		return observationContext.Operation(observation.Op{
			Name:              fmt.Sprintf("codeintel.autoindexing.transport.graphql.%s", name),
			MetricLabelValues: []string{name},
			Metrics:           metrics,
		})
	}

	return &operations{
		deleteLSIFIndex:                    op("DeleteLSIFIndex"),
		indexConfiguration:                 op("IndexConfiguration"),
		lsifIndexByID:                      op("LSIFIndexByID"),
		lsifIndexes:                        op("LSIFIndexes"),
		lsifIndexesByRepo:                  op("LSIFIndexesByRepo"),
		queueAutoIndexJobsForRepo:          op("QueueAutoIndexJobsForRepo"),
		updateRepositoryIndexConfiguration: op("UpdateRepositoryIndexConfiguration"),
	}
}
