package coursier

import (
	"fmt"
	"strings"

	"github.com/sourcegraph/sourcegraph/internal/metrics"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type Operations struct {
	fetchSources  *observation.Operation
	exists        *observation.Operation
	fetchByteCode *observation.Operation
	runCommand    *observation.Operation
}

func NewOperationsFromMetrics(observationContext *observation.Context) *Operations {
	metrics := metrics.NewOperationMetrics(
		observationContext.Registerer,
		"codeintel_coursier",
		metrics.WithLabels("op"),
		metrics.WithCountHelp("Total number of method invocations."),
	)

	op := func(name string) *observation.Operation {
		return observationContext.Operation(observation.Op{
			Name:              fmt.Sprintf("codeintel.coursier.%s", name),
			MetricLabelValues: []string{name},
			Metrics:           metrics,
			ErrorFilter: func(err error) observation.ErrorFilterBehaviour {
				if err != nil && strings.Contains(err.Error(), "not found") {
					return observation.EmitForMetrics | observation.EmitForTraces
				}
				return observation.EmitForAll
			},
		})
	}

	return &Operations{
		fetchSources:  op("FetchSources"),
		exists:        op("Exists"),
		fetchByteCode: op("FetchByteCode"),
		runCommand:    op("RunCommand"),
	}
}
