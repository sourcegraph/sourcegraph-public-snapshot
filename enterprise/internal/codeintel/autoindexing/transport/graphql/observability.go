package graphql

import (
	"fmt"

	"github.com/sourcegraph/sourcegraph/internal/metrics"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type operations struct {
	preciseIndexes                        *observation.Operation
	preciseIndexByID                      *observation.Operation
	deletePreciseIndex                    *observation.Operation
	deletePreciseIndexes                  *observation.Operation
	reindexPreciseIndex                   *observation.Operation
	reindexPreciseIndexes                 *observation.Operation
	indexConfiguration                    *observation.Operation
	updateIndexConfiguration              *observation.Operation
	codeIntelligenceInferenceScript       *observation.Operation
	updateCodeIntelligenceInferenceScript *observation.Operation
	summary                               *observation.Operation
	repositorySummary                     *observation.Operation
	getRecentIndexesSummary               *observation.Operation
	getLastIndexScanForRepository         *observation.Operation
	gitBlobCodeIntelInfo                  *observation.Operation
	inferAutoIndexJobsForRepo             *observation.Operation
	queueAutoIndexJobsForRepo             *observation.Operation
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
		preciseIndexes:                        op("PreciseIndexes"),
		preciseIndexByID:                      op("PreciseIndexByID"),
		deletePreciseIndex:                    op("DeletePreciseIndex"),
		deletePreciseIndexes:                  op("DeletePreciseIndexes"),
		reindexPreciseIndex:                   op("ReindexPreciseIndex"),
		reindexPreciseIndexes:                 op("ReindexPreciseIndexes"),
		indexConfiguration:                    op("IndexConfiguration"),
		updateIndexConfiguration:              op("UpdateIndexConfiguration"),
		codeIntelligenceInferenceScript:       op("CodeIntelligenceInferenceScript"),
		updateCodeIntelligenceInferenceScript: op("UpdateCodeIntelligenceInferenceScript"),
		summary:                               op("Summary"),
		repositorySummary:                     op("RepositorySummary"),
		getRecentIndexesSummary:               op("GetRecentIndexesSummary"),
		getLastIndexScanForRepository:         op("GetLastIndexScanForRepository"),
		gitBlobCodeIntelInfo:                  op("GitBlobCodeIntelInfo"),
		inferAutoIndexJobsForRepo:             op("InferAutoIndexJobsForRepo"),
		queueAutoIndexJobsForRepo:             op("QueueAutoIndexJobsForRepo"),
	}
}
