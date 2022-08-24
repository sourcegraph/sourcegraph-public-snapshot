package codeintel

import (
	"context"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/worker/job"
	workerdb "github.com/sourcegraph/sourcegraph/cmd/worker/shared/init/db"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/dependencies/background/cratesyncer"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
)

type cratesSyncerJob struct{}

func NewCratesSyncerJob() job.Job { return &cratesSyncerJob{} }

func (j *cratesSyncerJob) Description() string  { return "" }
func (j *cratesSyncerJob) Config() []env.Config { return nil }

func (j *cratesSyncerJob) Routines(ctx context.Context, logger log.Logger) ([]goroutine.BackgroundRoutine, error) {
	db, err := workerdb.Init()
	if err != nil {
		return nil, err
	}

	return []goroutine.BackgroundRoutine{
		cratesyncer.NewCratesSyncer(database.NewDB(logger, db)),
	}, nil
}
