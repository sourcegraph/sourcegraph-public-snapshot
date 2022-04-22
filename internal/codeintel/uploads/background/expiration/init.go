package expiration

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/goroutine"
)

func NewExpirer() goroutine.BackgroundRoutine {
	return goroutine.NewPeriodicGoroutine(context.Background(), ConfigInst.Interval, &expirer{})
}
