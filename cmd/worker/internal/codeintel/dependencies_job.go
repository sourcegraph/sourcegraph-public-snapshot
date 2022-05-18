package codeintel

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/worker/job"
	workerdb "github.com/sourcegraph/sourcegraph/cmd/worker/shared/init/db"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/dependencies/background/indexer"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/dependencies/background/resolver"
	livedependencies "github.com/sourcegraph/sourcegraph/internal/codeintel/dependencies/live"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/lib/log"
)

type dependenciesJob struct{}

func NewDependenciesJob() job.Job {
	return &dependenciesJob{}
}

func (j *dependenciesJob) Description() string {
	return ""
}

func (j *dependenciesJob) Config() []env.Config {
	return []env.Config{
		indexer.ConfigInst,
		resolver.ConfigInst,
	}
}

func (j *dependenciesJob) Routines(ctx context.Context, logger log.Logger) ([]goroutine.BackgroundRoutine, error) {
	db, err := workerdb.Init()
	if err != nil {
		return nil, err
	}

	return []goroutine.BackgroundRoutine{
		indexer.NewIndexer(),
		resolver.NewResolver(database.NewDB(db), livedependencies.NewSyncer()),
	}, nil
}
