package batches

import (
	"context"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/batches/store"
	"github.com/sourcegraph/sourcegraph/internal/batches/syncer"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/encryption"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

// InitBackgroundJobs starts all jobs required to run batches. Currently, it is called from
// repo-updater and in the future will be the main entry point for the batch changes worker.
func InitBackgroundJobs(
	ctx context.Context,
	db database.DB,
	key encryption.Key,
	cf *httpcli.Factory,
) syncer.ChangesetSyncRegistry {
	// We use an internal actor so that we can freely load dependencies from
	// the database without repository permissions being enforced.
	// We do check for repository permissions consciously in the Rewirer when
	// creating new changesets and in the executor, when talking to the code
	// host, we manually check for BatchChangesCredentials.
	ctx = actor.WithInternalActor(ctx)

	observationCtx := observation.NewContext(log.Scoped("batches.background"))
	bstore := store.New(db, observationCtx, key)

	syncRegistry := syncer.NewSyncRegistry(ctx, observationCtx, bstore, cf)

	go goroutine.MonitorBackgroundRoutines(ctx, syncRegistry)

	return syncRegistry
}
