package graphql

import (
	"fmt"

	"github.com/sourcegraph/sourcegraph/internal/metrics"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type operations struct {
	rankingSummary         *observation.Operation
	bumpDerivativeGraphKey *observation.Operation
	deleteRankingProgress  *observation.Operation
}

func newOperations(observationCtx *observation.Context) *operations {
	m := metrics.NewREDMetrics(
		observationCtx.Registerer,
		"codeintel_ranking_transport_graphql",
		metrics.WithLabels("op"),
		metrics.WithCountHelp("Total number of method invocations."),
	)

	op := func(name string) *observation.Operation {
		return observationCtx.Operation(observation.Op{
			Name:              fmt.Sprintf("codeintel.ranking.transport.graphql.%s", name),
			MetricLabelValues: []string{name},
			Metrics:           m,
		})
	}

	return &operations{
		rankingSummary:         op("RankingSummary"),
		bumpDerivativeGraphKey: op("BumpDerivativeGraphKey"),
		deleteRankingProgress:  op("DeleteRankingProgress"),
	}
}
