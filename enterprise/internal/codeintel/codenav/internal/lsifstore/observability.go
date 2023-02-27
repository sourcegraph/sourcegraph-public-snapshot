package lsifstore

import (
	"fmt"

	"github.com/sourcegraph/sourcegraph/internal/metrics"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type operations struct {
	getReferences          *observation.Operation
	getImplementations     *observation.Operation
	getHover               *observation.Operation
	getDefinitions         *observation.Operation
	getDiagnostics         *observation.Operation
	getRanges              *observation.Operation
	getStencil             *observation.Operation
	getExists              *observation.Operation
	getMonikersByPosition  *observation.Operation
	getPackageInformation  *observation.Operation
	getBulkMonikerResults  *observation.Operation
	getLocationsWithinFile *observation.Operation

	locations *observation.Operation
}

var m = new(metrics.SingletonREDMetrics)

func newOperations(observationCtx *observation.Context) *operations {
	redMetrics := m.Get(func() *metrics.REDMetrics {
		return metrics.NewREDMetrics(
			observationCtx.Registerer,
			"codeintel_codenav_lsifstore",
			metrics.WithLabels("op"),
			metrics.WithCountHelp("Total number of method invocations."),
		)
	})

	op := func(name string) *observation.Operation {
		return observationCtx.Operation(observation.Op{
			Name:              fmt.Sprintf("codeintel.codenav.lsifstore.%s", name),
			MetricLabelValues: []string{name},
			Metrics:           redMetrics,
		})
	}

	// suboperations do not have their own metrics but do have their
	// own opentracing spans. This allows us to more granularly track
	// the latency for parts of a request without noising up Prometheus.
	subOp := func(name string) *observation.Operation {
		return observationCtx.Operation(observation.Op{
			Name: fmt.Sprintf("codeintel.lsifstore.%s", name),
		})
	}

	return &operations{
		getReferences:          op("GetReferences"),
		getImplementations:     op("GetImplementations"),
		getHover:               op("GetHover"),
		getDefinitions:         op("GetDefinitions"),
		getDiagnostics:         op("GetDiagnostics"),
		getRanges:              op("GetRanges"),
		getStencil:             op("GetStencil"),
		getExists:              op("GetExists"),
		getMonikersByPosition:  op("GetMonikersByPosition"),
		getPackageInformation:  op("GetPackageInformation"),
		getBulkMonikerResults:  op("GetBulkMonikerResults"),
		getLocationsWithinFile: op("GetLocationsWithinFile"),

		locations: subOp("locations"),
	}
}
