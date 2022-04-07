package repomatcher

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/goroutine"
)

func NewRepositoryMatcher() goroutine.BackgroundRoutine {
	return goroutine.NewPeriodicGoroutine(context.Background(), ConfigInst.Interval, &repositoryMatcher{
		// TODO
	})
}
