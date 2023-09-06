package mapper

import (
	"context"

	rankingshared "github.com/sourcegraph/sourcegraph/internal/codeintel/ranking/internal/shared"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/ranking/internal/store"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/shared/background"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

func NewMapper(
	observationCtx *observation.Context,
	store store.Store,
	config *Config,
) goroutine.BackgroundRoutine {
	name := "codeintel.ranking.file-reference-count-mapper"

	return background.NewPipelineJob(context.Background(), background.PipelineOptions{
		Name:        name,
		Description: "Joins ranking definition and references together to create document path count records.",
		Interval:    config.Interval,
		Metrics:     background.NewPipelineMetrics(observationCtx, name),
		ProcessFunc: func(ctx context.Context) (numRecordsProcessed int, numRecordsAltered background.TaggedCounts, err error) {
			numReferencesScanned, nuPathCountInputsInserted, err := mapRankingGraph(ctx, store, config.BatchSize)
			if err != nil {
				return 0, nil, err
			}

			return numReferencesScanned, background.NewSingleCount(nuPathCountInputsInserted), err
		},
	})
}

func NewSeedMapper(
	observationCtx *observation.Context,
	store store.Store,
	config *Config,
) goroutine.BackgroundRoutine {
	name := "codeintel.ranking.file-reference-count-seed-mapper"

	return background.NewPipelineJob(context.Background(), background.PipelineOptions{
		Name:        name,
		Description: "Adds initial zero counts to files that may not contain any known references.",
		Interval:    config.Interval,
		Metrics:     background.NewPipelineMetrics(observationCtx, name),
		ProcessFunc: func(ctx context.Context) (numRecordsProcessed int, numRecordsAltered background.TaggedCounts, err error) {
			numInitialPathsScanned, nuPathCountInputsInserted, err := mapInitializerRankingGraph(ctx, store, config.BatchSize)
			if err != nil {
				return 0, nil, err
			}

			return numInitialPathsScanned, background.NewSingleCount(nuPathCountInputsInserted), err
		},
	})
}

func mapInitializerRankingGraph(
	ctx context.Context,
	s store.Store,
	batchSize int,
) (
	numInitialPathsProcessed int,
	numInitialPathRanksInserted int,
	err error,
) {
	if enabled := conf.CodeIntelRankingDocumentReferenceCountsEnabled(); !enabled {
		return 0, 0, nil
	}

	derivativeGraphKeyPrefix, _, err := store.DerivativeGraphKey(ctx, s)
	if err != nil {
		return 0, 0, err
	}

	return s.InsertInitialPathCounts(
		ctx,
		rankingshared.DerivativeGraphKeyFromPrefix(derivativeGraphKeyPrefix),
		batchSize,
	)
}

func mapRankingGraph(
	ctx context.Context,
	s store.Store,
	batchSize int,
) (numReferenceRecordsProcessed int, numInputsInserted int, err error) {
	if enabled := conf.CodeIntelRankingDocumentReferenceCountsEnabled(); !enabled {
		return 0, 0, nil
	}

	derivativeGraphKeyPrefix, _, err := store.DerivativeGraphKey(ctx, s)
	if err != nil {
		return 0, 0, err
	}

	return s.InsertPathCountInputs(
		ctx,
		rankingshared.DerivativeGraphKeyFromPrefix(derivativeGraphKeyPrefix),
		batchSize,
	)
}
