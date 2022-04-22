package codeintel

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/worker/job"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/uploads/background/expiration"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
)

type uploadExpirerJob struct{}

func NewUploadExpirerJob() job.Job {
	return &uploadExpirerJob{}
}

func (j *uploadExpirerJob) Config() []env.Config {
	return []env.Config{
		expiration.ConfigInst,
	}
}

func (j *uploadExpirerJob) Routines(ctx context.Context) ([]goroutine.BackgroundRoutine, error) {
	return []goroutine.BackgroundRoutine{
		expiration.NewExpirer(),
	}, nil
}
