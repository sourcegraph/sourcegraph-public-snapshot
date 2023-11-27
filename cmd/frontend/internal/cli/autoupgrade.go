package cli

import (
	"context"
	"database/sql"
	"os"
	"time"

	"github.com/jackc/pgerrcode"
	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	connections "github.com/sourcegraph/sourcegraph/internal/database/connections/live"
	"github.com/sourcegraph/sourcegraph/internal/database/migration"
	"github.com/sourcegraph/sourcegraph/internal/database/migration/multiversion"
	"github.com/sourcegraph/sourcegraph/internal/database/migration/runner"
	"github.com/sourcegraph/sourcegraph/internal/database/migration/schemas"
	"github.com/sourcegraph/sourcegraph/internal/database/migration/store"
	"github.com/sourcegraph/sourcegraph/internal/database/postgresdsn"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/oobmigration"
	"github.com/sourcegraph/sourcegraph/internal/oobmigration/migrations/register"
	"github.com/sourcegraph/sourcegraph/internal/service"
	"github.com/sourcegraph/sourcegraph/internal/version"
	"github.com/sourcegraph/sourcegraph/internal/version/upgradestore"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/output"
)

const appName = "frontend-autoupgrader"

var AutoUpgradeDone = make(chan struct{})

func tryAutoUpgrade(ctx context.Context, obsvCtx *observation.Context, ready service.ReadyFunc, hook store.RegisterMigratorsUsingConfAndStoreFactoryFunc) (err error) {
	defer func() {
		close(AutoUpgradeDone)
	}()

	sqlDB, err := connections.RawNewFrontendDB(obsvCtx, "", appName)
	if err != nil {
		return errors.Errorf("failed to connect to frontend database: %s", err)
	}
	defer sqlDB.Close()

	db := database.NewDB(obsvCtx.Logger, sqlDB)
	upgradestore := upgradestore.New(db)

	currentVersionStr, dbShouldAutoUpgrade, err := upgradestore.GetAutoUpgrade(ctx)
	// fresh instance
	if errors.Is(err, sql.ErrNoRows) || errors.HasPostgresCode(err, pgerrcode.UndefinedTable) {
		return nil
	} else if err != nil {
		return errors.Wrap(err, "autoupgradestore.GetAutoUpgrade")
	}
	if !dbShouldAutoUpgrade && !multiversion.EnvShouldAutoUpgrade {
		return nil
	}

	currentVersion, currentPatch, ok := oobmigration.NewVersionAndPatchFromString(currentVersionStr)
	if !ok {
		return errors.Newf("unexpected string for desired instance schema version, skipping auto-upgrade (%s)", currentVersionStr)
	}

	toVersionStr := version.Version()
	toVersion, toPatch, ok := oobmigration.NewVersionAndPatchFromString(toVersionStr)
	if !ok {
		obsvCtx.Logger.Warn("unexpected string for desired instance schema version, skipping auto-upgrade", log.String("version", toVersionStr))
		return nil
	}

	if oobmigration.CompareVersions(currentVersion, toVersion) == oobmigration.VersionOrderEqual && currentPatch >= toPatch {
		return nil
	}

	stopFunc, err := serveInternalServer(obsvCtx)
	if err != nil {
		return errors.Wrap(err, "failed to start configuration server")
	}
	defer stopFunc()

	stopFunc, err = serveExternalServer(obsvCtx, sqlDB, db)
	if err != nil {
		return errors.Wrap(err, "failed to start UI & healthcheck server")
	}
	defer stopFunc()

	ready()

	if err := upgradestore.EnsureUpgradeTable(ctx); err != nil {
		return errors.Wrap(err, "autoupgradestore.EnsureUpgradeTable")
	}

	stillNeedsMVU, err := claimAutoUpgradeLock(ctx, obsvCtx, db, toVersion)
	if err != nil {
		return err
	}
	if !stillNeedsMVU {
		// may not need an MVU (major/minor versions match), but still need to update for patch version difference
		if oobmigration.CompareVersions(currentVersion, toVersion) == oobmigration.VersionOrderEqual && currentPatch < toPatch {
			return finalMileMigrations(obsvCtx)
		}
		return nil
	}

	var success bool
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel()
		if err := upgradestore.SetUpgradeStatus(ctx, success); err != nil {
			obsvCtx.Logger.Error("failed to set auto-upgrade status", log.Error(err))
		}
	}()

	stopHeartbeat, err := heartbeatLoop(obsvCtx.Logger, db)
	if err != nil {
		return err
	}
	defer stopHeartbeat()

	plan, err := planMigration(currentVersion, toVersion)
	if err != nil {
		return errors.Wrap(err, "error planning auto-upgrade")
	}
	if err := upgradestore.SetUpgradePlan(ctx, multiversion.SerializeUpgradePlan(plan)); err != nil {
		return errors.Wrap(err, "error updating auto-upgrade plan")
	}
	if err := runMigration(ctx, obsvCtx, plan, db, hook); err != nil {
		return errors.Wrap(err, "error during auto-upgrade")
	}

	if err := upgradestore.SetAutoUpgrade(ctx, false); err != nil {
		return errors.Wrap(err, "autoupgradestore.SetAutoUpgrade")
	}

	if err := finalMileMigrations(obsvCtx); err != nil {
		return err
	}

	success = true
	obsvCtx.Logger.Info("Upgrade successful")
	return nil
}

