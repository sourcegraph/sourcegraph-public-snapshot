package codeintel

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/worker/job"
	"github.com/sourcegraph/sourcegraph/cmd/worker/shared/init/codeintel"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/sentinel"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type sentinelCVEScannerJob struct{}

func NewSentinelCVEScannerJob() job.Job {
	return &sentinelCVEScannerJob{}
}

func (j *sentinelCVEScannerJob) Description() string {
	return "code-intel sentinel vulnerability scanner"
}

func (j *sentinelCVEScannerJob) Config() []env.Config {
	return []env.Config{
		sentinel.DownloaderConfigInst,
		sentinel.MatcherConfigInst,
	}
}

func (j *sentinelCVEScannerJob) Routines(_ context.Context, observationCtx *observation.Context) ([]goroutine.BackgroundRoutine, error) {
	services, err := codeintel.InitServices(observationCtx)
	if err != nil {
		return nil, err
	}

	return sentinel.CVEScannerJob(observationCtx, services.SentinelService), nil
}
