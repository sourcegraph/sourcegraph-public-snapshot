package janitor

import (
	"context"
	"time"

	rankingshared "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/ranking/internal/shared"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/ranking/internal/store"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/shared/background"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

const recordTypeName = "path count inputs"

func NewSymbolDefinitionsJanitor(
	observationCtx *observation.Context,
	store store.Store,
	config *Config,
) goroutine.BackgroundRoutine {
	name := "codeintel.ranking.symbol-definitions-janitor"

	return background.NewJanitorJob(context.Background(), background.JanitorOptions{
		Name:        name,
		Description: "Soft-deletes stale data from the ranking definitions table.",
		Interval:    config.Interval,
		Metrics:     background.NewJanitorMetrics(observationCtx, name, recordTypeName),
		CleanupFunc: func(ctx context.Context) (numRecordsScanned int, numRecordsAltered int, err error) {
			return softDeleteStaleDefinitions(ctx, store)
		},
	})
}

func NewSymbolReferencesJanitor(
	observationCtx *observation.Context,
	store store.Store,
	config *Config,
) goroutine.BackgroundRoutine {
	name := "codeintel.ranking.symbol-references-janitor"

	return background.NewJanitorJob(context.Background(), background.JanitorOptions{
		Name:        name,
		Description: "Soft-deletes stale data from the ranking references table.",
		Interval:    config.Interval,
		Metrics:     background.NewJanitorMetrics(observationCtx, name, recordTypeName),
		CleanupFunc: func(ctx context.Context) (numRecordsScanned int, numRecordsAltered int, err error) {
			return softDeleteStaleReferences(ctx, store)
		},
	})
}

func NewSymbolInitialPathsJanitor(
	observationCtx *observation.Context,
	store store.Store,
	config *Config,
) goroutine.BackgroundRoutine {
	name := "codeintel.ranking.symbol-initial-paths-janitor"

	return background.NewJanitorJob(context.Background(), background.JanitorOptions{
		Name:        name,
		Description: "Soft-deletes stale data from the ranking initial paths table.",
		Interval:    config.Interval,
		Metrics:     background.NewJanitorMetrics(observationCtx, name, recordTypeName),
		CleanupFunc: func(ctx context.Context) (numRecordsScanned int, numRecordsAltered int, err error) {
			return softDeleteStaleInitialPaths(ctx, store)
		},
	})
}

func NewDeletedSymbolDefinitionsJanitor(
	observationCtx *observation.Context,
	store store.Store,
	config *Config,
) goroutine.BackgroundRoutine {
	name := "codeintel.ranking.deleted-symbol-definitions-janitor"

	return background.NewJanitorJob(context.Background(), background.JanitorOptions{
		Name:        name,
		Description: "Removes soft-deleted data from the ranking definitions table no longer being read by a mapper process.",
		Interval:    config.Interval,
		Metrics:     background.NewJanitorMetrics(observationCtx, name, recordTypeName),
		CleanupFunc: func(ctx context.Context) (numRecordsScanned int, numRecordsAltered int, err error) {
			numDeleted, err := vacuumDeletedDefinitions(ctx, store)
			return numDeleted, numDeleted, err
		},
	})
}

func NewDeletedSymbolReferencesJanitor(
	observationCtx *observation.Context,
	store store.Store,
	config *Config,
) goroutine.BackgroundRoutine {
	name := "codeintel.ranking.deleted-symbol-references-janitor"

	return background.NewJanitorJob(context.Background(), background.JanitorOptions{
		Name:        name,
		Description: "Removes soft-deleted data from the ranking references table no longer being read by a mapper process.",
		Interval:    config.Interval,
		Metrics:     background.NewJanitorMetrics(observationCtx, name, recordTypeName),
		CleanupFunc: func(ctx context.Context) (numRecordsScanned int, numRecordsAltered int, err error) {
			numDeleted, err := vacuumDeletedReferences(ctx, store)
			return numDeleted, numDeleted, err
		},
	})
}

func NewDeletedSymbolInitialPathsJanitor(
	observationCtx *observation.Context,
	store store.Store,
	config *Config,
) goroutine.BackgroundRoutine {
	name := "codeintel.ranking.deleted-symbol-initial-paths-janitor"

	return background.NewJanitorJob(context.Background(), background.JanitorOptions{
		Name:        name,
		Description: "Removes soft-deleted data from the ranking initial paths table no longer being read by a seed mapper process.",
		Interval:    config.Interval,
		Metrics:     background.NewJanitorMetrics(observationCtx, name, recordTypeName),
		CleanupFunc: func(ctx context.Context) (numRecordsScanned int, numRecordsAltered int, err error) {
			numDeleted, err := vacuumDeletedInitialPaths(ctx, store)
			return numDeleted, numDeleted, err
		},
	})
}

func NewAbandonedDefinitionsJanitor(
	observationCtx *observation.Context,
	store store.Store,
	config *Config,
) goroutine.BackgroundRoutine {
	name := "codeintel.ranking.abandoned-definitions-janitor"

	return background.NewJanitorJob(context.Background(), background.JanitorOptions{
		Name:        name,
		Description: "Removes definitions records for old graph keys.",
		Interval:    config.Interval,
		Metrics:     background.NewJanitorMetrics(observationCtx, name, recordTypeName),
		CleanupFunc: func(ctx context.Context) (numRecordsScanned int, numRecordsAltered int, err error) {
			numDeleted, err := vacuumAbandonedDefinitions(ctx, store)
			return numDeleted, numDeleted, err
		},
	})
}

func NewAbandonedReferencesJanitor(
	observationCtx *observation.Context,
	store store.Store,
	config *Config,
) goroutine.BackgroundRoutine {
	name := "codeintel.ranking.abandoned-references-janitor"

	return background.NewJanitorJob(context.Background(), background.JanitorOptions{
		Name:        name,
		Description: "Removes references records for old graph keys.",
		Interval:    config.Interval,
		Metrics:     background.NewJanitorMetrics(observationCtx, name, recordTypeName),
		CleanupFunc: func(ctx context.Context) (numRecordsScanned int, numRecordsAltered int, err error) {
			numDeleted, err := vacuumAbandonedReferences(ctx, store)
			return numDeleted, numDeleted, err
		},
	})
}

func NewAbandonedInitialCountsJanitor(
	observationCtx *observation.Context,
	store store.Store,
	config *Config,
) goroutine.BackgroundRoutine {
	name := "codeintel.ranking.abandoned-initial-counts-janitor"

	return background.NewJanitorJob(context.Background(), background.JanitorOptions{
		Name:        name,
		Description: "Removes initial count records for old graph keys.",
		Interval:    config.Interval,
		Metrics:     background.NewJanitorMetrics(observationCtx, name, recordTypeName),
		CleanupFunc: func(ctx context.Context) (numRecordsScanned int, numRecordsAltered int, err error) {
			numDeleted, err := vacuumAbandonedInitialPathCounts(ctx, store)
			return numDeleted, numDeleted, err
		},
	})
}

func NewRankCountsJanitor(
	observationCtx *observation.Context,
	store store.Store,
	config *Config,
) goroutine.BackgroundRoutine {
	name := "codeintel.ranking.rank-counts-janitor"

	return background.NewJanitorJob(context.Background(), background.JanitorOptions{
		Name:        name,
		Description: "Removes old path count input records.",
		Interval:    config.Interval,
		Metrics:     background.NewJanitorMetrics(observationCtx, name, recordTypeName),
		CleanupFunc: func(ctx context.Context) (numRecordsScanned int, numRecordsAltered int, err error) {
			numDeleted, err := vacuumStaleGraphs(ctx, store)
			return numDeleted, numDeleted, err
		},
	})
}

func NewRankJanitor(
	observationCtx *observation.Context,
	store store.Store,
	config *Config,
) goroutine.BackgroundRoutine {
	name := "codeintel.ranking.rank-janitor"

	return background.NewJanitorJob(context.Background(), background.JanitorOptions{
		Name:        name,
		Description: "Removes stale ranking data.",
		Interval:    config.Interval,
		Metrics:     background.NewJanitorMetrics(observationCtx, name, recordTypeName),
		CleanupFunc: func(ctx context.Context) (numRecordsScanned int, numRecordsAltered int, err error) {
			return vacuumStaleRanks(ctx, store)
		},
	})
}

func softDeleteStaleDefinitions(ctx context.Context, store store.Store) (int, int, error) {
	if enabled := conf.CodeIntelRankingDocumentReferenceCountsEnabled(); !enabled {
		return 0, 0, nil
	}

	numDefinitionRecordsScanned, numDefinitionRecordsRemoved, err := store.SoftDeleteStaleDefinitions(ctx, rankingshared.GraphKey())
	return numDefinitionRecordsScanned, numDefinitionRecordsRemoved, err
}

func softDeleteStaleReferences(ctx context.Context, store store.Store) (int, int, error) {
	if enabled := conf.CodeIntelRankingDocumentReferenceCountsEnabled(); !enabled {
		return 0, 0, nil
	}

	numReferenceRecordsScanned, numReferenceRecordsRemoved, err := store.SoftDeleteStaleReferences(ctx, rankingshared.GraphKey())
	return numReferenceRecordsScanned, numReferenceRecordsRemoved, err
}

func softDeleteStaleInitialPaths(ctx context.Context, store store.Store) (int, int, error) {
	if enabled := conf.CodeIntelRankingDocumentReferenceCountsEnabled(); !enabled {
		return 0, 0, nil
	}

	numPathRecordsScanned, numStalePathRecordsDeleted, err := store.SoftDeleteStaleInitialPaths(ctx, rankingshared.GraphKey())
	return numPathRecordsScanned, numStalePathRecordsDeleted, err
}

func vacuumDeletedDefinitions(ctx context.Context, store store.Store) (int, error) {
	if enabled := conf.CodeIntelRankingDocumentReferenceCountsEnabled(); !enabled {
		return 0, nil
	}

	return store.VacuumDeletedDefinitions(ctx, rankingshared.DerivativeGraphKeyFromTime(time.Now()))
}

func vacuumDeletedReferences(ctx context.Context, store store.Store) (int, error) {
	if enabled := conf.CodeIntelRankingDocumentReferenceCountsEnabled(); !enabled {
		return 0, nil
	}

	return store.VacuumDeletedReferences(ctx, rankingshared.DerivativeGraphKeyFromTime(time.Now()))
}

func vacuumDeletedInitialPaths(ctx context.Context, store store.Store) (int, error) {
	if enabled := conf.CodeIntelRankingDocumentReferenceCountsEnabled(); !enabled {
		return 0, nil
	}

	return store.VacuumDeletedInitialPaths(ctx, rankingshared.DerivativeGraphKeyFromTime(time.Now()))
}

const vacuumBatchSize = 100 // TODO - configure via envvar

func vacuumAbandonedDefinitions(ctx context.Context, store store.Store) (int, error) {
	if enabled := conf.CodeIntelRankingDocumentReferenceCountsEnabled(); !enabled {
		return 0, nil
	}

	return store.VacuumAbandonedDefinitions(ctx, rankingshared.GraphKey(), vacuumBatchSize)
}

func vacuumAbandonedReferences(ctx context.Context, store store.Store) (int, error) {
	if enabled := conf.CodeIntelRankingDocumentReferenceCountsEnabled(); !enabled {
		return 0, nil
	}

	return store.VacuumAbandonedReferences(ctx, rankingshared.GraphKey(), vacuumBatchSize)
}

func vacuumAbandonedInitialPathCounts(ctx context.Context, store store.Store) (int, error) {
	if enabled := conf.CodeIntelRankingDocumentReferenceCountsEnabled(); !enabled {
		return 0, nil
	}

	return store.VacuumAbandonedInitialPathCounts(ctx, rankingshared.GraphKey(), vacuumBatchSize)
}

func vacuumStaleGraphs(ctx context.Context, store store.Store) (int, error) {
	if enabled := conf.CodeIntelRankingDocumentReferenceCountsEnabled(); !enabled {
		return 0, nil
	}

	return store.VacuumStaleGraphs(ctx, rankingshared.DerivativeGraphKeyFromTime(time.Now()), vacuumBatchSize)
}

func vacuumStaleRanks(ctx context.Context, store store.Store) (int, int, error) {
	if enabled := conf.CodeIntelRankingDocumentReferenceCountsEnabled(); !enabled {
		return 0, 0, nil
	}

	return store.VacuumStaleRanks(ctx, rankingshared.DerivativeGraphKeyFromTime(time.Now()))
}