func planMigration(from, to oobmigration.Version) (multiversion.MigrationPlan, error) {
	versionRange, err := oobmigration.UpgradeRange(from, to)
	if err != nil {
		return multiversion.MigrationPlan{}, err
	}

	interrupts, err := oobmigration.ScheduleMigrationInterrupts(from, to)
	if err != nil {
		return multiversion.MigrationPlan{}, err
	}

	plan, err := multiversion.PlanMigration(from, to, versionRange, interrupts)
	if err != nil {
		return multiversion.MigrationPlan{}, err
	}

	return plan, nil
}

func runMigration(
	ctx context.Context,
	obsvCtx *observation.Context,
	plan multiversion.MigrationPlan,
	db database.DB,
	enterpriseMigratorsHook store.RegisterMigratorsUsingConfAndStoreFactoryFunc,
) error {
	registerMigrators := store.ComposeRegisterMigratorsFuncs(
		register.RegisterOSSMigratorsUsingConfAndStoreFactory,
		enterpriseMigratorsHook,
	)

	// tee := io.MultiWriter(&buffer, os.Stdout)
	out := output.NewOutput(os.Stdout, output.OutputOpts{})

	runnerFactory := func(schemaNames []string, schemas []*schemas.Schema) (*runner.Runner, error) {
		return migration.NewRunnerWithSchemas(
			obsvCtx,
			out,
			appName, schemaNames, schemas,
		)
	}

	return multiversion.RunMigration(
		ctx,
		db,
		runnerFactory,
		plan,
		runner.ApplyPrivilegedMigrations,
		nil, // only needed when ^ is NoopPrivilegedMigrations
		true,
		multiversion.EnvAutoUpgradeSkipDrift,
		false,
		true,
		false,
		registerMigrators,
		schemas.DefaultSchemaFactories,
		out,
	)
}

type dialer func(_ *observation.Context, dsn string, appName string) (*sql.DB, error)

// performs the role of `migrator up`, applying any migrations in the patch versions between the minor version we're at (that `upgrade` brings you to)
// and the patch version we desire to be at.
func finalMileMigrations(obsvCtx *observation.Context) error {
	dsns, err := postgresdsn.DSNsBySchema(schemas.SchemaNames)
	if err != nil {
		return err
	}

	migratorsBySchema := map[string]dialer{
		"frontend":     connections.MigrateNewFrontendDB,
		"codeintel":    connections.MigrateNewCodeIntelDB,
		"codeinsights": connections.MigrateNewCodeInsightsDB,
	}
	for schema, migrateLastMile := range migratorsBySchema {
		obsvCtx.Logger.Info("Running last-mile migrations", log.String("schema", schema))

		sqlDB, err := migrateLastMile(obsvCtx, dsns[schema], appName)
		if err != nil {
			return errors.Wrapf(err, "failed to perform last-mile migration for %s schema", schema)
		}
		sqlDB.Close()
	}

	return nil
}

