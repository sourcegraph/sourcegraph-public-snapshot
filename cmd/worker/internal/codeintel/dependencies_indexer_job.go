package codeintel

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/worker/job"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/dependencies/background/indexer"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/lib/log"
)

type dependenciesIndexerJob struct{}

func NewDependenciesIndexerJob() job.Job {
	return &dependenciesIndexerJob{}
}

func (j *dependenciesIndexerJob) Description() string {
	return ""
}

func (j *dependenciesIndexerJob) Config() []env.Config {
	return []env.Config{
		indexer.ConfigInst,
	}
}

func (j *dependenciesIndexerJob) Routines(ctx context.Context, logger log.Logger) ([]goroutine.BackgroundRoutine, error) {
	return []goroutine.BackgroundRoutine{
		indexer.NewIndexer(),
	}, nil
}
