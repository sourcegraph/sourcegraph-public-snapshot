package janitor

import (
	"context"

	rankingshared "github.com/sourcegraph/sourcegraph/internal/codeintel/ranking/internal/shared"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/ranking/internal/store"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/shared/background"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

func NewExportedUploadsJanitor(
	observationCtx *observation.Context,
	store store.Store,
	config *Config,
) goroutine.BackgroundRoutine {
	name := "codeintel.ranking.exported-uploads-janitor"

	return background.NewJanitorJob(context.Background(), background.JanitorOptions{
		Name:        name,
		Description: "Soft-deletes stale data from the ranking exported uploads table.",
		Interval:    config.Interval,
		Metrics:     background.NewJanitorMetrics(observationCtx, name),
		CleanupFunc: func(ctx context.Context) (numRecordsScanned int, numRecordsAltered int, err error) {
			return softDeleteStaleExportedUploads(ctx, store)
		},
	})
}

func NewDeletedUploadsJanitor(
	observationCtx *observation.Context,
	store store.Store,
	config *Config,
) goroutine.BackgroundRoutine {
	name := "codeintel.ranking.deleted-exported-uploads-janitor"

	return background.NewJanitorJob(context.Background(), background.JanitorOptions{
		Name:        name,
		Description: "Removes soft-deleted data from the ranking exported uploads table no longer being read by a mapper process.",
		Interval:    config.Interval,
		Metrics:     background.NewJanitorMetrics(observationCtx, name),
		CleanupFunc: func(ctx context.Context) (numRecordsScanned int, numRecordsAltered int, err error) {
			numDeleted, err := vacuumDeletedExportedUploads(ctx, store)
			return numDeleted, numDeleted, err
		},
	})
}

func NewAbandonedExportedUploadsJanitor(
	observationCtx *observation.Context,
	store store.Store,
	config *Config,
) goroutine.BackgroundRoutine {
	name := "codeintel.ranking.abandoned-exported-uploads-janitor"

	return background.NewJanitorJob(context.Background(), background.JanitorOptions{
		Name:        name,
		Description: "Removes ranking exported uploads records for old graph keys.",
		Interval:    config.Interval,
		Metrics:     background.NewJanitorMetrics(observationCtx, name),
		CleanupFunc: func(ctx context.Context) (numRecordsScanned int, numRecordsAltered int, err error) {
			numDeleted, err := vacuumAbandonedExportedUploads(ctx, store)
			return numDeleted, numDeleted, err
		},
	})
}

func NewProcessedReferencesJanitor(
	observationCtx *observation.Context,
	store store.Store,
	config *Config,
) goroutine.BackgroundRoutine {
	name := "codeintel.ranking.processed-references-janitor"

	return background.NewJanitorJob(context.Background(), background.JanitorOptions{
		Name:        name,
		Description: "Removes old processed reference input records.",
		Interval:    config.Interval,
		Metrics:     background.NewJanitorMetrics(observationCtx, name),
		CleanupFunc: func(ctx context.Context) (numRecordsScanned int, numRecordsAltered int, err error) {
			numDeleted, err := vacuumStaleProcessedReferences(ctx, store)
			return numDeleted, numDeleted, err
		},
	})
}

func NewProcessedPathsJanitor(
	observationCtx *observation.Context,
	store store.Store,
	config *Config,
) goroutine.BackgroundRoutine {
	name := "codeintel.ranking.processed-paths-janitor"

	return background.NewJanitorJob(context.Background(), background.JanitorOptions{
		Name:        name,
		Description: "Removes old processed path input records.",
		Interval:    config.Interval,
		Metrics:     background.NewJanitorMetrics(observationCtx, name),
		CleanupFunc: func(ctx context.Context) (numRecordsScanned int, numRecordsAltered int, err error) {
			numDeleted, err := vacuumStaleProcessedPaths(ctx, store)
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
		Metrics:     background.NewJanitorMetrics(observationCtx, name),
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
		Metrics:     background.NewJanitorMetrics(observationCtx, name),
		CleanupFunc: func(ctx context.Context) (numRecordsScanned int, numRecordsAltered int, err error) {
			return vacuumStaleRanks(ctx, store)
		},
	})
}

func softDeleteStaleExportedUploads(ctx context.Context, store store.Store) (int, int, error) {
	if enabled := conf.CodeIntelRankingDocumentReferenceCountsEnabled(); !enabled {
		return 0, 0, nil
	}

	return store.SoftDeleteStaleExportedUploads(ctx, rankingshared.GraphKey())
}

func vacuumDeletedExportedUploads(ctx context.Context, s store.Store) (int, error) {
	if enabled := conf.CodeIntelRankingDocumentReferenceCountsEnabled(); !enabled {
		return 0, nil
	}

	derivativeGraphKeyPrefix, _, err := store.DerivativeGraphKey(ctx, s)
	if err != nil {
		return 0, err
	}

	return s.VacuumDeletedExportedUploads(ctx, rankingshared.DerivativeGraphKeyFromPrefix(derivativeGraphKeyPrefix))
}

const (
	vacuumUploadsBatchSize     = 100
	vacuumMiscRecordsBatchSize = 10000
)

func vacuumAbandonedExportedUploads(ctx context.Context, store store.Store) (int, error) {
	if enabled := conf.CodeIntelRankingDocumentReferenceCountsEnabled(); !enabled {
		return 0, nil
	}

	return store.VacuumAbandonedExportedUploads(ctx, rankingshared.GraphKey(), vacuumUploadsBatchSize)
}

func vacuumStaleProcessedReferences(ctx context.Context, s store.Store) (int, error) {
	if enabled := conf.CodeIntelRankingDocumentReferenceCountsEnabled(); !enabled {
		return 0, nil
	}

	derivativeGraphKeyPrefix, _, err := store.DerivativeGraphKey(ctx, s)
	if err != nil {
		return 0, err
	}

	return s.VacuumStaleProcessedReferences(ctx, rankingshared.DerivativeGraphKeyFromPrefix(derivativeGraphKeyPrefix), vacuumMiscRecordsBatchSize)
}

func vacuumStaleProcessedPaths(ctx context.Context, s store.Store) (int, error) {
	if enabled := conf.CodeIntelRankingDocumentReferenceCountsEnabled(); !enabled {
		return 0, nil
	}

	derivativeGraphKeyPrefix, _, err := store.DerivativeGraphKey(ctx, s)
	if err != nil {
		return 0, err
	}

	return s.VacuumStaleProcessedPaths(ctx, rankingshared.DerivativeGraphKeyFromPrefix(derivativeGraphKeyPrefix), vacuumMiscRecordsBatchSize)
}

func vacuumStaleGraphs(ctx context.Context, s store.Store) (int, error) {
	if enabled := conf.CodeIntelRankingDocumentReferenceCountsEnabled(); !enabled {
		return 0, nil
	}

	derivativeGraphKeyPrefix, _, err := store.DerivativeGraphKey(ctx, s)
	if err != nil {
		return 0, err
	}

	return s.VacuumStaleGraphs(ctx, rankingshared.DerivativeGraphKeyFromPrefix(derivativeGraphKeyPrefix), vacuumMiscRecordsBatchSize)
}

func vacuumStaleRanks(ctx context.Context, s store.Store) (int, int, error) {
	if enabled := conf.CodeIntelRankingDocumentReferenceCountsEnabled(); !enabled {
		return 0, 0, nil
	}

	derivativeGraphKeyPrefix, _, err := store.DerivativeGraphKey(ctx, s)
	if err != nil {
		return 0, 0, err
	}

	return s.VacuumStaleRanks(ctx, rankingshared.DerivativeGraphKeyFromPrefix(derivativeGraphKeyPrefix))
}
