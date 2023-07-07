package httpapi

import (
	"fmt"

	"github.com/sourcegraph/sourcegraph/internal/metrics"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type Operations struct {
	upload *observation.Operation
}

func NewOperations(observationCtx *observation.Context) *Operations {
	m := metrics.NewREDMetrics(
		observationCtx.Registerer,
		"embeddings_httpapi",
		metrics.WithLabels("op"),
		metrics.WithCountHelp("Total number of method invocations."),
	)

	op := func(name string) *observation.Operation {
		return observationCtx.Operation(observation.Op{
			Name:              fmt.Sprintf("embeddings.httpapi.%s", name),
			MetricLabelValues: []string{name},
			Metrics:           m,
		})
	}

	return &Operations{
		upload: op("upload"),
	}
}
