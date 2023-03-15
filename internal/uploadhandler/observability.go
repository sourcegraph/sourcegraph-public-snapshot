package uploadhandler

import (
	"fmt"
	"syscall"

	"github.com/sourcegraph/sourcegraph/internal/metrics"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type Operations struct {
	handleEnqueue                  *observation.Operation
	handleEnqueueSinglePayload     *observation.Operation
	handleEnqueueMultipartSetup    *observation.Operation
	handleEnqueueMultipartUpload   *observation.Operation
	handleEnqueueMultipartFinalize *observation.Operation
}

func NewOperations(observationCtx *observation.Context, prefix string) *Operations {
	redMetrics := metrics.NewREDMetrics(
		observationCtx.Registerer,
		fmt.Sprintf("%s_uploadhandler", prefix),
		metrics.WithLabels("op"),
		metrics.WithCountHelp("Total number of method invocations."),
	)

	op := func(name string) *observation.Operation {
		return observationCtx.Operation(observation.Op{
			Name:              fmt.Sprintf("%s.uploadhandler.%s", prefix, name),
			MetricLabelValues: []string{name},
			Metrics:           redMetrics,
			ErrorFilter: func(err error) observation.ErrorFilterBehaviour {
				var errno syscall.Errno
				if errors.As(err, &errno) && errno == syscall.ECONNREFUSED {
					return observation.EmitForDefault ^ observation.EmitForSentry
				}
				return observation.EmitForDefault
			},
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
