package repomatcher

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/goroutine"
)

func NewMatcher(policySvc PolicyService, metrics *metrics) goroutine.BackgroundRoutine {
	return goroutine.NewPeriodicGoroutine(context.Background(), ConfigInst.Interval, &matcher{
		policySvc: policySvc,
		metrics:   metrics,
	})
}
