package luasandbox

import (
	"fmt"

	"github.com/sourcegraph/sourcegraph/internal/metrics"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type operations struct {
	createSandbox *observation.Operation
	runScript     *observation.Operation
	call          *observation.Operation
	callGenerator *observation.Operation
}

func newOperations(observationContext *observation.Context) *operations {
	metrics := metrics.NewREDMetrics(
		observationContext.Registerer,
		"luasandbox",
		metrics.WithLabels("op"),
		metrics.WithCountHelp("Total number of method invocations."),
	)

	op := func(name string) *observation.Operation {
		return observationContext.Operation(observation.Op{
			Name:              fmt.Sprintf("luasandbox.%s", name),
			MetricLabelValues: []string{name},
			Metrics:           metrics,
		})
	}

	return &operations{
		createSandbox: op("CreateSandbox"),
		runScript:     op("RunScript"),
		call:          op("Call"),
		callGenerator: op("CallGenerator"),
	}
}
