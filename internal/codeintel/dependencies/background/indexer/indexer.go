package indexer

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/goroutine"
)

type indexer struct{}

var _ goroutine.Handler = &indexer{}
var _ goroutine.ErrorHandler = &indexer{}

func (r *indexer) Handle(ctx context.Context) error {
	// To be implemented in https://github.com/sourcegraph/sourcegraph/issues/31643
	return nil
}

func (r *indexer) HandleError(err error) {
}