// claims a "lock" to prevent other frontends from attempting to autoupgrade concurrently, looping while the lock couldn't be claimed until either
// 1) the version is where we want to be at or
// 2) the lock was claimed by us
// and
// there are no named connections in pg_stat_activity besides frontend-autoupgrader.
func claimAutoUpgradeLock(ctx context.Context, obsvCtx *observation.Context, db database.DB, toVersion oobmigration.Version) (stillNeedsUpgrade bool, err error) {
	upgradestore := upgradestore.New(db)

	// try to claim
	for {
		obsvCtx.Logger.Info("attempting to claim autoupgrade lock")

		currentVersionStr, _, err := upgradestore.GetServiceVersion(ctx)
		if err != nil {
			return false, errors.Wrap(err, "autoupgradestore.GetServiceVersion")
		}

		currentVersion, ok := oobmigration.NewVersionFromString(currentVersionStr)
		if !ok {
			return false, errors.Newf("unexpected string for current instance schema version: %q", currentVersion)
		}

		if cmp := oobmigration.CompareVersions(currentVersion, toVersion); cmp == oobmigration.VersionOrderAfter || cmp == oobmigration.VersionOrderEqual {
			obsvCtx.Logger.Info("installation is up-to-date, nothing to do!")
			return false, nil
		}

		// we want to block until all named connections (which we make use of) besides 'frontend-autoupgrader' are no longer connected,
		// so that:
		// 1) we know old frontends are retired and not coming back (due to new frontends running health/ready server)
		// 2) dependent services have picked up the magic DSN and restarted
		// TODO: can we surface this in the UI?
		remainingConnections, err := checkForDisconnects(ctx, obsvCtx.Logger, db)
		if err != nil {
			return false, err
		}
		if len(remainingConnections) > 0 {
			obsvCtx.Logger.Warn("named postgres connections found, waiting for them to shutdown, manually shutdown any unexpected ones", log.Strings("applications", remainingConnections))

			time.Sleep(time.Second * 10)

			continue
		}

		claimed, err := upgradestore.ClaimAutoUpgrade(ctx, currentVersionStr, toVersion.String())
		if err != nil {
			return false, errors.Wrap(err, "autoupgradstore.ClaimAutoUpgrade")
		}

		if claimed {
			obsvCtx.Logger.Info("claimed autoupgrade lock")
			return true, nil
		}

		obsvCtx.Logger.Warn("unable to claim autoupgrade lock, sleeping...")

		time.Sleep(time.Second * 10)
	}
}

const heartbeatInterval = time.Second * 10

func heartbeatLoop(logger log.Logger, db database.DB) (func(), error) {
	upgradestore := upgradestore.New(db)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	if err := upgradestore.Heartbeat(ctx); err != nil {
		return nil, errors.Wrap(err, "error executing autoupgrade heartbeat")
	}

	ticker := time.NewTicker(heartbeatInterval)
	done := make(chan struct{})
	go func() {
		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				func() {
					ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
					defer cancel()
					if err := upgradestore.Heartbeat(ctx); err != nil {
						logger.Error("error executing autoupgrade heartbeat", log.Error(err))
					}
				}()
			}
		}
	}()

	return func() { ticker.Stop(); close(done) }, nil
}

func checkForDisconnects(ctx context.Context, _ log.Logger, db database.DB) (remaining []string, err error) {
	query := sqlf.Sprintf(`SELECT DISTINCT(application_name, datname) FROM pg_stat_activity
	WHERE application_name <> '' AND application_name <> %s AND application_name <> 'psql' AND datname <> '' AND datname <> %s`,
		appName)
	store := basestore.NewWithHandle(db.Handle())
	remaining, err = basestore.ScanStrings(store.Query(ctx, query))
	if err != nil {
		return nil, err
	}

	return remaining, nil
}
