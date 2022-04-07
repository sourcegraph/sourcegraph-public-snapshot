package expiration

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/goroutine"
)

type expirer struct{}

var _ goroutine.Handler = &expirer{}
var _ goroutine.ErrorHandler = &expirer{}

func (r *expirer) Handle(ctx context.Context) error {
	// TODO
	return nil
}

func (r *expirer) HandleError(err error) {
	// TODO
}
