package uploadhandler

import (
	"fmt"

	"github.com/sourcegraph/sourcegraph/internal/metrics"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type Operations struct {
	handleEnqueue                  *observation.Operation
	handleEnqueueSinglePayload     *observation.Operation
	handleEnqueueMultipartSetup    *observation.Operation
	handleEnqueueMultipartUpload   *observation.Operation
	handleEnqueueMultipartFinalize *observation.Operation
}

func NewOperations(observationContext *observation.Context) *Operations {
	metrics := metrics.NewREDMetrics(
		observationContext.Registerer,
		"uploadhandler",
		metrics.WithLabels("op"),
		metrics.WithCountHelp("Total number of method invocations."),
	)

	op := func(name string) *observation.Operation {
		return observationContext.Operation(observation.Op{
			Name:              fmt.Sprintf("uploadhandler.%s", name),
			MetricLabelValues: []string{name},
			Metrics:           metrics,
		})
	}

	return &Operations{
		handleEnqueue:                  op("HandleEnqueue"),
		handleEnqueueSinglePayload:     op("handleEnqueueSinglePayload"),
		handleEnqueueMultipartSetup:    op("handleEnqueueMultipartSetup"),
		handleEnqueueMultipartUpload:   op("handleEnqueueMultipartUpload"),
		handleEnqueueMultipartFinalize: op("handleEnqueueMultipartFinalize"),
	}
}
