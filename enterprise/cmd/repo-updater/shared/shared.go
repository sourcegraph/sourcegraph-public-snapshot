package shared

import (
	"context"
	"os"
	"strconv"
	"time"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/envvar"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/globals"
	"github.com/sourcegraph/sourcegraph/cmd/repo-updater/repoupdater"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/repo-updater/internal/authz"
	frontendAuthz "github.com/sourcegraph/sourcegraph/enterprise/internal/authz"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/batches"
	edb "github.com/sourcegraph/sourcegraph/enterprise/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/api"
	ossAuthz "github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/authz/permssync"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	ossDB "github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/debugserver"
	"github.com/sourcegraph/sourcegraph/internal/encryption/keyring"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/ratelimit"
	"github.com/sourcegraph/sourcegraph/internal/repos"
	"github.com/sourcegraph/sourcegraph/internal/timeutil"
)

func EnterpriseInit(
	observationCtx *observation.Context,
	db ossDB.DB,
	repoStore repos.Store,
	keyring keyring.Ring,
	cf *httpcli.Factory,
	server *repoupdater.Server,
) (debugDumpers map[string]debugserver.Dumper, enqueueRepoPermsJob func(context.Context, api.RepoID, string) error) {
	debug, _ := strconv.ParseBool(os.Getenv("DEBUG"))
	if debug {
		observationCtx.Logger.Info("enterprise edition")
	}
	// NOTE: Internal actor is required to have full visibility of the repo table
	// 	(i.e. bypass repository authorization).
	ctx := actor.WithInternalActor(context.Background())

	// No Batch Changes on dotcom, so we don't need to spawn the
	// background jobs for this feature.
	if !envvar.SourcegraphDotComMode() {
		syncRegistry := batches.InitBackgroundJobs(ctx, db, keyring.BatchChangesCredentialKey, cf)
		if server != nil {
			server.ChangesetSyncRegistry = syncRegistry
		}
	}

	permsStore := edb.Perms(observationCtx.Logger, db, timeutil.Now)
	permsSyncer := authz.NewPermsSyncer(observationCtx.Logger.Scoped("PermsSyncer", "repository and user permissions syncer"), db, repoStore, permsStore, timeutil.Now, ratelimit.DefaultRegistry)

	permsJobStore := db.PermissionSyncJobs()
	enqueueRepoPermsJob = func(ctx context.Context, repo api.RepoID, reason string) error {
		if authz.PermissionSyncingDisabled() {
			return nil
		}

		// If the feature flag is enabled, create job...
		if permssync.PermissionSyncWorkerEnabled(ctx, db, observationCtx.Logger) {
			opts := ossDB.PermissionSyncJobOpts{HighPriority: true, Reason: reason}
			return permsJobStore.CreateRepoSyncJob(ctx, repo, opts)
		}
		// ... otherwise, we just call the PermsSyncer
		permsSyncer.ScheduleRepos(ctx, repo)
		return nil
	}

	if server != nil {
		server.PermsSyncer = permsSyncer
	}

	workerStore := authz.MakeStore(observationCtx, db.Handle())
	worker := authz.MakeWorker(ctx, observationCtx, workerStore, permsSyncer)
	resetter := authz.MakeResetter(observationCtx, workerStore)

	go goroutine.MonitorBackgroundRoutines(ctx, worker, resetter)

	go startBackgroundPermsSync(ctx, permsSyncer, db)

	return map[string]debugserver.Dumper{"repoPerms": permsSyncer}, enqueueRepoPermsJob
}

// startBackgroundPermsSync sets up background permissions syncing.
func startBackgroundPermsSync(ctx context.Context, syncer *authz.PermsSyncer, db ossDB.DB) {
	globals.WatchPermissionsUserMapping()
	go func() {
		t := time.NewTicker(frontendAuthz.RefreshInterval())
		for range t.C {
			allowAccessByDefault, authzProviders, _, _, _ := frontendAuthz.ProvidersFromConfig(
				ctx,
				conf.Get(),
				db.ExternalServices(),
				db,
			)
			ossAuthz.SetProviders(allowAccessByDefault, authzProviders)
		}
	}()

	go syncer.Run(ctx)
}
