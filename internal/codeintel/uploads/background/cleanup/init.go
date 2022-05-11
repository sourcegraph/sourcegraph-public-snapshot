package cleanup

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/goroutine"
)

func NewJanitor() goroutine.BackgroundRoutine {
	return goroutine.NewPeriodicGoroutine(context.Background(), ConfigInst.Interval, &janitor{})
}
