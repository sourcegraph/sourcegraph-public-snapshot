package npm

import (
	"fmt"
	"strings"

	"github.com/sourcegraph/sourcegraph/internal/metrics"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type Operations struct {
	fetchSources *observation.Operation
	exists       *observation.Operation
	runCommand   *observation.Operation
}

func NewOperations(observationContext *observation.Context) *Operations {
	redMetrics := metrics.NewREDMetrics(
		observationContext.Registerer,
		"codeintel_npm",
		metrics.WithLabels("op"),
		metrics.WithCountHelp("Total number of method invocations."),
	)

	op := func(name string) *observation.Operation {
		return observationContext.Operation(observation.Op{
			Name:              fmt.Sprintf("codeintel.npm.%s", name),
			MetricLabelValues: []string{name},
			Metrics:           redMetrics,
			ErrorFilter: func(err error) observation.ErrorFilterBehaviour {
				if err != nil && strings.Contains(err.Error(), "not found") {
					return observation.EmitForMetrics | observation.EmitForTraces
				}
				return observation.EmitForDefault
			},
		})
	}

	return &Operations{
		fetchSources: op("FetchSources"),
		exists:       op("Exists"),
		runCommand:   op("RunCommand"),
	}
}
