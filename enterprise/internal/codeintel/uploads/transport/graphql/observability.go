package graphql

import (
	"fmt"

	"github.com/sourcegraph/sourcegraph/internal/metrics"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type operations struct {
	codeIntelSummary      *observation.Operation
	commitGraph           *observation.Operation
	deletePreciseIndex    *observation.Operation
	deletePreciseIndexes  *observation.Operation
	preciseIndexByID      *observation.Operation
	preciseIndexes        *observation.Operation
	reindexPreciseIndex   *observation.Operation
	reindexPreciseIndexes *observation.Operation
	repositorySummary     *observation.Operation
}

func newOperations(observationCtx *observation.Context) *operations {
	m := metrics.NewREDMetrics(
		observationCtx.Registerer,
		"codeintel_uploads_transport_graphql",
		metrics.WithLabels("op"),
		metrics.WithCountHelp("Total number of method invocations."),
	)

	op := func(name string) *observation.Operation {
		return observationCtx.Operation(observation.Op{
			Name:              fmt.Sprintf("codeintel.uploads.transport.graphql.%s", name),
			MetricLabelValues: []string{name},
			Metrics:           m,
		})
	}

	return &operations{
		codeIntelSummary:      op("CodeIntelSummary"),
		commitGraph:           op("CommitGraph"),
		deletePreciseIndex:    op("DeletePreciseIndex"),
		deletePreciseIndexes:  op("DeletePreciseIndexes"),
		preciseIndexByID:      op("PreciseIndexByID"),
		preciseIndexes:        op("PreciseIndexes"),
		reindexPreciseIndex:   op("ReindexPreciseIndex"),
		reindexPreciseIndexes: op("ReindexPreciseIndexes"),
		repositorySummary:     op("RepositorySummary"),
	}
}
