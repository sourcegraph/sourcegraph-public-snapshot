package codeintel

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/worker/job"
	"github.com/sourcegraph/sourcegraph/cmd/worker/workerdb"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/dependencies/background/indexer"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/dependencies/background/resolver"
	livedependencies "github.com/sourcegraph/sourcegraph/internal/codeintel/dependencies/live"
	"github.com/sourcegraph/sourcegraph/internal/database"
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
		resolver.ConfigInst,
	}
}

func (j *dependenciesIndexerJob) Routines(ctx context.Context, logger log.Logger) ([]goroutine.BackgroundRoutine, error) {
	db, err := workerdb.Init()
	if err != nil {
		return nil, err
	}

	return []goroutine.BackgroundRoutine{
		indexer.NewIndexer(),
		resolver.NewResolver(database.NewDB(db), livedependencies.NewSyncer()),
	}, nil
}
