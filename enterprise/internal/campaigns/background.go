package campaigns

import (
	"context"
	"database/sql"

	"github.com/sourcegraph/sourcegraph/cmd/repo-updater/repoupdater"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/campaigns/background"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/campaigns/store"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/campaigns/syncer"
	ossDB "github.com/sourcegraph/sourcegraph/internal/db"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/internal/repos"
	"github.com/sourcegraph/sourcegraph/internal/timeutil"
)

// InitBackgroundJobs starts all jobs required to run campaigns. Currently, it is called from
// repo-updater and in the future will be the main entry point for the campaigns worker.
func InitBackgroundJobs(
	ctx context.Context,
	db *sql.DB,
	rstore repos.Store,
	cf *httpcli.Factory,
	// TODO(eseliger): Remove this parameter as the sunset of repo-updater is approaching.
	// We should switch to our own polling mechanism instead of using repo-updaters.
	server *repoupdater.Server,
) {
	cstore := store.NewWithClock(db, timeutil.Now)

	repoStore := ossDB.NewRepoStoreWith(cstore)
	esStore := ossDB.NewExternalServicesStoreWith(cstore)

	syncRegistry := syncer.NewSyncRegistry(ctx, cstore, repoStore, esStore, cf)
	if server != nil {
		server.ChangesetSyncRegistry = syncRegistry
	}

	go goroutine.MonitorBackgroundRoutines(ctx, background.Routines(ctx, db, cstore, rstore, cf)...)
}
