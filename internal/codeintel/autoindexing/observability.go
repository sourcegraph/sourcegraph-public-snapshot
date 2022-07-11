package autoindexing

import (
	"fmt"

	"github.com/sourcegraph/sourcegraph/internal/metrics"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type operations struct {
	// Not used yet.
	delete                      *observation.Operation
	enqueue                     *observation.Operation
	get                         *observation.Operation
	getBatch                    *observation.Operation
	infer                       *observation.Operation
	list                        *observation.Operation
	updateIndexingConfiguration *observation.Operation
	inferIndexConfiguration     *observation.Operation // temporary
	queueIndex                  *observation.Operation // temporary
	queueIndexForPackage        *observation.Operation // temporary

	// Commits
	getStaleSourcedCommits *observation.Operation
	updateSourcedCommits   *observation.Operation
	deleteSourcedCommits   *observation.Operation

	// Indexes
	deleteIndexesWithoutRepository *observation.Operation
}

func newOperations(observationContext *observation.Context) *operations {
	metrics := metrics.NewREDMetrics(
		observationContext.Registerer,
		"codeintel_autoindexing",
		metrics.WithLabels("op"),
		metrics.WithCountHelp("Total number of method invocations."),
	)

	op := func(name string) *observation.Operation {
		return observationContext.Operation(observation.Op{
			Name:              fmt.Sprintf("codeintel.autoindexing.%s", name),
			MetricLabelValues: []string{name},
			Metrics:           metrics,
		})
	}

	return &operations{
		// Not used yet.
		delete:                      op("Delete"),
		enqueue:                     op("Enqueue"),
		get:                         op("Get"),
		getBatch:                    op("GetBatch"),
		infer:                       op("Infer"),
		list:                        op("List"),
		updateIndexingConfiguration: op("UpdateIndexingConfiguration"),
		inferIndexConfiguration:     op("InferIndexConfiguration"), // temporary
		queueIndex:                  op("QueueIndex"),              // temporary
		queueIndexForPackage:        op("QueueIndexForPackage"),    // temporary

		// Commits
		getStaleSourcedCommits: op("GetStaleSourcedCommits"),
		updateSourcedCommits:   op("UpdateSourcedCommits"),
		deleteSourcedCommits:   op("DeleteSourcedCommits"),

		// Indexes
		deleteIndexesWithoutRepository: op("DeleteIndexesWithoutRepository"),
	}
}
