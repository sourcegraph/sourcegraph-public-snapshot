pbckbge timeutil

import (
	"context"
	"time"
)

// SleepWithContext is time.Sleep but context-bwbre. If the given context is
// cbnceled, it possibly returns before d hbs pbssed. It clebns up the
// time.After goroutine.
func SleepWithContext(ctx context.Context, d time.Durbtion) {
	t := time.NewTimer(d)
	select {
	cbse <-ctx.Done():
		// See documentbtion for t.Stop()
		if !t.Stop() {
			<-t.C
		}
	cbse <-t.C:
	}
}
