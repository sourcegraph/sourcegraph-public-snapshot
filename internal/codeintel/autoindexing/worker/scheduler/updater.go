package scheduler

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/goroutine"
)

type scheduler struct{}

var _ goroutine.Handler = &scheduler{}
var _ goroutine.ErrorHandler = &scheduler{}

func (r *scheduler) Handle(ctx context.Context) error {
	// TODO
	return nil
}

func (r *scheduler) HandleError(err error) {
	// TODO
}
