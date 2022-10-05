package codeintel

import (
	"context"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/worker/job"
	"github.com/sourcegraph/sourcegraph/cmd/worker/shared/init/codeintel"
	workerdb "github.com/sourcegraph/sourcegraph/cmd/worker/shared/init/db"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/uploads"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/uploads/background/commitgraph"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
)

type commitGraphUpdaterJob struct{}

func NewCommitGraphUpdaterJob() job.Job {
	return &commitGraphUpdaterJob{}
}

func (j *commitGraphUpdaterJob) Description() string {
	return ""
}

func (j *commitGraphUpdaterJob) Config() []env.Config {
	return []env.Config{
		commitgraph.ConfigInst,
	}
}

func (j *commitGraphUpdaterJob) Routines(startupCtx context.Context, logger log.Logger) ([]goroutine.BackgroundRoutine, error) {
	rawDB, err := workerdb.Init()
	if err != nil {
		return nil, err
	}
	db := database.NewDB(logger, rawDB)

	rawCodeIntelDB, err := codeintel.InitCodeIntelDatabase()
	if err != nil {
		return nil, err
	}
	codeIntelDB := database.NewDB(logger, rawCodeIntelDB)

	gitserver, err := codeintel.InitGitserverClient()
	if err != nil {
		return nil, err
	}

	uploadSvc := uploads.GetService(db, codeIntelDB, gitserver)

	return commitgraph.NewUpdater(uploadSvc), nil
}
