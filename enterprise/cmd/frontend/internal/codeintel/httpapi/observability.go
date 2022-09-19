package httpapi

import (
	"fmt"

	generic "github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/codeintel/httpapi/generic"
	"github.com/sourcegraph/sourcegraph/internal/metrics"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type Operations struct {
	*generic.Operations
	authMiddleware *observation.Operation
}

func NewOperations(observationContext *observation.Context) *Operations {
	metrics := metrics.NewREDMetrics(
		observationContext.Registerer,
		"codeintel_httpapi_auth",
		metrics.WithLabels("op"),
		metrics.WithCountHelp("Total number of method invocations."),
	)

	op := func(name string) *observation.Operation {
		return observationContext.Operation(observation.Op{
			Name:              fmt.Sprintf("codeintel.httpapi.auth.%s", name),
			MetricLabelValues: []string{name},
			Metrics:           metrics,
		})
	}

	return &Operations{
		Operations:     generic.NewOperations(observationContext),
		authMiddleware: op("authMiddleware"),
	}
}
