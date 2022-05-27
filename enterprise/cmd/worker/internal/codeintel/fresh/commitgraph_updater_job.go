package codeintel

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/worker/job"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/uploads/background/commitgraph"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/lib/log"
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

func (j *commitGraphUpdaterJob) Routines(ctx context.Context, logger log.Logger) ([]goroutine.BackgroundRoutine, error) {
	return []goroutine.BackgroundRoutine{
		commitgraph.NewUpdater(),
	}, nil
}
