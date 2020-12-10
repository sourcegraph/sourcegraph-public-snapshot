package background

import (
	"context"
	"database/sql"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/campaigns/store"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/internal/repos"
)

func StartBackgroundJobs(ctx context.Context, db *sql.DB, campaignsStore *store.Store, repoStore repos.Store, cf *httpcli.Factory) {
	sourcer := repos.NewSourcer(cf)

	metrics := newMetrics()

	routines := []goroutine.BackgroundRoutine{
		newWorker(ctx, campaignsStore, gitserver.DefaultClient, sourcer, metrics),
		newWorkerResetter(campaignsStore, metrics),
		newSpecExpireWorker(ctx, campaignsStore),
	}

	go goroutine.MonitorBackgroundRoutines(ctx, routines...)
}
