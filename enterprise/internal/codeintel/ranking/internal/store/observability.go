package store

import (
	"fmt"

	"github.com/sourcegraph/sourcegraph/internal/metrics"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type operations struct {
	getStarRank                      *observation.Operation
	getDocumentRanks                 *observation.Operation
	getReferenceCountStatistics      *observation.Operation
	lastUpdatedAt                    *observation.Operation
	getUploadsForRanking             *observation.Operation
	vacuumAbandonedExportedUploads   *observation.Operation
	insertDefinitionsForRanking      *observation.Operation
	vacuumAbandonedDefinitions       *observation.Operation
	softDeleteStaleDefinitions       *observation.Operation
	vacuumDeletedDefinitions         *observation.Operation
	insertReferencesForRanking       *observation.Operation
	vacuumAbandonedReferences        *observation.Operation
	softDeleteStaleReferences        *observation.Operation
	vacuumDeletedReferences          *observation.Operation
	insertInitialPathRanks           *observation.Operation
	vacuumAbandonedInitialPathCounts *observation.Operation
	softDeleteStaleInitialPaths      *observation.Operation
	vacuumDeletedInitialPaths        *observation.Operation
	coordinate                       *observation.Operation
	insertPathCountInputs            *observation.Operation
	insertInitialPathCounts          *observation.Operation
	vacuumStaleGraphs                *observation.Operation
	insertPathRanks                  *observation.Operation
	vacuumStaleRanks                 *observation.Operation
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
		getStarRank:                      op("GetStarRank"),
		getDocumentRanks:                 op("GetDocumentRanks"),
		getReferenceCountStatistics:      op("GetReferenceCountStatistics"),
		lastUpdatedAt:                    op("LastUpdatedAt"),
		getUploadsForRanking:             op("GetUploadsForRanking"),
		vacuumAbandonedExportedUploads:   op("VacuumAbandonedExportedUploads"),
		insertDefinitionsForRanking:      op("InsertDefinitionsForRanking"),
		vacuumAbandonedDefinitions:       op("VacuumAbandonedDefinitions"),
		softDeleteStaleDefinitions:       op("SoftDeleteStaleDefinitions"),
		vacuumDeletedDefinitions:         op("VacuumDeletedDefinitions"),
		insertReferencesForRanking:       op("InsertReferencesForRanking"),
		vacuumAbandonedReferences:        op("VacuumAbandonedReferences"),
		softDeleteStaleReferences:        op("SoftDeleteStaleReferences"),
		vacuumDeletedReferences:          op("VacuumDeletedReferences"),
		insertInitialPathRanks:           op("InsertInitialPathRanks"),
		vacuumAbandonedInitialPathCounts: op("VacuumAbandonedInitialPathCounts"),
		softDeleteStaleInitialPaths:      op("SoftDeleteStaleInitialPaths"),
		vacuumDeletedInitialPaths:        op("VacuumDeletedInitialPaths"),
		coordinate:                       op("Coordinate"),
		insertPathCountInputs:            op("InsertPathCountInputs"),
		insertInitialPathCounts:          op("InsertInitialPathCounts"),
		vacuumStaleGraphs:                op("VacuumStaleGraphs"),
		insertPathRanks:                  op("InsertPathRanks"),
		vacuumStaleRanks:                 op("VacuumStaleRanks"),
	}
}
