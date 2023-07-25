package codeintel

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/worker/job"
	"github.com/sourcegraph/sourcegraph/cmd/worker/shared/init/codeintel"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/ranking"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type rankingJob struct{}

func NewRankingFileReferenceCounter() job.Job {
	return &rankingJob{}
}

func (j *rankingJob) Description() string {
	return ""
}

func (j *rankingJob) Config() []env.Config {
	return []env.Config{
		ranking.ExporterConfigInst,
		ranking.CoordinatorConfigInst,
		ranking.MapperConfigInst,
		ranking.ReducerConfigInst,
		ranking.JanitorConfigInst,
	}
}

func (j *rankingJob) Routines(_ context.Context, observationCtx *observation.Context) ([]goroutine.BackgroundRoutine, error) {
	services, err := codeintel.InitServices(observationCtx)
	if err != nil {
		return nil, err
	}

	routines := []goroutine.BackgroundRoutine{}
	routines = append(routines, ranking.NewSymbolExporter(observationCtx, services.RankingService))
	routines = append(routines, ranking.NewCoordinator(observationCtx, services.RankingService))
	routines = append(routines, ranking.NewMapper(observationCtx, services.RankingService)...)
	routines = append(routines, ranking.NewReducer(observationCtx, services.RankingService))
	routines = append(routines, ranking.NewSymbolJanitor(observationCtx, services.RankingService)...)
	return routines, nil
}
