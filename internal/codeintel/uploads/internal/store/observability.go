package store

import (
	"fmt"

	"github.com/sourcegraph/sourcegraph/internal/metrics"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type operations struct {
	// Not used yet.
	list *observation.Operation

	// Commits
	staleSourcedCommits  *observation.Operation
	deleteSourcedCommits *observation.Operation
	updateSourcedCommits *observation.Operation
	setRepositoryAsDirty *observation.Operation
	getDirtyRepositories *observation.Operation

	// Uploads
	getUploads                     *observation.Operation
	updateUploadRetention          *observation.Operation
	updateUploadsReferenceCounts   *observation.Operation
	deleteUploadsWithoutRepository *observation.Operation
	deleteUploadsStuckUploading    *observation.Operation
	softDeleteExpiredUploads       *observation.Operation
	hardDeleteUploadByID           *observation.Operation

	// Packages
	updatePackages *observation.Operation

	// References
	updatePackageReferences *observation.Operation

	// Audit logs
	deleteOldAuditLogs *observation.Operation
}

func newOperations(observationContext *observation.Context) *operations {
	metrics := metrics.NewREDMetrics(
		observationContext.Registerer,
		"codeintel_uploads_store",
		metrics.WithLabels("op"),
		metrics.WithCountHelp("Total number of method invocations."),
	)

	op := func(name string) *observation.Operation {
		return observationContext.Operation(observation.Op{
			Name:              fmt.Sprintf("codeintel.uploads.store.%s", name),
			MetricLabelValues: []string{name},
			Metrics:           metrics,
		})
	}

	return &operations{
		// Not used yet.
		list: op("List"),

		// Commits
		staleSourcedCommits:  op("StaleSourcedCommits"),
		deleteSourcedCommits: op("DeleteSourcedCommits"),
		updateSourcedCommits: op("UpdateSourcedCommits"),
		setRepositoryAsDirty: op("SetRepositoryAsDirty"),
		getDirtyRepositories: op("GetDirtyRepositories"),

		// Uploads
		getUploads:                     op("GetUploads"),
		updateUploadRetention:          op("UpdateUploadRetention"),
		updateUploadsReferenceCounts:   op("UpdateUploadsReferenceCounts"),
		deleteUploadsStuckUploading:    op("DeleteUploadsStuckUploading"),
		deleteUploadsWithoutRepository: op("DeleteUploadsWithoutRepository"),
		softDeleteExpiredUploads:       op("SoftDeleteExpiredUploads"),
		hardDeleteUploadByID:           op("HardDeleteUploadByID"),

		// Packages
		updatePackages: op("UpdatePackages"),

		// References
		updatePackageReferences: op("UpdatePackageReferences"),

		// Audit logs
		deleteOldAuditLogs: op("DeleteOldAuditLogs"),
	}
}
