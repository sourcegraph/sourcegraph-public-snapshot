package lsifstore

import (
	"fmt"

	"github.com/sourcegraph/sourcegraph/internal/metrics"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type operations struct {
	bulkMonikerResults *observation.Operation
	clear              *observation.Operation
	definitions        *observation.Operation
	diagnostics        *observation.Operation
	exists             *observation.Operation
	hover              *observation.Operation
	monikerResults     *observation.Operation
	monikersByPosition *observation.Operation
	packageInformation *observation.Operation
	ranges             *observation.Operation
	references         *observation.Operation
	writeDefinitions   *observation.Operation
	writeDocuments     *observation.Operation
	writeMeta          *observation.Operation
	writeReferences    *observation.Operation
	writeResultChunks  *observation.Operation

	locations           *observation.Operation
	locationsWithinFile *observation.Operation
}

func newOperations(observationContext *observation.Context) *operations {
	metrics := metrics.NewOperationMetrics(
		observationContext.Registerer,
		"codeintel_lsifstore",
		metrics.WithLabels("op"),
		metrics.WithCountHelp("Total number of method invocations."),
	)

	op := func(name string) *observation.Operation {
		return observationContext.Operation(observation.Op{
			Name:         fmt.Sprintf("codeintel.lsifstore.%s", name),
			MetricLabels: []string{name},
			Metrics:      metrics,
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
		bulkMonikerResults: op("BulkMonikerResults"),
		clear:              op("Clear"),
		definitions:        op("Definitions"),
		diagnostics:        op("Diagnostics"),
		exists:             op("Exists"),
		hover:              op("Hover"),
		monikerResults:     op("MonikerResults"),
		monikersByPosition: op("MonikersByPosition"),
		packageInformation: op("PackageInformation"),
		ranges:             op("Ranges"),
		references:         op("References"),
		writeDefinitions:   op("WriteDefinitions"),
		writeDocuments:     op("WriteDocuments"),
		writeMeta:          op("WriteMeta"),
		writeReferences:    op("WriteReferences"),
		writeResultChunks:  op("WriteResultChunks"),

		locations:           subOp("locations"),
		locationsWithinFile: subOp("locationsWithinFile"),
	}
}
