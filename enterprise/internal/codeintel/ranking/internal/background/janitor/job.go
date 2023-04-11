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
		Description: "Removes stale data from the ranking definitions table.",
		Interval:    config.Interval,
		Metrics:     background.NewJanitorMetrics(observationCtx, name, recordTypeName),
		CleanupFunc: func(ctx context.Context) (numRecordsScanned int, numRecordsAltered int, err error) {
			return vacuumStaleDefinitions(ctx, store)
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
		Description: "Removes stale data from the ranking references table.",
		Interval:    config.Interval,
		Metrics:     background.NewJanitorMetrics(observationCtx, name, recordTypeName),
		CleanupFunc: func(ctx context.Context) (numRecordsScanned int, numRecordsAltered int, err error) {
			return vacuumStaleReferences(ctx, store)
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
		Description: "Removes stale data from the ranking initial paths table.",
		Interval:    config.Interval,
		Metrics:     background.NewJanitorMetrics(observationCtx, name, recordTypeName),
		CleanupFunc: func(ctx context.Context) (numRecordsScanned int, numRecordsAltered int, err error) {
			return vacuumStaleInitialPaths(ctx, store)
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

func vacuumStaleDefinitions(ctx context.Context, store store.Store) (int, int, error) {
	if enabled := conf.CodeIntelRankingDocumentReferenceCountsEnabled(); !enabled {
		return 0, 0, nil
	}

	numDefinitionRecordsScanned, numDefinitionRecordsRemoved, err := store.VacuumStaleDefinitions(ctx, rankingshared.GraphKey())
	return numDefinitionRecordsScanned, numDefinitionRecordsRemoved, err
}

func vacuumStaleReferences(ctx context.Context, store store.Store) (int, int, error) {
	if enabled := conf.CodeIntelRankingDocumentReferenceCountsEnabled(); !enabled {
		return 0, 0, nil
	}

	numReferenceRecordsScanned, numReferenceRecordsRemoved, err := store.VacuumStaleReferences(ctx, rankingshared.GraphKey())
	return numReferenceRecordsScanned, numReferenceRecordsRemoved, err
}

func vacuumStaleInitialPaths(ctx context.Context, store store.Store) (int, int, error) {
	if enabled := conf.CodeIntelRankingDocumentReferenceCountsEnabled(); !enabled {
		return 0, 0, nil
	}

	numPathRecordsScanned, numStalePathRecordsDeleted, err := store.VacuumStaleInitialPaths(ctx, rankingshared.GraphKey())
	return numPathRecordsScanned, numStalePathRecordsDeleted, err
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
