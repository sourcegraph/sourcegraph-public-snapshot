package uploads

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/sourcegraph/log"

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
	setRepositoriesForRetentionScan         *observation.Operation
	getRepositoriesMaxStaleAge              *observation.Operation

	// Uploads
	getIndexers                          *observation.Operation
	getUploads                           *observation.Operation
	getUploadByID                        *observation.Operation
	getUploadsByIDs                      *observation.Operation
	getVisibleUploadsMatchingMonikers    *observation.Operation
	getUploadDocumentsForPath            *observation.Operation
	updateUploadsVisibleToCommits        *observation.Operation
	deleteUploadByID                     *observation.Operation
	inferClosestUploads                  *observation.Operation
	deleteUploadsWithoutRepository       *observation.Operation
	deleteUploadsStuckUploading          *observation.Operation
	softDeleteExpiredUploads             *observation.Operation
	softDeleteExpiredUploadsViaTraversal *observation.Operation
	hardDeleteUploadsByIDs               *observation.Operation
	deleteLsifDataByUploadIds            *observation.Operation

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
}

var (
	metricsMap = make(map[string]prometheus.Counter)
	m          = new(metrics.SingletonREDMetrics)
	metricsMu  sync.Mutex
)

func newOperations(observationCtx *observation.Context) *operations {
	m := m.Get(func() *metrics.REDMetrics {
		return metrics.NewREDMetrics(
			observationCtx.Registerer,
			"codeintel_uploads",
			metrics.WithLabels("op"),
			metrics.WithCountHelp("Total number of method invocations."),
		)
	})

	op := func(name string) *observation.Operation {
		return observationCtx.Operation(observation.Op{
			Name:              fmt.Sprintf("codeintel.uploads.%s", name),
			MetricLabelValues: []string{name},
			Metrics:           m,
		})
	}

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
		setRepositoriesForRetentionScan:         op("SetRepositoriesForRetentionScan"),
		getRepositoriesMaxStaleAge:              op("GetRepositoriesMaxStaleAge"),

		// Uploads
		getIndexers:                          op("GetIndexers"),
		getUploads:                           op("GetUploads"),
		getUploadByID:                        op("GetUploadByID"),
		getUploadsByIDs:                      op("GetUploadsByIDs"),
		getVisibleUploadsMatchingMonikers:    op("GetVisibleUploadsMatchingMonikers"),
		getUploadDocumentsForPath:            op("GetUploadDocumentsForPath"),
		updateUploadsVisibleToCommits:        op("UpdateUploadsVisibleToCommits"),
		deleteUploadByID:                     op("DeleteUploadByID"),
		inferClosestUploads:                  op("InferClosestUploads"),
		deleteUploadsWithoutRepository:       op("DeleteUploadsWithoutRepository"),
		deleteUploadsStuckUploading:          op("DeleteUploadsStuckUploading"),
		softDeleteExpiredUploads:             op("SoftDeleteExpiredUploads"),
		softDeleteExpiredUploadsViaTraversal: op("SoftDeleteExpiredUploadsViaTraversal"),
		hardDeleteUploadsByIDs:               op("HardDeleteUploadsByIDs"),
		deleteLsifDataByUploadIds:            op("DeleteLsifDataByUploadIds"),

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
	}
}

func MetricReporters(observationCtx *observation.Context, uploadSvc UploadService) {
	observationCtx.Registerer.MustRegister(prometheus.NewGaugeFunc(prometheus.GaugeOpts{
		Name: "src_codeintel_commit_graph_total",
		Help: "Total number of repositories with stale commit graphs.",
	}, func() float64 {
		dirtyRepositories, err := uploadSvc.GetDirtyRepositories(context.Background())
		if err != nil {
			observationCtx.Logger.Error("Failed to determine number of dirty repositories", log.Error(err))
		}

		return float64(len(dirtyRepositories))
	}))

	observationCtx.Registerer.MustRegister(prometheus.NewGaugeFunc(prometheus.GaugeOpts{
		Name: "src_codeintel_commit_graph_queued_duration_seconds_total",
		Help: "The maximum amount of time a repository has had a stale commit graph.",
	}, func() float64 {
		age, err := uploadSvc.GetRepositoriesMaxStaleAge(context.Background())
		if err != nil {
			observationCtx.Logger.Error("Failed to determine stale commit graph age", log.Error(err))
			return 0
		}

		return float64(age) / float64(time.Second)
	}))
}
