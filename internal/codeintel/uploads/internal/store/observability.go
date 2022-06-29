package store

import (
	"fmt"

	"github.com/sourcegraph/sourcegraph/internal/metrics"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type operations struct {
	list                           *observation.Operation
	staleSourcedCommits            *observation.Operation
	deleteSourcedCommits           *observation.Operation
	updateSourcedCommits           *observation.Operation
	getUploads                     *observation.Operation
	deleteUploadsStuckUploading    *observation.Operation
	deleteUploadsWithoutRepository *observation.Operation
	softDeleteExpiredUploads       *observation.Operation
	markRepositoryAsDirty          *observation.Operation
	dirtyRepositories              *observation.Operation
	updatePackages                 *observation.Operation
	updatePackageReferences        *observation.Operation
	updateUploadRetention          *observation.Operation
	updateUploadsReferenceCounts   *observation.Operation
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
		list:                           op("List"),
		staleSourcedCommits:            op("StaleSourcedCommits"),
		deleteSourcedCommits:           op("DeleteSourcedCommits"),
		updateSourcedCommits:           op("UpdateSourcedCommits"),
		getUploads:                     op("GetUploads"),
		deleteUploadsStuckUploading:    op("DeleteUploadsStuckUploading"),
		deleteUploadsWithoutRepository: op("DeleteUploadsWithoutRepository"),
		softDeleteExpiredUploads:       op("SoftDeleteExpiredUploads"),
		markRepositoryAsDirty:          op("MarkRepositoryAsDirty"),
		dirtyRepositories:              op("DirtyRepositories"),
		updatePackages:                 op("UpdatePackages"),
		updatePackageReferences:        op("UpdatePackageReferences"),
		updateUploadRetention:          op("UpdateUploadRetention"),
		updateUploadsReferenceCounts:   op("UpdateUploadsReferenceCounts"),
	}
}
