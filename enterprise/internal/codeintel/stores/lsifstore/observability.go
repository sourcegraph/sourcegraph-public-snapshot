package lsifstore

import (
	"fmt"

	"github.com/sourcegraph/sourcegraph/internal/metrics"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type operations struct {
	clear              *observation.Operation
	definitions        *observation.Operation
	diagnostics        *observation.Operation
	exists             *observation.Operation
	hover              *observation.Operation
	monikerResults     *observation.Operation
	monikersByPosition *observation.Operation
	packageInformation *observation.Operation
	pathsWithPrefix    *observation.Operation
	ranges             *observation.Operation
	readDefinitions    *observation.Operation
	readDocument       *observation.Operation
	readMeta           *observation.Operation
	readReferences     *observation.Operation
	readResultChunk    *observation.Operation
	reaResultChunk     *observation.Operation
	references         *observation.Operation
	writeDefinitions   *observation.Operation
	writeDocuments     *observation.Operation
	writeMeta          *observation.Operation
	writeReferences    *observation.Operation
	writeResultChunks  *observation.Operation
}

func makeOperations(observationContext *observation.Context) *operations {
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

	return &operations{
		clear:              op("Clear"),
		definitions:        op("Definitions"),
		diagnostics:        op("Diagnostics"),
		exists:             op("Exists"),
		hover:              op("Hover"),
		monikerResults:     op("MonikerResults"),
		monikersByPosition: op("MonikersByPosition"),
		packageInformation: op("PackageInformation"),
		pathsWithPrefix:    op("PathsWithPrefix"),
		ranges:             op("Ranges"),
		readDefinitions:    op("ReadDefinitions"),
		readDocument:       op("ReadDocument"),
		readMeta:           op("ReadMeta"),
		readReferences:     op("ReadReferences"),
		readResultChunk:    op("ReadResultChunk"),
		reaResultChunk:     op("ReaResultChunk"),
		references:         op("References"),
		writeDefinitions:   op("WriteDefinitions"),
		writeDocuments:     op("WriteDocuments"),
		writeMeta:          op("WriteMeta"),
		writeReferences:    op("WriteReferences"),
		writeResultChunks:  op("WriteResultChunks"),
	}
}
