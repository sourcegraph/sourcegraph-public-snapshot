package httpapi

import (
	"fmt"

	"github.com/sourcegraph/sourcegraph/internal/metrics"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type Operations struct {
	authMiddleware *observation.Operation
	serveHTTP      *observation.Operation
}

func NewOperations(observationContext *observation.Context) *Operations {
	m := metrics.NewREDMetrics(
		observationContext.Registerer,
		"batches_httpapi",
		metrics.WithLabels("op"),
		metrics.WithCountHelp("Total number of method invocations."),
	)

	op := func(name string) *observation.Operation {
		return observationContext.Operation(observation.Op{
			Name:              fmt.Sprintf("batches.httpapi.%s", name),
			MetricLabelValues: []string{name},
			Metrics:           m,
		})
	}

	return &Operations{
		authMiddleware: op("authMiddleware"),
		serveHTTP:      op("serveHTTP"),
	}
}
