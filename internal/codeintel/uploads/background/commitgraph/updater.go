package commitgraph

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/goroutine"
)

type updater struct{}

var _ goroutine.Handler = &updater{}
var _ goroutine.ErrorHandler = &updater{}

func (r *updater) Handle(ctx context.Context) error {
	// To be implemented in https://github.com/sourcegraph/sourcegraph/issues/33375
	return nil
}

func (r *updater) HandleError(err error) {
}
