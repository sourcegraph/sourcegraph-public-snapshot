package cleanup

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/goroutine"
)

type janitor struct{}

var _ goroutine.Handler = &janitor{}
var _ goroutine.ErrorHandler = &janitor{}

func (r *janitor) Handle(ctx context.Context) error {
	// TODO
	return nil
}

func (r *janitor) HandleError(err error) {
	// TODO
}
