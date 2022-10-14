package uploads

import (
	"fmt"

	"github.com/sourcegraph/sourcegraph/internal/metrics"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type operations struct {
	// Commits
	getCommitsVisibleToUpload *observation.Operation
	getCommitGraphMetadata    *observation.Operation
	getStaleSourcedCommits    *observation.Operation
	updateSourcedCommits      *observation.Operation
	deleteSourcedCommits      *observation.Operation

	// Repositories
	getRepoName                             *observation.Operation
	getRepositoriesForIndexScan             *observation.Operation
	getDirtyRepositories                    *observation.Operation
	getRecentUploadsSummary                 *observation.Operation
	getLastUploadRetentionScanForRepository *observation.Operation

	// Uploads
	getUploads                        *observation.Operation
	getUploadByID                     *observation.Operation
	getUploadsByIDs                   *observation.Operation
	getVisibleUploadsMatchingMonikers *observation.Operation
	getUploadDocumentsForPath         *observation.Operation
	updateUploadsVisibleToCommits     *observation.Operation
	deleteUploadByID                  *observation.Operation
	inferClosestUploads               *observation.Operation
	deleteUploadsWithoutRepository    *observation.Operation
	deleteUploadsStuckUploading       *observation.Operation
	softDeleteExpiredUploads          *observation.Operation
	hardDeleteUploadsByIDs            *observation.Operation
	deleteLsifDataByUploadIds         *observation.Operation

	// Dumps
	getDumpsWithDefinitionsForMonikers *observation.Operation
	getDumpsByIDs                      *observation.Operation

	// References
	referencesForUpload *observation.Operation

	// Audit Logs
	getAuditLogsForUpload *observation.Operation
	deleteOldAuditLogs    *observation.Operation

	// Tags
	getListTags *observation.Operation

	// // Worker metrics
	// uploadProcessor *observation.Operation
	// uploadSizeGuage prometheus.Gauge
}

func newOperations(observationContext *observation.Context) *operations {
	m := metrics.NewREDMetrics(
		observationContext.Registerer,
		"codeintel_uploads",
		metrics.WithLabels("op"),
		metrics.WithCountHelp("Total number of method invocations."),
	)

	op := func(name string) *observation.Operation {
		return observationContext.Operation(observation.Op{
			Name:              fmt.Sprintf("codeintel.uploads.%s", name),
			MetricLabelValues: []string{name},
			Metrics:           m,
		})
	}

	// honeyObservationContext := *observationContext
	// honeyObservationContext.HoneyDataset = &honey.Dataset{Name: "codeintel-worker"}
	// uploadProcessor := honeyObservationContext.Operation(observation.Op{
	// 	Name: "codeintel.uploadHandler",
	// 	ErrorFilter: func(err error) observation.ErrorFilterBehaviour {
	// 		return observation.EmitForTraces | observation.EmitForHoney
	// 	},
	// })

	// uploadSizeGuage := prometheus.NewGauge(prometheus.GaugeOpts{
	// 	Name: "src_codeintel_upload_processor_upload_size",
	// 	Help: "The combined size of uploads being processed at this instant by this worker.",
	// })
	// observationContext.Registerer.MustRegister(uploadSizeGuage)

	return &operations{
		// Commits
		getCommitsVisibleToUpload: op("GetCommitsVisibleToUpload"),
		getCommitGraphMetadata:    op("GetCommitGraphMetadata"),
		getStaleSourcedCommits:    op("GetStaleSourcedCommits"),
		updateSourcedCommits:      op("UpdateSourcedCommits"),
		deleteSourcedCommits:      op("DeleteSourcedCommits"),

		// Repositories
		getRepoName:                             op("GetRepoName"),
		getRepositoriesForIndexScan:             op("GetRepositoriesForIndexScan"),
		getDirtyRepositories:                    op("GetDirtyRepositories"),
		getRecentUploadsSummary:                 op("GetRecentUploadsSummary"),
		getLastUploadRetentionScanForRepository: op("GetLastUploadRetentionScanForRepository"),

		// Uploads
		getUploads:                        op("GetUploads"),
		getUploadByID:                     op("GetUploadByID"),
		getUploadsByIDs:                   op("GetUploadsByIDs"),
		getVisibleUploadsMatchingMonikers: op("GetVisibleUploadsMatchingMonikers"),
		getUploadDocumentsForPath:         op("GetUploadDocumentsForPath"),
		updateUploadsVisibleToCommits:     op("UpdateUploadsVisibleToCommits"),
		deleteUploadByID:                  op("DeleteUploadByID"),
		inferClosestUploads:               op("InferClosestUploads"),
		deleteUploadsWithoutRepository:    op("DeleteUploadsWithoutRepository"),
		deleteUploadsStuckUploading:       op("DeleteUploadsStuckUploading"),
		softDeleteExpiredUploads:          op("SoftDeleteExpiredUploads"),
		hardDeleteUploadsByIDs:            op("HardDeleteUploadsByIDs"),
		deleteLsifDataByUploadIds:         op("DeleteLsifDataByUploadIds"),

		// Dumps
		getDumpsWithDefinitionsForMonikers: op("GetDumpsWithDefinitionsForMonikers"),
		getDumpsByIDs:                      op("GetDumpsByIDs"),

		// References
		referencesForUpload: op("ReferencesForUpload"),

		// Audit Logs
		getAuditLogsForUpload: op("GetAuditLogsForUpload"),
		deleteOldAuditLogs:    op("DeleteOldAuditLogs"),

		// Tags
		getListTags: op("GetListTags"),

		// // Worker metrics
		// uploadProcessor: uploadProcessor,
		// uploadSizeGuage: uploadSizeGuage,
	}
}
