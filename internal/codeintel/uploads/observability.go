package uploads

import (
	"fmt"

	"github.com/sourcegraph/sourcegraph/internal/metrics"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type operations struct {
	// Not used yet.
	list     *observation.Operation
	get      *observation.Operation
	getBatch *observation.Operation
	enqueue  *observation.Operation
	delete   *observation.Operation

	uploadsVisibleTo *observation.Operation

	// Commits
	getOldestCommitDate       *observation.Operation
	getCommitsVisibleToUpload *observation.Operation
	getStaleSourcedCommits    *observation.Operation
	updateSourcedCommits      *observation.Operation
	deleteSourcedCommits      *observation.Operation

	// Repositories
	getRepoName                     *observation.Operation
	getRepositoriesMaxStaleAge      *observation.Operation
	getDirtyRepositories            *observation.Operation
	setRepositoryAsDirty            *observation.Operation
	updateDirtyRepositories         *observation.Operation
	setRepositoriesForRetentionScan *observation.Operation

	// Uploads
	getUploads                        *observation.Operation
	getVisibleUploadsMatchingMonikers *observation.Operation
	updateUploadsVisibleToCommits     *observation.Operation
	updateUploadRetention             *observation.Operation
	backfillReferenceCountBatch       *observation.Operation
	updateUploadsReferenceCounts      *observation.Operation
	softDeleteExpiredUploads          *observation.Operation
	deleteUploadsWithoutRepository    *observation.Operation
	deleteUploadsStuckUploading       *observation.Operation
	hardDeleteUploads                 *observation.Operation
	inferClosestUploads               *observation.Operation
	backfillCommittedAtBatch          *observation.Operation

	// Dumps
	findClosestDumps                   *observation.Operation
	findClosestDumpsFromGraphFragment  *observation.Operation
	getDumpsWithDefinitionsForMonikers *observation.Operation
	getDumpsByIDs                      *observation.Operation

	// Packages
	updatePackages *observation.Operation

	// References
	updatePackageReferences *observation.Operation

	// Audit Logs
	deleteOldAuditLogs *observation.Operation
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

	return &operations{
		// Not used yet.
		list:             op("List"),
		get:              op("Get"),
		getBatch:         op("GetBatch"),
		enqueue:          op("Enqueue"),
		delete:           op("Delete"),
		uploadsVisibleTo: op("UploadsVisibleTo"),

		// Commits
		getOldestCommitDate:       op("GetOldestCommitDate"),
		getCommitsVisibleToUpload: op("GetCommitsVisibleToUpload"),
		getStaleSourcedCommits:    op("GetStaleSourcedCommits"),
		updateSourcedCommits:      op("UpdateSourcedCommits"),
		deleteSourcedCommits:      op("DeleteSourcedCommits"),

		// Repositories
		getRepoName:                     op("GetRepoName"),
		getRepositoriesMaxStaleAge:      op("GetRepositoriesMaxStaleAge"),
		getDirtyRepositories:            op("GetDirtyRepositories"),
		setRepositoryAsDirty:            op("SetRepositoryAsDirty"),
		updateDirtyRepositories:         op("UpdateDirtyRepositories"),
		setRepositoriesForRetentionScan: op("SetRepositoriesForRetentionScan"),

		// Uploads
		getUploads:                        op("GetUploads"),
		getVisibleUploadsMatchingMonikers: op("GetVisibleUploadsMatchingMonikers"),
		updateUploadsVisibleToCommits:     op("UpdateUploadsVisibleToCommits"),
		updateUploadRetention:             op("UpdateUploadRetention"),
		backfillReferenceCountBatch:       op("BackfillReferenceCountBatch"),
		updateUploadsReferenceCounts:      op("UpdateUploadsReferenceCounts"),
		deleteUploadsWithoutRepository:    op("DeleteUploadsWithoutRepository"),
		deleteUploadsStuckUploading:       op("DeleteUploadsStuckUploading"),
		softDeleteExpiredUploads:          op("SoftDeleteExpiredUploads"),
		hardDeleteUploads:                 op("HardDeleteUploads"),
		inferClosestUploads:               op("InferClosestUploads"),
		backfillCommittedAtBatch:          op("BackfillCommittedAtBatch"),

		// Dumps
		findClosestDumps:                   op("FindClosestDumps"),
		findClosestDumpsFromGraphFragment:  op("FindClosestDumpsFromGraphFragment"),
		getDumpsWithDefinitionsForMonikers: op("GetDumpsWithDefinitionsForMonikers"),
		getDumpsByIDs:                      op("GetDumpsByIDs"),

		// Packages
		updatePackages: op("UpdatePackages"),

		// References
		updatePackageReferences: op("UpdatePackageReferences"),

		// Audit Logs
		deleteOldAuditLogs: op("DeleteOldAuditLogs"),
	}
}
