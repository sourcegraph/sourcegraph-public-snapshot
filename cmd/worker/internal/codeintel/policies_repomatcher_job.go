package codeintel

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/worker/job"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/policies/background/repomatcher"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
)

type policiesRepositoryMatcherJob struct{}

func NewPoliciesRepositoryMatcherJob() job.Job {
	return &policiesRepositoryMatcherJob{}
}

func (j *policiesRepositoryMatcherJob) Config() []env.Config {
	return []env.Config{
		repomatcher.ConfigInst,
	}
}

func (j *policiesRepositoryMatcherJob) Routines(ctx context.Context) ([]goroutine.BackgroundRoutine, error) {
	return []goroutine.BackgroundRoutine{
		repomatcher.NewMatcher(),
	}, nil
}
