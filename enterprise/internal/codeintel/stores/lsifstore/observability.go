package lsifstore

import (
	"fmt"

	"github.com/sourcegraph/sourcegraph/internal/metrics"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type operations struct {
	bulkMonikerResults         *observation.Operation
	clear                      *observation.Operation
	definitions                *observation.Operation
	diagnostics                *observation.Operation
	exists                     *observation.Operation
	hover                      *observation.Operation
	monikersByPosition         *observation.Operation
	monikers                   *observation.Operation
	packageInformation         *observation.Operation
	ranges                     *observation.Operation
	references                 *observation.Operation
	documentationPage          *observation.Operation
	documentationPathInfo      *observation.Operation
	packages                   *observation.Operation
	readSymbols                *observation.Operation // TODO(sqs): why readSymbols and symbols?
	symbols                    *observation.Operation
	symbol                     *observation.Operation
	writeDefinitions           *observation.Operation
	writeDocuments             *observation.Operation
	writeMeta                  *observation.Operation
	writeReferences            *observation.Operation
	writeResultChunks          *observation.Operation
	writeDocumentationPages    *observation.Operation
	writeDocumentationPathInfo *observation.Operation
	writeSymbols               *observation.Operation

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
		bulkMonikerResults:         op("BulkMonikerResults"),
		clear:                      op("Clear"),
		definitions:                op("Definitions"),
		diagnostics:                op("Diagnostics"),
		exists:                     op("Exists"),
		hover:                      op("Hover"),
		monikersByPosition:         op("MonikersByPosition"),
		monikers:                   op("Monikers"),
		packageInformation:         op("PackageInformation"),
		ranges:                     op("Ranges"),
		references:                 op("References"),
		documentationPage:          op("DocumentationPage"),
		documentationPathInfo:      op("DocumentationPathInfo"),
		packages:                   op("Packages"),
		readSymbols:                op("ReadSymbols"),
		symbols:                    op("Symbols"),
		symbol:                     op("Symbol"),
		writeDefinitions:           op("WriteDefinitions"),
		writeDocuments:             op("WriteDocuments"),
		writeMeta:                  op("WriteMeta"),
		writeReferences:            op("WriteReferences"),
		writeResultChunks:          op("WriteResultChunks"),
		writeDocumentationPages:    op("WriteDocumentationPages"),
		writeDocumentationPathInfo: op("WriteDocumentationPathInfo"),
		writeSymbols:               op("WriteSymbols"),

		locations:           subOp("locations"),
		locationsWithinFile: subOp("locationsWithinFile"),
	}
}
