package cleanup

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/goroutine"
)

type janitor struct{}

var _ goroutine.Handler = &janitor{}
var _ goroutine.ErrorHandler = &janitor{}

func (r *janitor) Handle(ctx context.Context) error {
	// To be implemented in https://github.com/sourcegraph/sourcegraph/issues/33375
	return nil
}

func (r *janitor) HandleError(err error) {
}
