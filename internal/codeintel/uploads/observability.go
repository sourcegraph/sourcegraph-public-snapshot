package uploads

import (
	"fmt"

	"github.com/sourcegraph/sourcegraph/internal/metrics"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type operations struct {
	list             *observation.Operation
	get              *observation.Operation
	getBatch         *observation.Operation
	enqueue          *observation.Operation
	delete           *observation.Operation
	commitsVisibleTo *observation.Operation
	uploadsVisibleTo *observation.Operation

	staleSourcedCommits  *observation.Operation
	updateSourcedCommits *observation.Operation
	deleteSourcedCommits *observation.Operation

	deleteUploadsWithoutRepository *observation.Operation
	deleteIndexesWithoutRepository *observation.Operation
}

func newOperations(observationContext *observation.Context) *operations {
	metrics := metrics.NewREDMetrics(
		observationContext.Registerer,
		"codeintel_uploads",
		metrics.WithLabels("op"),
		metrics.WithCountHelp("Total number of method invocations."),
	)

	op := func(name string) *observation.Operation {
		return observationContext.Operation(observation.Op{
			Name:              fmt.Sprintf("codeintel.uploads.%s", name),
			MetricLabelValues: []string{name},
			Metrics:           metrics,
		})
	}

	return &operations{
		list:             op("List"),
		get:              op("Get"),
		getBatch:         op("GetBatch"),
		enqueue:          op("Enqueue"),
		delete:           op("Delete"),
		commitsVisibleTo: op("CommitsVisibleTo"),
		uploadsVisibleTo: op("UploadsVisibleTo"),

		staleSourcedCommits:  op("StaleSourcedCommits"),
		updateSourcedCommits: op("UpdateSourcedCommits"),
		deleteSourcedCommits: op("DeleteSourcedCommits"),

		deleteUploadsWithoutRepository: op("DeleteUploadsWithoutRepository"),
		deleteIndexesWithoutRepository: op("DeleteIndexesWithoutRepository"),
	}
}
