package graphql

import (
	"fmt"

	"github.com/sourcegraph/sourcegraph/internal/metrics"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type operations struct {
	codeIntelligenceInferenceScript       *observation.Operation
	indexConfiguration                    *observation.Operation
	inferAutoIndexJobsForRepo             *observation.Operation
	queueAutoIndexJobsForRepo             *observation.Operation
	updateCodeIntelligenceInferenceScript *observation.Operation
	updateRepositoryIndexConfiguration    *observation.Operation
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
		codeIntelligenceInferenceScript:       op("CodeIntelligenceInferenceScript"),
		indexConfiguration:                    op("IndexConfiguration"),
		inferAutoIndexJobsForRepo:             op("InferAutoIndexJobsForRepo"),
		queueAutoIndexJobsForRepo:             op("QueueAutoIndexJobsForRepo"),
		updateCodeIntelligenceInferenceScript: op("UpdateCodeIntelligenceInferenceScript"),
		updateRepositoryIndexConfiguration:    op("UpdateRepositoryIndexConfiguration"),
	}
}
