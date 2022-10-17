package lsifstore

import (
	"fmt"

	"github.com/sourcegraph/sourcegraph/internal/metrics"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type operations struct {
	deleteLsifDataByUploadIds *observation.Operation
	writeMeta                 *observation.Operation
	writeDocuments            *observation.Operation
	writeResultChunks         *observation.Operation
	writeDefinitions          *observation.Operation
	writeReferences           *observation.Operation
	writeImplementations      *observation.Operation
}

func newOperations(observationContext *observation.Context) *operations {
	metrics := metrics.NewREDMetrics(
		observationContext.Registerer,
		"codeintel_uploads_lsifstore",
		metrics.WithLabels("op"),
		metrics.WithCountHelp("Total number of method invocations."),
	)

	op := func(name string) *observation.Operation {
		return observationContext.Operation(observation.Op{
			Name:              fmt.Sprintf("codeintel.uploads.lsifstore.%s", name),
			MetricLabelValues: []string{name},
			Metrics:           metrics,
		})
	}

	return &operations{
		deleteLsifDataByUploadIds: op("DeleteLsifDataByUploadIds"),
		writeMeta:                 op("WriteMeta"),
		writeDocuments:            op("WriteDocuments"),
		writeResultChunks:         op("WriteResultChunks"),
		writeDefinitions:          op("WriteDefinitions"),
		writeReferences:           op("WriteReferences"),
		writeImplementations:      op("WriteImplementations"),
	}
}
