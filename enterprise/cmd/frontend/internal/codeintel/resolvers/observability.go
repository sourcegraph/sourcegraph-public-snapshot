package resolvers

import (
	"context"
	"fmt"
	"time"

	"github.com/inconshreveable/log15"

	"github.com/sourcegraph/sourcegraph/internal/honey"
	"github.com/sourcegraph/sourcegraph/internal/metrics"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/trace"
)

type operations struct {
	queryResolver *observation.Operation
	definitions   *observation.Operation
	diagnostics   *observation.Operation
	hover         *observation.Operation
	ranges        *observation.Operation
	references    *observation.Operation

	findClosestDumps *observation.Operation
}

func newOperations(observationContext *observation.Context) *operations {
	metrics := metrics.NewOperationMetrics(
		observationContext.Registerer,
		"codeintel_resolvers",
		metrics.WithLabels("op"),
		metrics.WithCountHelp("Total number of resolver invocations."),
	)

	op := func(name string) *observation.Operation {
		return observationContext.Operation(observation.Op{
			Name:         fmt.Sprintf("codeintel.resolvers.%s", name),
			MetricLabels: []string{name},
			Metrics:      metrics,
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
		queryResolver: op("QueryResolver"),
		definitions:   op("Definitions"),
		diagnostics:   op("Diagnostics"),
		hover:         op("Hover"),
		ranges:        op("Ranges"),
		references:    op("References"),

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
	ctx, traceLog, endObservation := operation.WithAndLogger(ctx, err, observationArgs)

	return ctx, traceLog, func() {
		duration := time.Since(start)
		endObservation(1, observation.Args{})

		if duration >= threshold {
			lowSlowRequest(name, duration, err, observationArgs)
		}
		if honey.Enabled() {
			_ = honey.EventWithFields("codeintel", codeintelHoneyEventFields(ctx, err, observationArgs, map[string]interface{}{
				"type":        name,
				"duration_ms": duration.Milliseconds(),
			}))
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

func codeintelHoneyEventFields(ctx context.Context, err *error, observationArgs observation.Args, extra map[string]interface{}) map[string]interface{} {
	fields := observationArgs.LogFieldMap()
	if err != nil && *err != nil {
		fields["error"] = (*err).Error()
	}
	if spanURL := trace.SpanURLFromContext(ctx); spanURL != "" {
		fields["trace"] = spanURL
	}
	for key, value := range extra {
		fields[key] = value
	}

	return fields
}
