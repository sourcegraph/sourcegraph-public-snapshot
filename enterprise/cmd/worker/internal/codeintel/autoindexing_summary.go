package codeintel

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/worker/job"
	"github.com/sourcegraph/sourcegraph/cmd/worker/shared/init/codeintel"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/autoindexing"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"

	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type autoindexingSummaryBuilder struct{}

func NewAutoindexingSummaryBuilder() job.Job {
	return &autoindexingSummaryBuilder{}
}

func (j *autoindexingSummaryBuilder) Description() string {
	return ""
}

func (j *autoindexingSummaryBuilder) Config() []env.Config {
	return []env.Config{
		autoindexing.SummaryConfigInst,
	}
}

func (j *autoindexingSummaryBuilder) Routines(_ context.Context, observationCtx *observation.Context) ([]goroutine.BackgroundRoutine, error) {
	services, err := codeintel.InitServices(observationCtx)
	if err != nil {
		return nil, err
	}

	return autoindexing.NewSummaryBuilder(
		observationCtx,
		services.AutoIndexingService,
		services.UploadsService,
	), nil
}
