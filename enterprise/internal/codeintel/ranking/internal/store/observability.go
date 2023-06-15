package store

import (
	"fmt"

	"github.com/sourcegraph/sourcegraph/internal/metrics"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type operations struct {
	summaries                      *observation.Operation
	getStarRank                    *observation.Operation
	getDocumentRanks               *observation.Operation
	getReferenceCountStatistics    *observation.Operation
	coverageCounts                 *observation.Operation
	lastUpdatedAt                  *observation.Operation
	getUploadsForRanking           *observation.Operation
	vacuumAbandonedExportedUploads *observation.Operation
	softDeleteStaleExportedUploads *observation.Operation
	vacuumDeletedExportedUploads   *observation.Operation
	insertDefinitionsForRanking    *observation.Operation
	insertReferencesForRanking     *observation.Operation
	insertInitialPathRanks         *observation.Operation
	derivativeGraphKey             *observation.Operation
	bumpDerivativeGraphKey         *observation.Operation
	deleteRankingProgress          *observation.Operation
	coordinate                     *observation.Operation
	insertPathCountInputs          *observation.Operation
	insertInitialPathCounts        *observation.Operation
	vacuumStaleProcessedReferences *observation.Operation
	vacuumStaleProcessedPaths      *observation.Operation
	vacuumStaleGraphs              *observation.Operation
	insertPathRanks                *observation.Operation
	vacuumStaleRanks               *observation.Operation
}

var m = new(metrics.SingletonREDMetrics)

func newOperations(observationCtx *observation.Context) *operations {
	m := m.Get(func() *metrics.REDMetrics {
		return metrics.NewREDMetrics(
			observationCtx.Registerer,
			"codeintel_ranking_store",
			metrics.WithLabels("op"),
			metrics.WithCountHelp("Total number of method invocations."),
		)
	})

	op := func(name string) *observation.Operation {
		return observationCtx.Operation(observation.Op{
			Name:              fmt.Sprintf("codeintel.ranking.store.%s", name),
			MetricLabelValues: []string{name},
			Metrics:           m,
		})
	}

	return &operations{
		summaries:                      op("Summaries"),
		getStarRank:                    op("GetStarRank"),
		getDocumentRanks:               op("GetDocumentRanks"),
		getReferenceCountStatistics:    op("GetReferenceCountStatistics"),
		coverageCounts:                 op("CoverageCounts"),
		lastUpdatedAt:                  op("LastUpdatedAt"),
		getUploadsForRanking:           op("GetUploadsForRanking"),
		vacuumAbandonedExportedUploads: op("VacuumAbandonedExportedUploads"),
		softDeleteStaleExportedUploads: op("SoftDeleteStaleExportedUploads"),
		vacuumDeletedExportedUploads:   op("VacuumDeletedExportedUploads"),
		insertDefinitionsForRanking:    op("InsertDefinitionsForRanking"),
		insertReferencesForRanking:     op("InsertReferencesForRanking"),
		insertInitialPathRanks:         op("InsertInitialPathRanks"),
		coordinate:                     op("Coordinate"),
		derivativeGraphKey:             op("DerivativeGraphKey"),
		bumpDerivativeGraphKey:         op("BumpDerivativeGraphKey"),
		deleteRankingProgress:          op("DeleteRankingProgress"),
		insertPathCountInputs:          op("InsertPathCountInputs"),
		insertInitialPathCounts:        op("InsertInitialPathCounts"),
		vacuumStaleProcessedReferences: op("VacuumStaleProcessedReferences"),
		vacuumStaleProcessedPaths:      op("VacuumStaleProcessedPaths"),
		vacuumStaleGraphs:              op("VacuumStaleGraphs"),
		insertPathRanks:                op("InsertPathRanks"),
		vacuumStaleRanks:               op("VacuumStaleRanks"),
	}
}
