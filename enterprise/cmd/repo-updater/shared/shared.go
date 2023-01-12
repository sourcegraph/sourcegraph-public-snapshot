package shared

import (
	"context"
	"os"
	"strconv"
	"time"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/envvar"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/globals"
	"github.com/sourcegraph/sourcegraph/cmd/repo-updater/repoupdater"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/repo-updater/internal/authz"
	frontendAuthz "github.com/sourcegraph/sourcegraph/enterprise/internal/authz"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/batches"
	edb "github.com/sourcegraph/sourcegraph/enterprise/internal/database"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/licensing"
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
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func EnterpriseInit(
	logger log.Logger,
	db ossDB.DB,
	repoStore repos.Store,
	keyring keyring.Ring,
	cf *httpcli.Factory,
	server *repoupdater.Server,
	syncer *repos.Syncer,
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

	enterpriseDB := edb.NewEnterpriseDB(db)
	_, err := enterpriseDB.FreeLicense().Init(ctx)
	if err != nil {
		logger.Fatal("failed to initialize free license", log.Error(err))
	}

	syncer.EnterpriseCreateRepoHook = enterpriseCreateRepoHook
	syncer.EnterpriseUpdateRepoHook = enterpriseUpdateRepoHook

	permsStore := enterpriseDB.Perms()
	permsSyncer := authz.NewPermsSyncer(logger.Scoped("PermsSyncer", "repository and user permissions syncer"), db, repoStore, permsStore, timeutil.Now, ratelimit.DefaultRegistry)
	go startBackgroundPermsSync(ctx, permsSyncer, db)
	if server != nil {
		server.PermsSyncer = permsSyncer
	}

	return map[string]debugserver.Dumper{
		"repoPerms": permsSyncer,
	}
}

func enterpriseCreateRepoHook(ctx context.Context, s repos.Store, repo *types.Repo) error {
	if !repo.Private {
		return nil
	}

	if prFeature := (*licensing.FeaturePrivateRepositories)(nil); licensing.Check(prFeature) == nil {
		if prFeature.Unrestricted {
			return nil
		}

		numPrivateRepos, err := s.RepoStore().Count(ctx, ossDB.ReposListOptions{OnlyPrivate: true})
		if err != nil {
			return err
		}

		if numPrivateRepos >= prFeature.MaxNumPrivateRepos {
			return errors.Newf("maximum number of private repositories (%d) reached", prFeature.MaxNumPrivateRepos)
		}
	}

	return licensing.NewFeatureNotActivatedError("The private repositories feature is not activated for this license. Please upgrade your license to use this feature.")
}

func enterpriseUpdateRepoHook(ctx context.Context, s repos.Store, existingRepo *types.Repo, newRepo *types.Repo) error {
	// If the privacy of the repo remains the same, or changes to public,
	// we don't need to do any checks
	if existingRepo.Private == newRepo.Private || !newRepo.Private {
		return nil
	}

	if prFeature := (*licensing.FeaturePrivateRepositories)(nil); licensing.Check(prFeature) == nil {
		if prFeature.Unrestricted {
			return nil
		}

		numPrivateRepos, err := s.RepoStore().Count(ctx, ossDB.ReposListOptions{OnlyPrivate: true})
		if err != nil {
			return err
		}

		if numPrivateRepos >= prFeature.MaxNumPrivateRepos {
			return errors.Newf("maximum number of private repositories (%d) reached", prFeature.MaxNumPrivateRepos)
		}
	}

	return licensing.NewFeatureNotActivatedError("The private repositories feature is not activated for this license. Please upgrade your license to use this feature.")
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
