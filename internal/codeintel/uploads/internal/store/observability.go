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
	getCommitGraphMetadata    *observation.Operation
	hasCommit                 *observation.Operation

	// Repositories
	getRepositoriesForIndexScan     *observation.Operation
	getRepositoriesMaxStaleAge      *observation.Operation
	setRepositoryAsDirty            *observation.Operation
	setRepositoryAsDirtyWithTx      *observation.Operation
	getDirtyRepositories            *observation.Operation
	repoName                        *observation.Operation
	setRepositoriesForRetentionScan *observation.Operation
	hasRepository                   *observation.Operation

	// Uploads
	getUploads                        *observation.Operation
	updateUploadsVisibleToCommits     *observation.Operation
	writeVisibleUploads               *observation.Operation
	persistNearestUploads             *observation.Operation
	persistNearestUploadsLinks        *observation.Operation
	persistUploadsVisibleAtTip        *observation.Operation
	updateUploadRetention             *observation.Operation
	backfillReferenceCountBatch       *observation.Operation
	updateCommittedAt                 *observation.Operation
	sourcedCommitsWithoutCommittedAt  *observation.Operation
	updateUploadsReferenceCounts      *observation.Operation
	deleteUploadsWithoutRepository    *observation.Operation
	deleteUploadsStuckUploading       *observation.Operation
	softDeleteExpiredUploads          *observation.Operation
	hardDeleteUploadsByIDs            *observation.Operation
	getVisibleUploadsMatchingMonikers *observation.Operation

	// Dumps
	findClosestDumps                   *observation.Operation
	findClosestDumpsFromGraphFragment  *observation.Operation
	getDumpsWithDefinitionsForMonikers *observation.Operation
	getDumpsByIDs                      *observation.Operation

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
		getCommitGraphMetadata:    op("GetCommitGraphMetadata"),
		deleteSourcedCommits:      op("DeleteSourcedCommits"),
		updateSourcedCommits:      op("UpdateSourcedCommits"),
		hasCommit:                 op("HasCommit"),

		// Repositories
		getRepositoriesForIndexScan:     op("GetRepositoriesForIndexScan"),
		getRepositoriesMaxStaleAge:      op("GetRepositoriesMaxStaleAge"),
		getDirtyRepositories:            op("GetDirtyRepositories"),
		setRepositoryAsDirty:            op("SetRepositoryAsDirty"),
		setRepositoryAsDirtyWithTx:      op("SetRepositoryAsDirtyWithTx"),
		repoName:                        op("RepoName"),
		setRepositoriesForRetentionScan: op("SetRepositoriesForRetentionScan"),
		hasRepository:                   op("HasRepository"),

		// Uploads
		getUploads:                        op("GetUploads"),
		updateUploadsVisibleToCommits:     op("UpdateUploadsVisibleToCommits"),
		updateUploadRetention:             op("UpdateUploadRetention"),
		backfillReferenceCountBatch:       op("BackfillReferenceCountBatch"),
		updateCommittedAt:                 op("UpdateCommittedAt"),
		sourcedCommitsWithoutCommittedAt:  op("SourcedCommitsWithoutCommittedAt"),
		updateUploadsReferenceCounts:      op("UpdateUploadsReferenceCounts"),
		deleteUploadsStuckUploading:       op("DeleteUploadsStuckUploading"),
		deleteUploadsWithoutRepository:    op("DeleteUploadsWithoutRepository"),
		softDeleteExpiredUploads:          op("SoftDeleteExpiredUploads"),
		hardDeleteUploadsByIDs:            op("HardDeleteUploadsByIDs"),
		getVisibleUploadsMatchingMonikers: op("GetVisibleUploadsMatchingMonikers"),

		writeVisibleUploads:        op("writeVisibleUploads"),
		persistNearestUploads:      op("persistNearestUploads"),
		persistNearestUploadsLinks: op("persistNearestUploadsLinks"),
		persistUploadsVisibleAtTip: op("persistUploadsVisibleAtTip"),

		// Dumps
		findClosestDumps:                   op("FindClosestDumps"),
		findClosestDumpsFromGraphFragment:  op("FindClosestDumpsFromGraphFragment"),
		getDumpsWithDefinitionsForMonikers: op("GetUploadsWithDefinitionsForMonikers"),
		getDumpsByIDs:                      op("GetDumpsByIDs"),

		// Packages
		updatePackages: op("UpdatePackages"),

		// References
		updatePackageReferences: op("UpdatePackageReferences"),

		// Audit logs
		deleteOldAuditLogs: op("DeleteOldAuditLogs"),
	}
}
