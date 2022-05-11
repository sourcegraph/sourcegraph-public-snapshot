package repomatcher

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/goroutine"
)

type matcher struct{}

var _ goroutine.Handler = &matcher{}
var _ goroutine.ErrorHandler = &matcher{}

func (r *matcher) Handle(ctx context.Context) error {
	// To be implemented in https://github.com/sourcegraph/sourcegraph/issues/33376
	return nil
}

func (r *matcher) HandleError(err error) {
}
