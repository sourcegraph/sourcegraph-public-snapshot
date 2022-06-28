package scheduler

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/autoindexing"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
)

type scheduler struct {
	autoindexingSvc *autoindexing.Service
	dbStore         DBStore
	policyMatcher   PolicyMatcher
}

var _ goroutine.Handler = &scheduler{}
var _ goroutine.ErrorHandler = &scheduler{}

func (r *scheduler) Handle(ctx context.Context) error {
	return r.handle(ctx)
}

func (r *scheduler) HandleError(err error) {
}
