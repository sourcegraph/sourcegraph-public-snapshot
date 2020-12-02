package dbstore

import (
	"fmt"

	"github.com/sourcegraph/sourcegraph/internal/metrics"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type operations struct {
	addUploadPart                         *observation.Operation
	calculateVisibleUploads               *observation.Operation
	deleteIndexByID                       *observation.Operation
	deleteIndexesWithoutRepository        *observation.Operation
	deleteOverlappingDumps                *observation.Operation
	deleteUploadByID                      *observation.Operation
	deleteUploadsStuckUploading           *observation.Operation
	deleteUploadsWithoutRepository        *observation.Operation
	dequeue                               *observation.Operation
	dequeueIndex                          *observation.Operation
	dirtyRepositories                     *observation.Operation
	findClosestDumps                      *observation.Operation
	findClosestDumpsFromGraphFragment     *observation.Operation
	getDumpByID                           *observation.Operation
	getIndexByID                          *observation.Operation
	getIndexConfigurationByRepositoryID   *observation.Operation
	getIndexes                            *observation.Operation
	getPackage                            *observation.Operation
	getRepositoriesWithIndexConfiguration *observation.Operation
	getUploadByID                         *observation.Operation
	getUploads                            *observation.Operation
	hardDeleteUploadByID                  *observation.Operation
	hasCommit                             *observation.Operation
	hasRepository                         *observation.Operation
	indexableRepositories                 *observation.Operation
	indexQueueSize                        *observation.Operation
	insertIndex                           *observation.Operation
	insertUpload                          *observation.Operation
	isQueued                              *observation.Operation
	lock                                  *observation.Operation
	markComplete                          *observation.Operation
	markErrored                           *observation.Operation
	markFailed                            *observation.Operation
	markIndexComplete                     *observation.Operation
	markIndexErrored                      *observation.Operation
	markQueued                            *observation.Operation
	markRepositoryAsDirty                 *observation.Operation
	packageReferencePager                 *observation.Operation
	queueSize                             *observation.Operation
	repoName                              *observation.Operation
	repoUsageStatistics                   *observation.Operation
	requeue                               *observation.Operation
	requeueIndex                          *observation.Operation
	resetIndexableRepositories            *observation.Operation
	sameRepoPager                         *observation.Operation
	softDeleteOldDumps                    *observation.Operation
	updateIndexableRepository             *observation.Operation
	updatePackageReferences               *observation.Operation
	updatePackages                        *observation.Operation
}

func makeOperations(observationContext *observation.Context) *operations {
	metrics := metrics.NewOperationMetrics(
		observationContext.Registerer,
		"codeintel_dbstore",
		metrics.WithLabels("op"),
		metrics.WithCountHelp("Total number of method invocations."),
	)

	op := func(name string) *observation.Operation {
		return observationContext.Operation(observation.Op{
			Name:         fmt.Sprintf("codeintel.dbstore.%s", name),
			MetricLabels: []string{name},
			Metrics:      metrics,
		})
	}

	return &operations{
		addUploadPart:                         op("AddUploadPart"),
		calculateVisibleUploads:               op("CalculateVisibleUploads"),
		deleteIndexByID:                       op("DeleteIndexByID"),
		deleteIndexesWithoutRepository:        op("DeleteIndexesWithoutRepository"),
		deleteOverlappingDumps:                op("DeleteOverlappingDumps"),
		deleteUploadByID:                      op("DeleteUploadByID"),
		deleteUploadsStuckUploading:           op("DeleteUploadsStuckUploading"),
		deleteUploadsWithoutRepository:        op("DeleteUploadsWithoutRepository"),
		dequeue:                               op("Dequeue"),
		dequeueIndex:                          op("DequeueIndex"),
		dirtyRepositories:                     op("DirtyRepositories"),
		findClosestDumps:                      op("FindClosestDumps"),
		findClosestDumpsFromGraphFragment:     op("FindClosestDumpsFromGraphFragment"),
		getDumpByID:                           op("GetDumpByID"),
		getIndexByID:                          op("GetIndexByID"),
		getIndexConfigurationByRepositoryID:   op("GetIndexConfigurationByRepositoryID"),
		getIndexes:                            op("GetIndexes"),
		getPackage:                            op("GetPackage"),
		getRepositoriesWithIndexConfiguration: op("GetRepositoriesWithIndexConfiguration"),
		getUploadByID:                         op("GetUploadByID"),
		getUploads:                            op("GetUploads"),
		hardDeleteUploadByID:                  op("HardDeleteUploadByID"),
		hasCommit:                             op("HasCommit"),
		hasRepository:                         op("HasRepository"),
		indexableRepositories:                 op("IndexableRepositories"),
		indexQueueSize:                        op("IndexQueueSize"),
		insertIndex:                           op("InsertIndex"),
		insertUpload:                          op("InsertUpload"),
		isQueued:                              op("IsQueued"),
		lock:                                  op("Lock"),
		markComplete:                          op("MarkComplete"),
		markErrored:                           op("MarkErrored"),
		markFailed:                            op("MarkFailed"),
		markIndexComplete:                     op("MarkIndexComplete"),
		markIndexErrored:                      op("MarkIndexErrored"),
		markQueued:                            op("MarkQueued"),
		markRepositoryAsDirty:                 op("MarkRepositoryAsDirty"),
		packageReferencePager:                 op("PackageReferencePager"),
		queueSize:                             op("QueueSize"),
		repoName:                              op("RepoName"),
		repoUsageStatistics:                   op("RepoUsageStatistics"),
		requeue:                               op("Requeue"),
		requeueIndex:                          op("RequeueIndex"),
		resetIndexableRepositories:            op("ResetIndexableRepositories"),
		sameRepoPager:                         op("SameRepoPager"),
		softDeleteOldDumps:                    op("SoftDeleteOldDumps"),
		updateIndexableRepository:             op("UpdateIndexableRepository"),
		updatePackageReferences:               op("UpdatePackageReferences"),
		updatePackages:                        op("UpdatePackages"),
	}
}
