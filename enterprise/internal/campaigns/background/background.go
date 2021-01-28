package background

import (
	"context"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/campaigns/store"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/internal/repos"
)

func Routines(ctx context.Context, cstore *store.Store, cf *httpcli.Factory) []goroutine.BackgroundRoutine {
	sourcer := repos.NewSourcer(cf)

	metrics := newMetrics()

	routines := []goroutine.BackgroundRoutine{
		newWorker(ctx, cstore, gitserver.DefaultClient, sourcer, metrics),
		newWorkerResetter(cstore, metrics),
		newSpecExpireWorker(ctx, cstore),
	}
	return routines
}
