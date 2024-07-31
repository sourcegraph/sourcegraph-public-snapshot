package graphql

import (
	"context"
	"fmt"
	"time"

	"github.com/sourcegraph/log"
	"go.opentelemetry.io/otel/attribute"

	"github.com/sourcegraph/sourcegraph/internal/metrics"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type operations struct {
	gitBlobLsifData *observation.Operation
	codeGraphData   *observation.Operation
	occurrences     *observation.Operation
	hover           *observation.Operation
	definitions     *observation.Operation
	references      *observation.Operation
	implementations *observation.Operation
	prototypes      *observation.Operation
	diagnostics     *observation.Operation
	stencil         *observation.Operation
	ranges          *observation.Operation
	snapshot        *observation.Operation
	visibleIndexes  *observation.Operation
	usagesForSymbol *observation.Operation
}

func newOperations(observationCtx *observation.Context) *operations {
	m := metrics.NewREDMetrics(
		observationCtx.Registerer,
		"codeintel_codenav_transport_graphql",
		metrics.WithLabels("op"),
		metrics.WithCountHelp("Total number of method invocations."),
	)

	op := func(name string) *observation.Operation {
		return observationCtx.Operation(observation.Op{
			Name:              fmt.Sprintf("codeintel.codenav.transport.graphql.%s", name),
			MetricLabelValues: []string{name},
			Metrics:           m,
		})
	}

	return &operations{
		gitBlobLsifData: op("GitBlobLsifData"),
		codeGraphData:   op("CodeGraphData"),
		occurrences:     op("Occurrences"),
		hover:           op("Hover"),
		definitions:     op("Definitions"),
		references:      op("References"),
		implementations: op("Implementations"),
		prototypes:      op("Prototypes"),
		diagnostics:     op("Diagnostics"),
		stencil:         op("Stencil"),
		ranges:          op("Ranges"),
		snapshot:        op("Snapshot"),
		visibleIndexes:  op("VisibleIndexes"),
		usagesForSymbol: op("UsagesForSymbol"),
	}
}

func observeResolver(
	ctx context.Context,
	err *error,
	operation *observation.Operation,
	threshold time.Duration, //nolint:unparam // same value everywhere but probably want to keep this
	observationArgs observation.Args,
) (context.Context, observation.TraceLogger, func()) { //nolint:unparam // observation.TraceLogger is never used, but it makes sense API wise
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
	fields := []log.Field{log.Duration("duration", duration)}
	if err != nil && *err != nil {
		fields = append(fields, log.Error(*err))
	}
	logger.Warn("Slow codeintel request", fields...)
}

func getObservationArgs[T HasAttrs](args T) observation.Args {
	return observation.Args{Attrs: args.Attrs()}
}

type HasAttrs interface {
	Attrs() []attribute.KeyValue
}
