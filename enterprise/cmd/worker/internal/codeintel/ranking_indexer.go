package codeintel

import (
	"context"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/worker/job"
	"github.com/sourcegraph/sourcegraph/cmd/worker/shared/init/codeintel"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/ranking/background/indexer"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
)

type rankingIndexerJob struct{}

func NewRankingIndexerJob() job.Job {
	return &rankingIndexerJob{}
}

func (j *rankingIndexerJob) Description() string {
	return ""
}

func (j *rankingIndexerJob) Config() []env.Config {
	return []env.Config{
		indexer.ConfigInst,
	}
}

func (j *rankingIndexerJob) Routines(startupCtx context.Context, logger log.Logger) ([]goroutine.BackgroundRoutine, error) {
	services, err := codeintel.InitServices()
	if err != nil {
		return nil, err
	}

	return indexer.NewIndexer(services.RankingService), nil
}
