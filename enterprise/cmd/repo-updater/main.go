package main

import (
	"context"
	"os"
	"strconv"
	"time"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/envvar"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/globals"
	"github.com/sourcegraph/sourcegraph/cmd/repo-updater/repoupdater"
	"github.com/sourcegraph/sourcegraph/cmd/repo-updater/shared"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/repo-updater/internal/authz"
	frontendAuthz "github.com/sourcegraph/sourcegraph/enterprise/internal/authz"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/batches"
	edb "github.com/sourcegraph/sourcegraph/enterprise/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	ossAuthz "github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	ossDB "github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/debugserver"
	"github.com/sourcegraph/sourcegraph/internal/encryption/keyring"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/internal/ratelimit"
	"github.com/sourcegraph/sourcegraph/internal/repos"
	"github.com/sourcegraph/sourcegraph/internal/timeutil"
)

func main() {
	shared.Main(enterpriseInit)
}

func enterpriseInit(
	logger log.Logger,
	db ossDB.DB,
	repoStore repos.Store,
	keyring keyring.Ring,
	cf *httpcli.Factory,
	server *repoupdater.Server,
) (debugDumpers map[string]debugserver.Dumper) {
	debug, _ := strconv.ParseBool(os.Getenv("DEBUG"))
	if debug {
		logger.Info("enterprise edition")
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

	permsStore := edb.Perms(logger, db, timeutil.Now)
	permsSyncer := authz.NewPermsSyncer(logger.Scoped("PermsSyncer", "repository and user permissions syncer"), db, repoStore, permsStore, timeutil.Now, ratelimit.DefaultRegistry)
	go startBackgroundPermsSync(ctx, permsSyncer, db)
	if server != nil {
		server.PermsSyncer = permsSyncer
	}

	return map[string]debugserver.Dumper{
		"repoPerms": permsSyncer,
	}
}

// startBackgroundPermsSync sets up background permissions syncing.
func startBackgroundPermsSync(ctx context.Context, syncer *authz.PermsSyncer, db ossDB.DB) {
	globals.WatchPermissionsUserMapping()
	go func() {
		t := time.NewTicker(frontendAuthz.RefreshInterval())
		for range t.C {
			allowAccessByDefault, authzProviders, _, _, _ :=
				frontendAuthz.ProvidersFromConfig(
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
