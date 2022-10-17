package inference

import (
	"fmt"

	"github.com/sourcegraph/sourcegraph/internal/metrics"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type operations struct {
	createSandbox              *observation.Operation
	inferIndexJobHints         *observation.Operation
	inferIndexJobs             *observation.Operation
	invokeLinearizedRecognizer *observation.Operation
	invokeRecognizers          *observation.Operation
	resolveFileContents        *observation.Operation
	resolvePaths               *observation.Operation
	setupRecognizers           *observation.Operation
}

func newOperations(observationContext *observation.Context) *operations {
	metrics := metrics.NewREDMetrics(
		observationContext.Registerer,
		"codeintel_autoindexing_inference",
		metrics.WithLabels("op"),
		metrics.WithCountHelp("Total number of method invocations."),
	)

	op := func(name string) *observation.Operation {
		return observationContext.Operation(observation.Op{
			Name:              fmt.Sprintf("codeintel.autoindexing.inference.%s", name),
			MetricLabelValues: []string{name},
			Metrics:           metrics,
		})
	}

	return &operations{
		createSandbox:              op("createSandbox"),
		inferIndexJobHints:         op("InferIndexJobHints"),
		inferIndexJobs:             op("InferIndexJobs"),
		invokeLinearizedRecognizer: op("invokeLinearizedRecognizer"),
		invokeRecognizers:          op("invokeRecognizers"),
		resolveFileContents:        op("resolveFileContents"),
		resolvePaths:               op("resolvePaths"),
		setupRecognizers:           op("setupRecognizers"),
	}
}
