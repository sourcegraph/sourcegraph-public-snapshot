package background

import (
	"fmt"

	"github.com/sourcegraph/sourcegraph/internal/metrics"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type operations struct {
	handleRankingGraphSerializer *observation.Operation
}

func newOperations(observationContext *observation.Context) *operations {
	m := metrics.NewREDMetrics(
		observationContext.Registerer,
		"codeintel_codenav_background",
		metrics.WithLabels("op"),
		metrics.WithCountHelp("Total number of method invocations."),
	)

	op := func(name string) *observation.Operation {
		return observationContext.Operation(observation.Op{
			Name:              fmt.Sprintf("codeintel.codenav.background.%s", name),
			MetricLabelValues: []string{name},
			Metrics:           m,
		})
	}

	handleRankingGraphSerializer := observationContext.Operation(observation.Op{
		Name:              "codeintel.codenav.HandleRankingGraphSerializer",
		MetricLabelValues: []string{"HandleRankingGraphSerializer"},
		Metrics:           m,
	})

	_ = op
	return &operations{
		handleRankingGraphSerializer: handleRankingGraphSerializer,
	}
}
