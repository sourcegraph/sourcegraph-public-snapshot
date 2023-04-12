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
	processStaleExportedUploads      *observation.Operation
	insertDefinitionsForRanking      *observation.Operation
	vacuumAbandonedDefinitions       *observation.Operation
	vacuumStaleDefinitions           *observation.Operation
	insertReferencesForRanking       *observation.Operation
	vacuumAbandonedReferences        *observation.Operation
	vacuumStaleReferences            *observation.Operation
	insertInitialPathRanks           *observation.Operation
	vacuumAbandonedInitialPathCounts *observation.Operation
	vacuumStaleInitialPaths          *observation.Operation
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
		processStaleExportedUploads:      op("ProcessStaleExportedUploads"),
		insertDefinitionsForRanking:      op("InsertDefinitionsForRanking"),
		vacuumAbandonedDefinitions:       op("VacuumAbandonedDefinitions"),
		vacuumStaleDefinitions:           op("VacuumStaleDefinitions"),
		insertReferencesForRanking:       op("InsertReferencesForRanking"),
		vacuumAbandonedReferences:        op("VacuumAbandonedReferences"),
		vacuumStaleReferences:            op("VacuumStaleReferences"),
		insertInitialPathRanks:           op("InsertInitialPathRanks"),
		vacuumAbandonedInitialPathCounts: op("VacuumAbandonedInitialPathCounts"),
		vacuumStaleInitialPaths:          op("VacuumStaleInitialPaths"),
		insertPathCountInputs:            op("InsertPathCountInputs"),
		insertInitialPathCounts:          op("InsertInitialPathCounts"),
		vacuumStaleGraphs:                op("VacuumStaleGraphs"),
		insertPathRanks:                  op("InsertPathRanks"),
		vacuumStaleRanks:                 op("VacuumStaleRanks"),
	}
}
