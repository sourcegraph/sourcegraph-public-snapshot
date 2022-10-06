package ranking

import (
	"fmt"

	"github.com/sourcegraph/sourcegraph/internal/metrics"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type operations struct {
	getRepoRank       *observation.Operation
	getDocumentRanks  *observation.Operation
	indexRepositories *observation.Operation
}

func newOperations(observationContext *observation.Context) *operations {
	m := metrics.NewREDMetrics(
		observationContext.Registerer,
		"codeintel_ranking",
		metrics.WithLabels("op"),
		metrics.WithCountHelp("Total number of method invocations."),
	)

	op := func(name string) *observation.Operation {
		return observationContext.Operation(observation.Op{
			Name:              fmt.Sprintf("codeintel.ranking.%s", name),
			MetricLabelValues: []string{name},
			Metrics:           m,
		})
	}

	return &operations{
		getRepoRank:       op("GetRepoRank"),
		getDocumentRanks:  op("GetDocumentRanks"),
		indexRepositories: op("IndexRepositories"),
	}
}
