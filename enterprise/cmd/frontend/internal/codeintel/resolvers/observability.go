package resolvers

import (
	"context"
	"fmt"
	"time"

	"github.com/inconshreveable/log15"

	"github.com/sourcegraph/sourcegraph/internal/honey"
	"github.com/sourcegraph/sourcegraph/internal/metrics"
	"github.com/sourcegraph/sourcegraph/internal/observation"
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
			sendHoneyEvent(name, duration, err, observationArgs)
		}
	}
}

func lowSlowRequest(name string, duration time.Duration, err *error, observationArgs observation.Args) {
	pairs := append(
		logFieldToPairs(observationArgs),
		"type", name,
		"duration_ms", duration.Milliseconds(),
	)
	if err != nil && *err != nil {
		pairs = append(pairs, "error", (*err).Error())
	}

	log15.Warn("Slow codeintel request", pairs...)
}

func logFieldToPairs(observationArgs observation.Args) []interface{} {
	pairs := make([]interface{}, 0, len(observationArgs.LogFields)*2)
	for _, field := range observationArgs.LogFields {
		pairs = append(pairs, field.Key(), field.Value())
	}

	return pairs
}

func sendHoneyEvent(name string, duration time.Duration, err *error, observationArgs observation.Args) {
	ev := honey.Event("codeintel")
	for _, field := range observationArgs.LogFields {
		ev.AddField(field.Key(), field.Value())
	}
	ev.AddField("type", name)
	ev.AddField("duration_ms", duration.Milliseconds())
	if err != nil && *err != nil {
		ev.AddField("error", (*err).Error())
	}

	_ = ev.Send()
}
