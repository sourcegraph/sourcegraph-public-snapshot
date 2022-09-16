package dbstore

import (
	"fmt"

	"github.com/sourcegraph/sourcegraph/internal/metrics"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type operations struct {
	addUploadPart                               *observation.Operation
	calculateVisibleUploads                     *observation.Operation
	commitGraphMetadata                         *observation.Operation
	commitsVisibleToUpload                      *observation.Operation
	createConfigurationPolicy                   *observation.Operation
	definitionDumps                             *observation.Operation
	deleteConfigurationPolicyByID               *observation.Operation
	deleteIndexByID                             *observation.Operation
	deleteIndexesWithoutRepository              *observation.Operation
	deleteOverlappingDumps                      *observation.Operation
	deleteUploadByID                            *observation.Operation
	dequeue                                     *observation.Operation
	dequeueIndex                                *observation.Operation
	dirtyRepositories                           *observation.Operation
	findClosestDumps                            *observation.Operation
	findClosestDumpsFromGraphFragment           *observation.Operation
	getConfigurationPolicies                    *observation.Operation
	getConfigurationPolicyByID                  *observation.Operation
	getDumpsByIDs                               *observation.Operation
	getIndexByID                                *observation.Operation
	getIndexConfigurationByRepositoryID         *observation.Operation
	getIndexes                                  *observation.Operation
	getIndexesByIDs                             *observation.Operation
	getOldestCommitDate                         *observation.Operation
	getUploadByID                               *observation.Operation
	getUploads                                  *observation.Operation
	getUploadsByIDs                             *observation.Operation
	hardDeleteUploadByID                        *observation.Operation
	hasCommit                                   *observation.Operation
	hasRepository                               *observation.Operation
	indexQueueSize                              *observation.Operation
	insertCloneableDependencyRepo               *observation.Operation
	insertDependencyIndexingJob                 *observation.Operation
	insertDependencySyncingJob                  *observation.Operation
	insertIndex                                 *observation.Operation
	insertUpload                                *observation.Operation
	isQueued                                    *observation.Operation
	languagesRequestedBy                        *observation.Operation
	lastIndexScanForRepository                  *observation.Operation
	lastUploadRetentionScanForRepository        *observation.Operation
	markComplete                                *observation.Operation
	markErrored                                 *observation.Operation
	markFailed                                  *observation.Operation
	markIndexComplete                           *observation.Operation
	markIndexErrored                            *observation.Operation
	markQueued                                  *observation.Operation
	markRepositoryAsDirty                       *observation.Operation
	maxStaleAge                                 *observation.Operation
	queueSize                                   *observation.Operation
	recentIndexesSummary                        *observation.Operation
	recentUploadsSummary                        *observation.Operation
	referenceIDs                                *observation.Operation
	referencesForUpload                         *observation.Operation
	refreshCommitResolvability                  *observation.Operation
	repoIDsByGlobPatterns                       *observation.Operation
	repoName                                    *observation.Operation
	requestLanguageSupport                      *observation.Operation
	requeue                                     *observation.Operation
	requeueIndex                                *observation.Operation
	selectPoliciesForRepositoryMembershipUpdate *observation.Operation
	selectRepositoriesForRetentionScan          *observation.Operation
	selectRepositoriesForLockfileIndexScan      *observation.Operation
	updateCommitedAt                            *observation.Operation
	updateConfigurationPolicy                   *observation.Operation
	updateIndexConfigurationByRepositoryID      *observation.Operation
	updatePackageReferences                     *observation.Operation
	updatePackages                              *observation.Operation
	updateReferenceCounts                       *observation.Operation
	updateReposMatchingPatterns                 *observation.Operation
	updateUploadRetention                       *observation.Operation

	persistNearestUploads      *observation.Operation
	persistNearestUploadsLinks *observation.Operation
	persistUploadsVisibleAtTip *observation.Operation
	writeVisibleUploads        *observation.Operation
}

func newOperations(observationContext *observation.Context, metrics *metrics.REDMetrics) *operations {
	op := func(name string) *observation.Operation {
		return observationContext.Operation(observation.Op{
			Name:              fmt.Sprintf("codeintel.dbstore.%s", name),
			MetricLabelValues: []string{name},
			Metrics:           metrics,
			ErrorFilter: func(err error) observation.ErrorFilterBehaviour {
				return observation.EmitForDefault
			},
		})
	}

	// suboperations do not have their own metrics but do have their
	// own opentracing spans. This allows us to more granularly track
	// the latency for parts of a request without noising up Prometheus.
	subOp := func(name string) *observation.Operation {
		return observationContext.Operation(observation.Op{
			Name: fmt.Sprintf("codeintel.dbstore.%s", name),
		})
	}

	return &operations{
		addUploadPart:                        op("AddUploadPart"),
		calculateVisibleUploads:              op("CalculateVisibleUploads"),
		commitGraphMetadata:                  op("CommitGraphMetadata"),
		commitsVisibleToUpload:               op("CommitsVisibleToUpload"),
		createConfigurationPolicy:            op("CreateConfigurationPolicy"),
		definitionDumps:                      op("DefinitionDumps"),
		deleteConfigurationPolicyByID:        op("DeleteConfigurationPolicyByID"),
		deleteIndexByID:                      op("DeleteIndexByID"),
		deleteIndexesWithoutRepository:       op("DeleteIndexesWithoutRepository"),
		deleteOverlappingDumps:               op("DeleteOverlappingDumps"),
		deleteUploadByID:                     op("DeleteUploadByID"),
		dequeue:                              op("Dequeue"),
		dequeueIndex:                         op("DequeueIndex"),
		dirtyRepositories:                    op("DirtyRepositories"),
		findClosestDumps:                     op("FindClosestDumps"),
		findClosestDumpsFromGraphFragment:    op("FindClosestDumpsFromGraphFragment"),
		getConfigurationPolicies:             op("GetConfigurationPolicies"),
		getConfigurationPolicyByID:           op("GetConfigurationPolicyByID"),
		getDumpsByIDs:                        op("GetDumpsByIDs"),
		getIndexByID:                         op("GetIndexByID"),
		getIndexConfigurationByRepositoryID:  op("GetIndexConfigurationByRepositoryID"),
		getIndexes:                           op("GetIndexes"),
		getIndexesByIDs:                      op("GetIndexesByIDs"),
		getOldestCommitDate:                  op("GetOldestCommitDate"),
		getUploadByID:                        op("GetUploadByID"),
		getUploads:                           op("GetUploads"),
		getUploadsByIDs:                      op("GetUploadsByIDs"),
		hardDeleteUploadByID:                 op("HardDeleteUploadByID"),
		hasCommit:                            op("HasCommit"),
		hasRepository:                        op("HasRepository"),
		indexQueueSize:                       op("IndexQueueSize"),
		insertCloneableDependencyRepo:        op("InsertCloneableDependencyRepo"),
		insertDependencyIndexingJob:          op("InsertDependencyIndexingJob"),
		insertDependencySyncingJob:           op("InsertDependencySyncingJob"),
		insertIndex:                          op("InsertIndex"),
		insertUpload:                         op("InsertUpload"),
		isQueued:                             op("IsQueued"),
		languagesRequestedBy:                 op("LanguagesRequestedBy"),
		lastIndexScanForRepository:           op("LastIndexScanForRepository"),
		lastUploadRetentionScanForRepository: op("LastUploadRetentionScanForRepository"),
		markComplete:                         op("MarkComplete"),
		markErrored:                          op("MarkErrored"),
		markFailed:                           op("MarkFailed"),
		markIndexComplete:                    op("MarkIndexComplete"),
		markIndexErrored:                     op("MarkIndexErrored"),
		markQueued:                           op("MarkQueued"),
		markRepositoryAsDirty:                op("MarkRepositoryAsDirty"),
		maxStaleAge:                          op("MaxStaleAge"),
		queueSize:                            op("QueueSize"),
		recentIndexesSummary:                 op("RecentIndexesSummary"),
		recentUploadsSummary:                 op("RecentUploadsSummary"),
		referenceIDs:                         op("ReferenceIDs"),
		referencesForUpload:                  op("ReferencesForUpload"),
		refreshCommitResolvability:           op("RefreshCommitResolvability"),
		repoIDsByGlobPatterns:                op("repoIDsByGlobPatterns"),
		repoName:                             op("RepoName"),
		requestLanguageSupport:               op("RequestLanguageSupport"),
		requeue:                              op("Requeue"),
		requeueIndex:                         op("RequeueIndex"),

		selectPoliciesForRepositoryMembershipUpdate: op("selectPoliciesForRepositoryMembershipUpdate"),
		selectRepositoriesForRetentionScan:          op("SelectRepositoriesForRetentionScan"),
		selectRepositoriesForLockfileIndexScan:      op("SelectRepositoriesForLockfileIndexScan"),
		updateCommitedAt:                            op("UpdateCommitedAt"),
		updateConfigurationPolicy:                   op("UpdateConfigurationPolicy"),
		updateReferenceCounts:                       op("UpdateReferenceCounts"),

		updateIndexConfigurationByRepositoryID: op("UpdateIndexConfigurationByRepositoryID"),
		updatePackageReferences:                op("UpdatePackageReferences"),
		updatePackages:                         op("UpdatePackages"),
		updateReposMatchingPatterns:            op("UpdateReposMatchingPatterns"),
		updateUploadRetention:                  op("UpdateUploadRetention"),

		persistNearestUploads:      subOp("persistNearestUploads"),
		persistNearestUploadsLinks: subOp("persistNearestUploadsLinks"),
		persistUploadsVisibleAtTip: subOp("persistUploadsVisibleAtTip"),
		writeVisibleUploads:        subOp("writeVisibleUploads"),
	}
}
