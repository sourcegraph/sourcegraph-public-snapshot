package timeutil

import (
	"context"
	"time"
)

// SleepWithContext is time.Sleep but context-aware. If the given context is
// canceled, it possibly returns before d has passed. It cleans up the
// time.After goroutine.
func SleepWithContext(ctx context.Context, d time.Duration) {
	t := time.NewTimer(d)
	select {
	case <-ctx.Done():
		// See documentation for t.Stop()
		if !t.Stop() {
			<-t.C
		}
	case <-t.C:
	}
}
