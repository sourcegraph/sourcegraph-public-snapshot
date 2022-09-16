package codenav

import (
	"context"
	"fmt"
	"time"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/metrics"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type operations struct {
	getReferences                        *observation.Operation
	getImplementations                   *observation.Operation
	getDiagnostics                       *observation.Operation
	getHover                             *observation.Operation
	getDefinitions                       *observation.Operation
	getRanges                            *observation.Operation
	getStencil                           *observation.Operation
	getMonikersByPosition                *observation.Operation
	getBulkMonikerLocations              *observation.Operation
	getPackageInformation                *observation.Operation
	getUploadsWithDefinitionsForMonikers *observation.Operation
	getUploadIDsWithReferences           *observation.Operation
	getDumpsByIDs                        *observation.Operation
	getClosestDumpsForBlob               *observation.Operation
	getLanguagesRequestedBy              *observation.Operation
	setRequestLanguageSupport            *observation.Operation
}

func newOperations(observationContext *observation.Context) *operations {
	metrics := metrics.NewREDMetrics(
		observationContext.Registerer,
		"codeintel_symbols",
		metrics.WithLabels("op"),
		metrics.WithCountHelp("Total number of method invocations."),
	)

	op := func(name string) *observation.Operation {
		return observationContext.Operation(observation.Op{
			Name:              fmt.Sprintf("codeintel.symbols.%s", name),
			MetricLabelValues: []string{name},
			Metrics:           metrics,
		})
	}

	return &operations{
		getReferences:                        op("getReferences"),
		getImplementations:                   op("getImplementations"),
		getDiagnostics:                       op("getDiagnostics"),
		getHover:                             op("getHover"),
		getDefinitions:                       op("getDefinitions"),
		getRanges:                            op("getRanges"),
		getStencil:                           op("getStencil"),
		getMonikersByPosition:                op("GetMonikersByPosition"),
		getBulkMonikerLocations:              op("GetBulkMonikerLocations"),
		getPackageInformation:                op("GetPackageInformation"),
		getUploadsWithDefinitionsForMonikers: op("GetUploadsWithDefinitionsForMonikers"),
		getUploadIDsWithReferences:           op("GetUploadIDsWithReferences"),
		getDumpsByIDs:                        op("GetDumpsByIDs"),
		getClosestDumpsForBlob:               op("GetClosestDumpsForBlob"),
		getLanguagesRequestedBy:              op("GetLanguagesRequestedBy"),
		setRequestLanguageSupport:            op("SetRequestLanguageSupport"),
	}
}

var serviceObserverThreshold = time.Second

func observeResolver(ctx context.Context, err *error, operation *observation.Operation, threshold time.Duration, observationArgs observation.Args) (context.Context, observation.TraceLogger, func()) {
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
