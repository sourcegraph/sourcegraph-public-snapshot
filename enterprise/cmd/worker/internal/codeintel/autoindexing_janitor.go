package codeintel

import (
	"context"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/worker/job"
	"github.com/sourcegraph/sourcegraph/cmd/worker/shared/init/codeintel"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/autoindexing/background/cleanup"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
)

type autoindexingJanitorJob struct{}

func NewAutoindexingJanitorJob() job.Job {
	return &autoindexingJanitorJob{}
}

func (j *autoindexingJanitorJob) Description() string {
	return ""
}

func (j *autoindexingJanitorJob) Config() []env.Config {
	return []env.Config{
		cleanup.ConfigInst,
	}
}

func (j *autoindexingJanitorJob) Routines(startupCtx context.Context, logger log.Logger) ([]goroutine.BackgroundRoutine, error) {
	services, err := codeintel.InitServices()
	if err != nil {
		return nil, err
	}

	return append(
		cleanup.NewJanitor(services.AutoIndexingService),
		cleanup.NewResetters(services.AutoIndexingService)...,
	), nil
}
