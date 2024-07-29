package reconciler

import (
	"context"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
)

func New(ctx context.Context, db database.DB, interval time.Duration) goroutine.BackgroundRoutine {
	return goroutine.NewPeriodicGoroutine(
		actor.WithInternalActor(ctx),
		goroutine.HandlerFunc(func(ctx context.Context) error {
			return nil
		}),
		goroutine.WithName("tenantmanager.reconciler"),
		goroutine.WithDescription("syncs tenant state with the SoT with this instance"),
		goroutine.WithInterval(interval),
	)
}
