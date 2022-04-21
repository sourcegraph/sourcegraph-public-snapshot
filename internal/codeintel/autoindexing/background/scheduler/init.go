package scheduler

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/goroutine"
)

func NewScheduler() goroutine.BackgroundRoutine {
	return goroutine.NewPeriodicGoroutine(context.Background(), ConfigInst.Interval, &scheduler{})
}
