package httpapi

import (
	"fmt"

	"github.com/sourcegraph/sourcegraph/internal/metrics"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type Operations struct {
	authMiddleware                 *observation.Operation
	handleEnqueue                  *observation.Operation
	handleEnqueueSinglePayload     *observation.Operation
	handleEnqueueMultipartSetup    *observation.Operation
	handleEnqueueMultipartUpload   *observation.Operation
	handleEnqueueMultipartFinalize *observation.Operation
}

func NewOperations(observationContext *observation.Context) *Operations {
	metrics := metrics.NewREDMetrics(
		observationContext.Registerer,
		"codeintel_httpapi",
		metrics.WithLabels("op"),
		metrics.WithCountHelp("Total number of method invocations."),
	)

	op := func(name string) *observation.Operation {
		return observationContext.Operation(observation.Op{
			Name:              fmt.Sprintf("codeintel.httpapi.%s", name),
			MetricLabelValues: []string{name},
			Metrics:           metrics,
		})
	}

	return &Operations{
		authMiddleware:                 op("authMiddleware"),
		handleEnqueue:                  op("HandleEnqueue"),
		handleEnqueueSinglePayload:     op("handleEnqueueSinglePayload"),
		handleEnqueueMultipartSetup:    op("handleEnqueueMultipartSetup"),
		handleEnqueueMultipartUpload:   op("handleEnqueueMultipartUpload"),
		handleEnqueueMultipartFinalize: op("handleEnqueueMultipartFinalize"),
	}
}
