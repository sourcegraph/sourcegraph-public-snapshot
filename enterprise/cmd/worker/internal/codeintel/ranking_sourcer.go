package codeintel

import (
	"context"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/worker/job"
	"github.com/sourcegraph/sourcegraph/cmd/worker/shared/init/codeintel"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/ranking/background/indexer"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/ranking/background/loader"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
)

type rankingSourcerJob struct{}

func NewRankingSourcerJob() job.Job {
	return &rankingSourcerJob{}
}

func (j *rankingSourcerJob) Description() string {
	return ""
}

func (j *rankingSourcerJob) Config() []env.Config {
	return []env.Config{
		indexer.ConfigInst,
		loader.ConfigInst,
	}
}

func (j *rankingSourcerJob) Routines(startupCtx context.Context, logger log.Logger) ([]goroutine.BackgroundRoutine, error) {
	services, err := codeintel.InitServices()
	if err != nil {
		return nil, err
	}

	return append(
		indexer.NewIndexer(services.RankingService),
		loader.NewPageRankLoader(services.RankingService)...,
	), nil
}
