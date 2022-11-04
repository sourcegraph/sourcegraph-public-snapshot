package graphql

import (
	"fmt"

	"github.com/sourcegraph/sourcegraph/internal/metrics"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type operations struct {
	getIndexByID                            *observation.Operation
	getUploadDocumentsForPath               *observation.Operation
	getCommitsVisibleToUpload               *observation.Operation
	getCommitGraphMetadata                  *observation.Operation
	getAuditLogsForUpload                   *observation.Operation
	getRecentUploadsSummary                 *observation.Operation
	getLastUploadRetentionScanForRepository *observation.Operation
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
		getIndexByID:                            op("GetIndexByID"),
		getUploadDocumentsForPath:               op("GetUploadDocumentsForPath"),
		getCommitsVisibleToUpload:               op("GetCommitsVisibleToUpload"),
		getCommitGraphMetadata:                  op("GetCommitGraphMetadata"),
		getAuditLogsForUpload:                   op("GetAuditLogsForUpload"),
		getRecentUploadsSummary:                 op("GetRecentUploadsSummary"),
		getLastUploadRetentionScanForRepository: op("GetLastUploadRetentionScanForRepository"),
	}
}
