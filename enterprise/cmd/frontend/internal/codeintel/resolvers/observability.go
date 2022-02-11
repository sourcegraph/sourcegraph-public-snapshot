package resolvers

import (
	"context"
	"fmt"
	"time"

	"github.com/inconshreveable/log15"

	"github.com/sourcegraph/sourcegraph/internal/metrics"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type operations struct {
	definitions               *observation.Operation
	diagnostics               *observation.Operation
	documentation             *observation.Operation
	documentationIDsToPathIDs *observation.Operation
	documentationPage         *observation.Operation
	documentationPathInfo     *observation.Operation
	documentationReferences   *observation.Operation
	documentationSearch       *observation.Operation
	hover                     *observation.Operation
	queryResolver             *observation.Operation
	ranges                    *observation.Operation
	references                *observation.Operation
	implementations           *observation.Operation
	stencil                   *observation.Operation

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
		definitions:               op("Definitions"),
		diagnostics:               op("Diagnostics"),
		documentation:             op("Documentation"),
		documentationIDsToPathIDs: op("DocumentationIDsToPathIDs"),
		documentationPage:         op("DocumentationPage"),
		documentationPathInfo:     op("DocumentationPathInfo"),
		documentationReferences:   op("DocumentationReferences"),
		documentationSearch:       op("DocumentationSearch"),
		hover:                     op("Hover"),
		implementations:           op("Implementations"),
		ranges:                    op("Ranges"),
		references:                op("References"),
		stencil:                   op("Stencil"),
		queryResolver:             op("QueryResolver"),

		findClosestDumps: subOp("findClosestDumps"),
	}
}

func observeResolver(
	ctx context.Context,
	err *error,
	name string,
	operation *observation.Operation,
	threshold time.Duration,
	observationArgs observation.Args,
) (context.Context, observation.TraceLogger, func()) {
	start := time.Now()
	ctx, trace, endObservation := operation.WithAndLogger(ctx, err, observationArgs)

	return ctx, trace, func() {
		duration := time.Since(start)
		endObservation(1, observation.Args{})

		if duration >= threshold {
			lowSlowRequest(name, duration, err, observationArgs)
		}
	}
}

func lowSlowRequest(name string, duration time.Duration, err *error, observationArgs observation.Args) {
	pairs := append(
		observationArgs.LogFieldPairs(),
		"type", name,
		"duration_ms", duration.Milliseconds(),
	)
	if err != nil && *err != nil {
		pairs = append(pairs, "error", (*err).Error())
	}

	log15.Warn("Slow codeintel request", pairs...)
}
