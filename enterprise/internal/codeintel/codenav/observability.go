package codenav

import (
	"context"
	"fmt"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/metrics"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type operations struct {
	getReferences          *observation.Operation
	getImplementations     *observation.Operation
	getDiagnostics         *observation.Operation
	getHover               *observation.Operation
	getDefinitions         *observation.Operation
	getRanges              *observation.Operation
	getStencil             *observation.Operation
	getDumpsByIDs          *observation.Operation
	getClosestDumpsForBlob *observation.Operation

	numUploadsRead         prometheus.Counter
	numBytesUploaded       prometheus.Counter
	numStaleRecordsDeleted prometheus.Counter
	numBytesDeleted        prometheus.Counter
}

func newOperations(observationContext *observation.Context) *operations {
	metrics := metrics.NewREDMetrics(
		observationContext.Registerer,
		"codeintel_codenav",
		metrics.WithLabels("op"),
		metrics.WithCountHelp("Total number of method invocations."),
	)

	op := func(name string) *observation.Operation {
		return observationContext.Operation(observation.Op{
			Name:              fmt.Sprintf("codeintel.codenav.%s", name),
			MetricLabelValues: []string{name},
			Metrics:           metrics,
		})
	}

	counter := func(name, help string) prometheus.Counter {
		counter := prometheus.NewCounter(prometheus.CounterOpts{
			Name: name,
			Help: help,
		})

		observationContext.Registerer.MustRegister(counter)
		return counter
	}

	numUploadsRead := counter(
		"src_codeintel_codenav_ranking_uploads_read_total",
		"The number of upload records read.",
	)
	numBytesUploaded := counter(
		"src_codeintel_codenav_ranking_bytes_uploaded_total",
		"The number of bytes uploaded to GCS.",
	)
	numStaleRecordsDeleted := counter(
		"src_codeintel_codenav_ranking_stale_uploads_removed_total",
		"The number of stale upload records removed from GCS.",
	)
	numBytesDeleted := counter(
		"src_codeintel_codenav_ranking_bytes_deleted_total",
		"The number of bytes deleted from GCS.",
	)

	return &operations{
		getReferences:          op("getReferences"),
		getImplementations:     op("getImplementations"),
		getDiagnostics:         op("getDiagnostics"),
		getHover:               op("getHover"),
		getDefinitions:         op("getDefinitions"),
		getRanges:              op("getRanges"),
		getStencil:             op("getStencil"),
		getDumpsByIDs:          op("GetDumpsByIDs"),
		getClosestDumpsForBlob: op("GetClosestDumpsForBlob"),

		numUploadsRead:         numUploadsRead,
		numBytesUploaded:       numBytesUploaded,
		numStaleRecordsDeleted: numStaleRecordsDeleted,
		numBytesDeleted:        numBytesDeleted,
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
