package luasandbox

import (
	"fmt"

	"github.com/sourcegraph/sourcegraph/internal/metrics"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type operations struct {
	call           *observation.Operation
	callGenerator  *observation.Operation
	createSandbox  *observation.Operation
	runGoCallback  *observation.Operation
	runScript      *observation.Operation
	runScriptNamed *observation.Operation
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
		call:           op("Call"),
		callGenerator:  op("CallGenerator"),
		createSandbox:  op("CreateSandbox"),
		runGoCallback:  op("RunGoCallback"),
		runScript:      op("RunScript"),
		runScriptNamed: op("RunScriptNamed"),
	}
}
