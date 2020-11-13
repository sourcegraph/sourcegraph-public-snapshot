package main

import (
	"context"
	"database/sql"
	"log"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/inconshreveable/log15"
	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/globals"
	"github.com/sourcegraph/sourcegraph/cmd/repo-updater/repos"
	"github.com/sourcegraph/sourcegraph/cmd/repo-updater/repoupdater"
	"github.com/sourcegraph/sourcegraph/cmd/repo-updater/shared"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/repo-updater/authz"
	frontendAuthz "github.com/sourcegraph/sourcegraph/enterprise/internal/authz"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/campaigns"
	campaignsBackground "github.com/sourcegraph/sourcegraph/enterprise/internal/campaigns/background"
	edb "github.com/sourcegraph/sourcegraph/enterprise/internal/db"
	ossAuthz "github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	ossDB "github.com/sourcegraph/sourcegraph/internal/db"
	"github.com/sourcegraph/sourcegraph/internal/db/dbconn"
	"github.com/sourcegraph/sourcegraph/internal/db/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/debugserver"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/internal/ratelimit"
)

func main() {
	debug, _ := strconv.ParseBool(os.Getenv("DEBUG"))
	if debug {
		log.Println("enterprise edition")
	}
	shared.Main(enterpriseInit)
}

func enterpriseInit(
	db *sql.DB,
	repoStore repos.Store,
	cf *httpcli.Factory,
	server *repoupdater.Server,
) (debugDumpers []debugserver.Dumper) {
	ctx := context.Background()
	clock := func() time.Time {
		return time.Now().UTC().Truncate(time.Microsecond)
	}

	campaignsStore := campaigns.NewStoreWithClock(db, clock)

	syncRegistry := campaigns.NewSyncRegistry(ctx, campaignsStore, repoStore, cf)
	if server != nil {
		server.ChangesetSyncRegistry = syncRegistry
	}

	campaignsBackground.StartBackgroundJobs(ctx, db, campaignsStore, repoStore, cf)

	// TODO(jchen): This is an unfortunate compromise to not rewrite ossDB.ExternalServices for now.
	dbconn.Global = db
	permsStore := edb.NewPermsStore(db, clock)
	permsSyncer := authz.NewPermsSyncer(repoStore, permsStore, clock, ratelimit.DefaultRegistry)
	go startBackgroundPermsSync(ctx, permsSyncer, db)
	debugDumpers = append(debugDumpers, permsSyncer)
	if server != nil {
		server.PermsSyncer = permsSyncer
	}

	return debugDumpers
}

// startBackgroundPermsSync sets up background permissions syncing.
func startBackgroundPermsSync(ctx context.Context, syncer *authz.PermsSyncer, db dbutil.DB) {
	globals.WatchPermissionsUserMapping()
	go func() {
		// TODO(jchen): Delete this migration in 3.23
		// We only need to do this once at start because the write paths have taken
		// care of updating this value.
		var migrateExternalServiceUnrestricted sync.Once

		t := time.NewTicker(5 * time.Second)
		for range t.C {
			allowAccessByDefault, authzProviders, _, _ :=
				frontendAuthz.ProvidersFromConfig(ctx, conf.Get(), ossDB.ExternalServices)
			ossAuthz.SetProviders(allowAccessByDefault, authzProviders)

			migrateExternalServiceUnrestricted.Do(func() {
				// Collect IDs of external services which enforce repository permissions
				// and set others' `external_services.unrestricted` to `true`.
				esIDs := make([]*sqlf.Query, len(authzProviders))
				for i, p := range authzProviders {
					_, id := extsvc.DecodeURN(p.URN())
					esIDs[i] = sqlf.Sprintf("%s", id)
				}

				q := sqlf.Sprintf(`
UPDATE external_services
SET unrestricted = TRUE
WHERE
	id NOT IN (%s)
AND NOT unrestricted
`, sqlf.Join(esIDs, ","))
				_, err := db.ExecContext(ctx, q.Query(sqlf.PostgresBindVar), q.Args()...)
				if err != nil {
					log15.Error("Failed to update 'external_services.unrestricted'", "error", err)
					return
				}
			})
		}
	}()

	go syncer.Run(ctx)
}
