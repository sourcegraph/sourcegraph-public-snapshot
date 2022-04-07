package repomatcher

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/goroutine"
)

type repositoryMatcher struct{}

var _ goroutine.Handler = &repositoryMatcher{}
var _ goroutine.ErrorHandler = &repositoryMatcher{}

func (r *repositoryMatcher) Handle(ctx context.Context) error {
	// TODO
	return nil
}

func (r *repositoryMatcher) HandleError(err error) {
	// TODO
}
