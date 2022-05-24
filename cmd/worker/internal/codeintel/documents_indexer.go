package codeintel

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/worker/job"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/documents/background/indexer"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/lib/log"
)

type documentsIndexerJob struct{}

func NewDocumentsIndexerJob() job.Job {
	return &documentsIndexerJob{}
}

func (j *documentsIndexerJob) Description() string {
	return ""
}

func (j *documentsIndexerJob) Config() []env.Config {
	return []env.Config{
		indexer.ConfigInst,
	}
}

func (j *documentsIndexerJob) Routines(ctx context.Context, logger log.Logger) ([]goroutine.BackgroundRoutine, error) {
	return []goroutine.BackgroundRoutine{
		indexer.NewIndexer(),
	}, nil
}
