package codeintel

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/worker/job"
	workerdb "github.com/sourcegraph/sourcegraph/cmd/worker/shared/init/db"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/dependencies"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type packageFilterApplicatorJob struct{}

func NewPackagesFilterApplicatorJob() job.Job {
	return &packageFilterApplicatorJob{}
}

func (j *packageFilterApplicatorJob) Description() string {
	return "package repo filters applicator"
}

func (j *packageFilterApplicatorJob) Config() []env.Config {
	return nil
}

func (j *packageFilterApplicatorJob) Routines(_ context.Context, observationCtx *observation.Context) ([]goroutine.BackgroundRoutine, error) {
	db, err := workerdb.InitDB(observationCtx)
	if err != nil {
		return nil, err
	}

	return dependencies.PackageFiltersJob(observationCtx, db), nil
}
