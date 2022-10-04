package graphql

import (
	"fmt"

	"github.com/sourcegraph/sourcegraph/internal/metrics"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type operations struct {
	// LSIF Uploads
	lsifUploadByID    *observation.Operation
	lsifUploadsByRepo *observation.Operation
	deleteLsifUpload  *observation.Operation
	deleteLsifUploads *observation.Operation

	// Commit Graph
	commitGraph *observation.Operation
}

func newOperations(observationContext *observation.Context) *operations {
	m := metrics.NewREDMetrics(
		observationContext.Registerer,
		"codeintel_uploads_transport_graphql",
		metrics.WithLabels("op"),
		metrics.WithCountHelp("Total number of method invocations."),
	)

	op := func(name string) *observation.Operation {
		return observationContext.Operation(observation.Op{
			Name:              fmt.Sprintf("codeintel.uploads.transport.graphql.%s", name),
			MetricLabelValues: []string{name},
			Metrics:           m,
		})
	}

	return &operations{
		// LSIF Uploads
		lsifUploadByID:    op("LSIFUploadByID"),
		lsifUploadsByRepo: op("LSIFUploadsByRepo"),
		deleteLsifUpload:  op("DeleteLSIFUpload"),
		deleteLsifUploads: op("DeleteLSIFUploads"),

		// Commit Graph
		commitGraph: op("CommitGraph"),
	}
}
