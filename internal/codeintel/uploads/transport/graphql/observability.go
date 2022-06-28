package graphql

import (
	"fmt"

	"github.com/sourcegraph/sourcegraph/internal/metrics"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type operations struct {
	commitGraph       *observation.Operation
	deleteLSIFUpload  *observation.Operation
	lsifUploadByID    *observation.Operation
	lsifUploads       *observation.Operation
	lsifUploadsByRepo *observation.Operation
}

func newOperations(observationContext *observation.Context) *operations {
	metrics := metrics.NewREDMetrics(
		observationContext.Registerer,
		"codeintel_uploads_transport_graphql",
		metrics.WithLabels("op"),
		metrics.WithCountHelp("Total number of method invocations."),
	)

	op := func(name string) *observation.Operation {
		return observationContext.Operation(observation.Op{
			Name:              fmt.Sprintf("codeintel.uploads.transport.graphql.%s", name),
			MetricLabelValues: []string{name},
			Metrics:           metrics,
		})
	}

	return &operations{
		commitGraph:       op("CommitGraph"),
		deleteLSIFUpload:  op("DeleteLSIFUpload"),
		lsifUploadByID:    op("LSIFUploadByID"),
		lsifUploads:       op("LSIFUploads"),
		lsifUploadsByRepo: op("LSIFUploadsByRepo"),
	}
}
