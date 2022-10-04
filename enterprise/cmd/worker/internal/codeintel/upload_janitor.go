package codeintel

import (
	"context"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/worker/job"
	"github.com/sourcegraph/sourcegraph/cmd/worker/shared/init/codeintel"
	workerdb "github.com/sourcegraph/sourcegraph/cmd/worker/shared/init/db"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/uploads"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/uploads/background/cleanup"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
)

type uploadJanitorJob struct{}

func NewUploadJanitorJob() job.Job {
	return &uploadJanitorJob{}
}

func (j *uploadJanitorJob) Description() string {
	return ""
}

func (j *uploadJanitorJob) Config() []env.Config {
	return []env.Config{
		cleanup.ConfigInst,
	}
}

func (j *uploadJanitorJob) Routines(startupCtx context.Context, logger log.Logger) ([]goroutine.BackgroundRoutine, error) {
	// Initialize stores
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

	// Initialize clients
	gitserverClient, err := codeintel.InitGitserverClient()
	if err != nil {
		return nil, err
	}

	// Initialize services
	uploadSvc := uploads.GetService(db, codeIntelDB, gitserverClient)
	resetters := cleanup.NewResetters(uploadSvc)

	return append(cleanup.NewJanitor(uploadSvc), resetters...), nil
}
