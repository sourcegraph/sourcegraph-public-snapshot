package lsifstore

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/sourcegraph/sourcegraph/internal/memo"
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

var m = memo.NewMemoizedConstructorWithArg(func(r prometheus.Registerer) (*metrics.REDMetrics, error) {
	return metrics.NewREDMetrics(
		r,
		"codeintel_codenav_lsifstore",
		metrics.WithLabels("op"),
		metrics.WithCountHelp("Total number of method invocations."),
	), nil
})

func newOperations(observationContext *observation.Context) *operations {
	metrics, _ := m.Init(observationContext.Registerer)

	op := func(name string) *observation.Operation {
		return observationContext.Operation(observation.Op{
			Name:              fmt.Sprintf("codeintel.codenav.lsifstore.%s", name),
			MetricLabelValues: []string{name},
			Metrics:           metrics,
		})
	}

	// suboperations do not have their own metrics but do have their
	// own opentracing spans. This allows us to more granularly track
	// the latency for parts of a request without noising up Prometheus.
	subOp := func(name string) *observation.Operation {
		return observationContext.Operation(observation.Op{
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
