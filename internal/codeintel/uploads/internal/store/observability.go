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
	getStaleSourcedCommits    *observation.Operation
	deleteSourcedCommits      *observation.Operation
	updateSourcedCommits      *observation.Operation
	getCommitsVisibleToUpload *observation.Operation
	getOldestCommitDate       *observation.Operation

	// Repositories
	getRepositoriesMaxStaleAge      *observation.Operation
	setRepositoryAsDirty            *observation.Operation
	getDirtyRepositories            *observation.Operation
	repoName                        *observation.Operation
	setRepositoriesForRetentionScan *observation.Operation

	// Uploads
	getUploads                     *observation.Operation
	updateUploadsVisibleToCommits  *observation.Operation
	writeVisibleUploads            *observation.Operation
	persistNearestUploads          *observation.Operation
	persistNearestUploadsLinks     *observation.Operation
	persistUploadsVisibleAtTip     *observation.Operation
	updateUploadRetention          *observation.Operation
	updateUploadsReferenceCounts   *observation.Operation
	deleteUploadsWithoutRepository *observation.Operation
	deleteUploadsStuckUploading    *observation.Operation
	softDeleteExpiredUploads       *observation.Operation
	hardDeleteUploadsByIDs         *observation.Operation

	// Dumps
	findClosestDumps                  *observation.Operation
	findClosestDumpsFromGraphFragment *observation.Operation

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
		getCommitsVisibleToUpload: op("CommitsVisibleToUploads"),
		getOldestCommitDate:       op("GetOldestCommitDate"),
		getStaleSourcedCommits:    op("GetStaleSourcedCommits"),
		deleteSourcedCommits:      op("DeleteSourcedCommits"),
		updateSourcedCommits:      op("UpdateSourcedCommits"),

		// Repositories
		getRepositoriesMaxStaleAge:      op("GetRepositoriesMaxStaleAge"),
		getDirtyRepositories:            op("GetDirtyRepositories"),
		setRepositoryAsDirty:            op("SetRepositoryAsDirty"),
		repoName:                        op("RepoName"),
		setRepositoriesForRetentionScan: op("SetRepositoriesForRetentionScan"),

		// Uploads
		getUploads:                     op("GetUploads"),
		updateUploadsVisibleToCommits:  op("UpdateUploadsVisibleToCommits"),
		updateUploadRetention:          op("UpdateUploadRetention"),
		updateUploadsReferenceCounts:   op("UpdateUploadsReferenceCounts"),
		deleteUploadsStuckUploading:    op("DeleteUploadsStuckUploading"),
		deleteUploadsWithoutRepository: op("DeleteUploadsWithoutRepository"),
		softDeleteExpiredUploads:       op("SoftDeleteExpiredUploads"),
		hardDeleteUploadsByIDs:         op("HardDeleteUploadsByIDs"),

		writeVisibleUploads:        op("writeVisibleUploads"),
		persistNearestUploads:      op("persistNearestUploads"),
		persistNearestUploadsLinks: op("persistNearestUploadsLinks"),
		persistUploadsVisibleAtTip: op("persistUploadsVisibleAtTip"),

		// Dumps
		findClosestDumps:                  op("FindClosestDumps"),
		findClosestDumpsFromGraphFragment: op("FindClosestDumpsFromGraphFragment"),

		// Packages
		updatePackages: op("UpdatePackages"),

		// References
		updatePackageReferences: op("UpdatePackageReferences"),

		// Audit logs
		deleteOldAuditLogs: op("DeleteOldAuditLogs"),
	}
}
