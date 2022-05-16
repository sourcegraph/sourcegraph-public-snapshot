package repomatcher

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/goroutine"
)

func NewMatcher() goroutine.BackgroundRoutine {
	return goroutine.NewPeriodicGoroutine(context.Background(), ConfigInst.Interval, &matcher{})
}
