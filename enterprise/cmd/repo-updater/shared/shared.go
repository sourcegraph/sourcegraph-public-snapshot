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
	"github.com/sourcegraph/sourcegraph/internal/actor"
	ossAuthz "github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/authz/providers"
	"github.com/sourcegraph/sourcegraph/internal/batches"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/encryption/keyring"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/github/auth"
	ghaauth "github.com/sourcegraph/sourcegraph/internal/github_apps/auth"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/repos"
	"github.com/sourcegraph/sourcegraph/internal/timeutil"
)

func EnterpriseInit(
	observationCtx *observation.Context,
	db database.DB,
	repoStore repos.Store,
	keyring keyring.Ring,
	cf *httpcli.Factory,
	server *repoupdater.Server,
) {
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

	ghAppsStore := db.GitHubApps().WithEncryptionKey(keyring.GitHubAppKey)
	auth.FromConnection = ghaauth.CreateEnterpriseFromConnection(ghAppsStore, keyring.GitHubAppKey)

	permsStore := database.Perms(observationCtx.Logger, db, timeutil.Now)
	permsSyncer := authz.NewPermsSyncer(observationCtx.Logger.Scoped("PermsSyncer", "repository and user permissions syncer"), db, repoStore, permsStore, timeutil.Now)

	if server != nil {
		if server.Syncer != nil {
			server.Syncer.EnterpriseCreateRepoHook = enterpriseCreateRepoHook
			server.Syncer.EnterpriseUpdateRepoHook = enterpriseUpdateRepoHook
		}
	}

	repoWorkerStore := authz.MakeStore(observationCtx, db.Handle(), authz.SyncTypeRepo)
	userWorkerStore := authz.MakeStore(observationCtx, db.Handle(), authz.SyncTypeUser)
	permissionSyncJobStore := database.PermissionSyncJobsWith(observationCtx.Logger, db)
	repoSyncWorker := authz.MakeWorker(ctx, observationCtx, repoWorkerStore, permsSyncer, authz.SyncTypeRepo, permissionSyncJobStore)
	userSyncWorker := authz.MakeWorker(ctx, observationCtx, userWorkerStore, permsSyncer, authz.SyncTypeUser, permissionSyncJobStore)
	// Type of store (repo/user) for resetter doesn't matter, because it has its
	// separate name for logging and metrics.
	resetter := authz.MakeResetter(observationCtx, repoWorkerStore)

	go goroutine.MonitorBackgroundRoutines(ctx, repoSyncWorker, userSyncWorker, resetter)
	go watchAuthzProviders(ctx, db)
}

// watchAuthzProviders updates authz providers if config changes.
func watchAuthzProviders(ctx context.Context, db database.DB) {
	globals.WatchPermissionsUserMapping()
	go func() {
		t := time.NewTicker(providers.RefreshInterval())
		for range t.C {
			allowAccessByDefault, authzProviders, _, _, _ := providers.ProvidersFromConfig(
				ctx,
				conf.Get(),
				db.ExternalServices(),
				db,
			)
			ossAuthz.SetProviders(allowAccessByDefault, authzProviders)
		}
	}()
}
