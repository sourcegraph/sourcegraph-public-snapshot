package resolvers

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"

	"github.com/sourcegraph/sourcegraph/internal/metrics"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/lib/log"
)

type operations struct {
	definitions     *observation.Operation
	diagnostics     *observation.Operation
	hover           *observation.Operation
	queryResolver   *observation.Operation
	ranges          *observation.Operation
	references      *observation.Operation
	implementations *observation.Operation
	stencil         *observation.Operation

	findClosestDumps *observation.Operation
}

func newOperations(observationContext *observation.Context) *operations {
	metrics := metrics.NewREDMetrics(
		observationContext.Registerer,
		"codeintel_resolvers",
		metrics.WithLabels("op"),
		metrics.WithCountHelp("Total number of resolver invocations."),
	)

	op := func(name string) *observation.Operation {
		return observationContext.Operation(observation.Op{
			Name:              fmt.Sprintf("codeintel.resolvers.%s", name),
			MetricLabelValues: []string{name},
			Metrics:           metrics,
			ErrorFilter: func(err error) observation.ErrorFilterBehaviour {
				return observation.EmitForSentry | observation.EmitForDefault
			},
		})
	}

	// suboperations do not have their own metrics but do have their
	// own opentracing spans. This allows us to more granularly track
	// the latency for parts of a request without noising up Prometheus.
	subOp := func(name string) *observation.Operation {
		return observationContext.Operation(observation.Op{
			Name: fmt.Sprintf("codeintel.resolvers.%s", name),
		})
	}

	return &operations{
		definitions:     op("Definitions"),
		diagnostics:     op("Diagnostics"),
		hover:           op("Hover"),
		implementations: op("Implementations"),
		ranges:          op("Ranges"),
		references:      op("References"),
		stencil:         op("Stencil"),
		queryResolver:   op("QueryResolver"),

		findClosestDumps: subOp("findClosestDumps"),
	}
}

func observeResolver(
	ctx context.Context,
	err *error,
	operation *observation.Operation,
	threshold time.Duration,
	observationArgs observation.Args,
) (context.Context, observation.TraceLogger, func()) {
	start := time.Now()
	ctx, trace, endObservation := operation.With(ctx, err, observationArgs)

	return ctx, trace, func() {
		duration := time.Since(start)
		endObservation(1, observation.Args{})

		if duration >= threshold {
			// use trace logger which includes all relevant fields
			lowSlowRequest(trace, duration, err)
		}
	}
}

func lowSlowRequest(logger log.Logger, duration time.Duration, err *error) {
	fields := []log.Field{zap.Duration("duration", duration)}
	if err != nil && *err != nil {
		fields = append(fields, log.Error(*err))
	}
	logger.Warn("Slow codeintel request", fields...)
}
