package codeintel

import (
	"context"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/worker/job"
	"github.com/sourcegraph/sourcegraph/cmd/worker/shared/init/codeintel"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/uploads"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/uploads/background/backfill"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
)

type uploadBackfillerJob struct{}

func NewUploadBackfillerJob() job.Job {
	return &uploadBackfillerJob{}
}

func (j *uploadBackfillerJob) Description() string {
	return ""
}

func (j *uploadBackfillerJob) Config() []env.Config {
	return []env.Config{
		backfill.ConfigInst,
	}
}

func (j *uploadBackfillerJob) Routines(startupCtx context.Context, logger log.Logger) ([]goroutine.BackgroundRoutine, error) {
	services, err := codeintel.InitServices()
	if err != nil {
		return nil, err
	}

	return backfill.NewCommittedAtBackfiller(uploads.GetBackgroundJob(services.UploadsService)), nil
}
