package codeintel

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/worker/job"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/policies/background/repomatcher"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/lib/log"
)

type policiesRepositoryMatcherJob struct{}

func NewPoliciesRepositoryMatcherJob() job.Job {
	return &policiesRepositoryMatcherJob{}
}

func (j *policiesRepositoryMatcherJob) Description() string {
	return ""
}

func (j *policiesRepositoryMatcherJob) Config() []env.Config {
	return []env.Config{
		repomatcher.ConfigInst,
	}
}

func (j *policiesRepositoryMatcherJob) Routines(ctx context.Context, logger log.Logger) ([]goroutine.BackgroundRoutine, error) {
	return []goroutine.BackgroundRoutine{
		repomatcher.NewMatcher(),
	}, nil
}
