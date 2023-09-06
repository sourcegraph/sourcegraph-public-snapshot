package reducer

import (
	"context"

	rankingshared "github.com/sourcegraph/sourcegraph/internal/codeintel/ranking/internal/shared"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/ranking/internal/store"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/shared/background"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

func NewReducer(
	observationCtx *observation.Context,
	store store.Store,
	config *Config,
) goroutine.BackgroundRoutine {
	name := "codeintel.ranking.file-reference-count-reducer"

	return background.NewPipelineJob(context.Background(), background.PipelineOptions{
		Name:        name,
		Description: "Aggregates records from `codeintel_ranking_path_counts_inputs` into `codeintel_path_ranks`.",
		Interval:    config.Interval,
		Metrics:     background.NewPipelineMetrics(observationCtx, name),
		ProcessFunc: func(ctx context.Context) (numRecordsProcessed int, numRecordsAltered background.TaggedCounts, err error) {
			numPathCountInputsScanned, numRanksUpdated, err := reduceRankingGraph(ctx, store, config.BatchSize)
			return numPathCountInputsScanned, background.NewSingleCount(numRanksUpdated), err
		},
	})
}

func reduceRankingGraph(
	ctx context.Context,
	s store.Store,
	batchSize int,
) (numPathRanksInserted int, numPathCountInputsProcessed int, err error) {
	if enabled := conf.CodeIntelRankingDocumentReferenceCountsEnabled(); !enabled {
		return 0, 0, nil
	}

	derivativeGraphKeyPrefix, _, err := store.DerivativeGraphKey(ctx, s)
	if err != nil {
		return 0, 0, err
	}

	return s.InsertPathRanks(
		ctx,
		rankingshared.DerivativeGraphKeyFromPrefix(derivativeGraphKeyPrefix),
		batchSize,
	)
}
