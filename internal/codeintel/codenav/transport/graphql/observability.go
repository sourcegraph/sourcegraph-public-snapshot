package graphql

import (
	"context"
	"fmt"
	"time"

	traceLog "github.com/opentracing/opentracing-go/log"
	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/codenav/shared"
	"github.com/sourcegraph/sourcegraph/internal/metrics"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type operations struct {
	hover           *observation.Operation
	definitions     *observation.Operation
	references      *observation.Operation
	implementations *observation.Operation
	diagnostics     *observation.Operation
	stencil         *observation.Operation
	ranges          *observation.Operation

	getSupportedByCtags        *observation.Operation
	getGitBlobLSIFDataResolver *observation.Operation
	getLanguagesRequestedBy    *observation.Operation
	setRequestLanguageSupport  *observation.Operation
}

func newOperations(observationContext *observation.Context) *operations {
	m := metrics.NewREDMetrics(
		observationContext.Registerer,
		"codeintel_symbols_transport_graphql",
		metrics.WithLabels("op"),
		metrics.WithCountHelp("Total number of method invocations."),
	)

	op := func(name string) *observation.Operation {
		return observationContext.Operation(observation.Op{
			Name:              fmt.Sprintf("codeintel.codenav.transport.graphql.%s", name),
			MetricLabelValues: []string{name},
			Metrics:           m,
		})
	}

	return &operations{
		hover:           op("Hover"),
		definitions:     op("Definitions"),
		references:      op("References"),
		implementations: op("Implementations"),
		diagnostics:     op("Diagnostics"),
		stencil:         op("Stencil"),
		ranges:          op("Ranges"),

		getSupportedByCtags:        op("GetSupportedByCtags"),
		getGitBlobLSIFDataResolver: op("GetGitBlobLSIFDataResolver"),
		getLanguagesRequestedBy:    op("GetLanguagesRequestedBy"),
		setRequestLanguageSupport:  op("SetRequestLanguageSupport"),
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
	fields := []log.Field{log.Duration("duration", duration)}
	if err != nil && *err != nil {
		fields = append(fields, log.Error(*err))
	}
	logger.Warn("Slow codeintel request", fields...)
}

func getObservationArgs(args shared.RequestArgs) observation.Args {
	return observation.Args{
		LogFields: []traceLog.Field{
			traceLog.Int("repositoryID", args.RepositoryID),
			traceLog.String("commit", args.Commit),
			traceLog.String("path", args.Path),
			traceLog.Int("line", args.Line),
			traceLog.Int("character", args.Character),
			traceLog.Int("limit", args.Limit),
		},
	}
}
