package scheduler

import (
	"context"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/autoindexing"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
)

type onDemandScheduler struct {
	autoindexingSvc *autoindexing.Service
	batchSize       int
	logger          log.Logger
}

var (
	_ goroutine.Handler      = &scheduler{}
	_ goroutine.ErrorHandler = &scheduler{}
)

func (s *onDemandScheduler) Handle(ctx context.Context) error {
	if !autoIndexingEnabled() {
		return nil
	}

	return s.autoindexingSvc.ProcessRepoRevs(ctx, s.batchSize)
}

func (s *onDemandScheduler) HandleError(err error) {
	s.logger.Error("Failed to schedule on-demand index jobs", log.Error(err))
}
