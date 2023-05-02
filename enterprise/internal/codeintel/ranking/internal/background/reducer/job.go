package reducer

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
		Metrics:     background.NewPipelineMetrics(observationCtx, name, recordTypeName),
		ProcessFunc: func(ctx context.Context) (numRecordsProcessed int, numRecordsAltered background.TaggedCounts, err error) {
			numPathCountInputsScanned, numRanksUpdated, err := reduceRankingGraph(ctx, store, config.BatchSize)
			return numPathCountInputsScanned, background.NewSingleCount(numRanksUpdated), err
		},
	})
}

func reduceRankingGraph(
	ctx context.Context,
	store store.Store,
	batchSize int,
) (numPathRanksInserted int, numPathCountInputsProcessed int, err error) {
	if enabled := conf.CodeIntelRankingDocumentReferenceCountsEnabled(); !enabled {
		return 0, 0, nil
	}

	return store.InsertPathRanks(
		ctx,
		rankingshared.DerivativeGraphKeyFromTime(time.Now()),
		batchSize,
	)
}
