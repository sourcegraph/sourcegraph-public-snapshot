package scheduler

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/goroutine"
)

type scheduler struct{}

var _ goroutine.Handler = &scheduler{}
var _ goroutine.ErrorHandler = &scheduler{}

func (r *scheduler) Handle(ctx context.Context) error {
	// To be implemented in https://github.com/sourcegraph/sourcegraph/pull/33614
	return nil
}

func (r *scheduler) HandleError(err error) {
}
