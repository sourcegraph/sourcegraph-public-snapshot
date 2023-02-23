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

	// Ranking
	exportRankingGraph *observation.Operation
	mapRankingGraph    *observation.Operation
	reduceRankingGraph *observation.Operation

	numUploadsRead                   prometheus.Counter
	numBytesUploaded                 prometheus.Counter
	numStaleRecordsDeleted           prometheus.Counter
	numDefinitionsInserted           prometheus.Counter
	numReferencesInserted            prometheus.Counter
	numStaleDefinitionRecordsDeleted prometheus.Counter
	numStaleReferenceRecordsDeleted  prometheus.Counter
	numMetadataRecordsDeleted        prometheus.Counter
	numInputRecordsDeleted           prometheus.Counter
	numBytesDeleted                  prometheus.Counter
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

	counter := func(name, help string) prometheus.Counter {
		metricsMu.Lock()
		defer metricsMu.Unlock()

		if c, ok := metricsMap[name]; ok {
			return c
		}

		counter := prometheus.NewCounter(prometheus.CounterOpts{
			Name: name,
			Help: help,
		})
		observationCtx.Registerer.MustRegister(counter)

		metricsMap[name] = counter

		return counter
	}

	numUploadsRead := counter(
		"src_codeintel_uploads_ranking_uploads_read_total",
		"The number of upload records read.",
	)
	numBytesUploaded := counter(
		"src_codeintel_uploads_ranking_bytes_uploaded_total",
		"The number of bytes uploaded to GCS.",
	)
	numStaleRecordsDeleted := counter(
		"src_codeintel_uploads_ranking_stale_uploads_removed_total",
		"The number of stale upload records removed from GCS.",
	)
	numDefinitionsInserted := counter(
		"src_codeintel_uploads_ranking_num_definitions_inserted_total",
		"The number of definition records inserted into Postgres.",
	)
	numReferencesInserted := counter(
		"src_codeintel_uploads_ranking_num_references_inserted_total",
		"The number of reference records inserted into Postgres.",
	)
	numStaleDefinitionRecordsDeleted := counter(
		"src_codeintel_uploads_num_stale_definition_records_deleted_total",
		"The number of stale definition records removed from Postgres.",
	)
	numStaleReferenceRecordsDeleted := counter(
		"src_codeintel_uploads_num_stale_reference_records_deleted_total",
		"The number of stale reference records removed from Postgres.",
	)
	numMetadataRecordsDeleted := counter(
		"src_codeintel_uploads_num_metadata_records_deleted_total",
		"The number of stale metadata records removed from Postgres.",
	)
	numInputRecordsDeleted := counter(
		"src_codeintel_uploads_num_input_records_deleted_total",
		"The number of stale input records removed from Postgres.",
	)
	numBytesDeleted := counter(
		"src_codeintel_uploads_ranking_bytes_deleted_total",
		"The number of bytes deleted from GCS.",
	)

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

		// Ranking
		exportRankingGraph: op("ExportRankingGraph"),
		mapRankingGraph:    op("MapRankingGraph"),
		reduceRankingGraph: op("ReduceRankingGraph"),

		numUploadsRead:                   numUploadsRead,
		numBytesUploaded:                 numBytesUploaded,
		numStaleRecordsDeleted:           numStaleRecordsDeleted,
		numDefinitionsInserted:           numDefinitionsInserted,
		numReferencesInserted:            numReferencesInserted,
		numStaleDefinitionRecordsDeleted: numStaleDefinitionRecordsDeleted,
		numStaleReferenceRecordsDeleted:  numStaleReferenceRecordsDeleted,
		numMetadataRecordsDeleted:        numMetadataRecordsDeleted,
		numInputRecordsDeleted:           numInputRecordsDeleted,
		numBytesDeleted:                  numBytesDeleted,
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
