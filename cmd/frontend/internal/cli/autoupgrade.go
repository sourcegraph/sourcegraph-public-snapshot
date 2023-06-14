package cli

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"os"
	"time"

	gcontext "github.com/gorilla/context"
	"github.com/gorilla/mux"
	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/app/assetsutil"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/httpapi"
	apirouter "github.com/sourcegraph/sourcegraph/cmd/frontend/internal/httpapi/router"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	connections "github.com/sourcegraph/sourcegraph/internal/database/connections/live"
	"github.com/sourcegraph/sourcegraph/internal/database/dbconn"
	"github.com/sourcegraph/sourcegraph/internal/database/migration"
	"github.com/sourcegraph/sourcegraph/internal/database/migration/cliutil"
	"github.com/sourcegraph/sourcegraph/internal/database/migration/multiversion"
	"github.com/sourcegraph/sourcegraph/internal/database/migration/runner"
	"github.com/sourcegraph/sourcegraph/internal/database/migration/schemas"
	"github.com/sourcegraph/sourcegraph/internal/database/migration/store"
	"github.com/sourcegraph/sourcegraph/internal/database/postgresdsn"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/httpserver"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/oobmigration"
	"github.com/sourcegraph/sourcegraph/internal/oobmigration/migrations"
	"github.com/sourcegraph/sourcegraph/internal/service"
	"github.com/sourcegraph/sourcegraph/internal/version"
	"github.com/sourcegraph/sourcegraph/internal/version/upgradestore"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/output"
	"github.com/sourcegraph/sourcegraph/schema"
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
	if errors.Is(err, sql.ErrNoRows) {
		return nil
	} else if err != nil {
		return errors.Wrap(err, "autoupgradestore.GetAutoUpgrade")
	}
	if !dbShouldAutoUpgrade && !cliutil.EnvShouldAutoUpgrade {
		return nil
	}

	currentVersion, ok := oobmigration.NewVersionFromString(currentVersionStr)
	if !ok {
		return errors.Newf("unexpected string for desired instance schema version, skipping auto-upgrade (%s)", currentVersionStr)
	}

	stopFunc, err := serveInternalServer(obsvCtx)
	if err != nil {
		return errors.Wrap(err, "failed to start configuration server")
	}
	defer stopFunc()

	stopFunc, err = serveExternalServer(db)
	if err != nil {
		return errors.Wrap(err, "failed to start UI & healthcheck server")
	}
	defer stopFunc()

	ready()

	if err := blockForDisconnects(ctx, obsvCtx.Logger, db); err != nil {
		return errors.Wrap(err, "error blocking on postgres client disconnects")
	}

	toVersion, ok := oobmigration.NewVersionFromString(version.Version())
	if !ok {
		obsvCtx.Logger.Warn("unexpected string for desired instance schema version, skipping auto-upgrade", log.String("version", version.Version()))
		return nil
	}

	if err := upgradestore.EnsureUpgradeTable(ctx); err != nil {
		return errors.Wrap(err, "autoupgradestore.EnsureUpgradeTable")
	}

	stillNeedsUpgrade, err := claimAutoUpgradeLock(ctx, obsvCtx, upgradestore, toVersion)
	if err != nil {
		return err
	}
	if !stillNeedsUpgrade {
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

	if err := runMigration(ctx, obsvCtx, currentVersion, toVersion, db, hook); err != nil {
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

func runMigration(ctx context.Context,
	obsvCtx *observation.Context,
	from,
	to oobmigration.Version,
	db database.DB,
	enterpriseMigratorsHook store.RegisterMigratorsUsingConfAndStoreFactoryFunc,
) error {
	versionRange, err := oobmigration.UpgradeRange(from, to)
	if err != nil {
		return err
	}

	interrupts, err := oobmigration.ScheduleMigrationInterrupts(from, to)
	if err != nil {
		return err
	}

	plan, err := multiversion.PlanMigration(from, to, versionRange, interrupts)
	if err != nil {
		return err
	}

	registerMigrators := store.ComposeRegisterMigratorsFuncs(
		migrations.RegisterOSSMigratorsUsingConfAndStoreFactory,
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
		true,
		false,
		true,
		false,
		registerMigrators,
		nil, // only needed for drift
		out,
	)
}

// performs the role of `migrator up`, applying any migrations in the patch versions between the minor version we're at (that `upgrade` brings you to)
// and the patch version we desire to be at.
func finalMileMigrations(obsvCtx *observation.Context) error {
	dsns, err := postgresdsn.DSNsBySchema(schemas.SchemaNames)
	if err != nil {
		return err
	}
	schemas := []string{"frontend", "codeintel", "codeinsights"}
	for i, fn := range []func(_ *observation.Context, dsn string, appName string) (*sql.DB, error){
		connections.MigrateNewFrontendDB, connections.MigrateNewCodeIntelDB, connections.MigrateNewCodeInsightsDB,
	} {
		sqlDB, err := fn(obsvCtx, dsns[schemas[i]], "frontend")
		if err != nil {
			return errors.Wrapf(err, "failed to perform last-mile migration for %s schema", schemas[i])
		}
		sqlDB.Close()
	}

	return nil
}

type UpgradeStore interface {
	GetServiceVersion(ctx context.Context) (string, bool, error)
	ClaimAutoUpgrade(ctx context.Context, from, to string) (claimed bool, err error)
}

// claims a "lock" to prevent other frontends from attempting to autoupgrade concurrently, looping while the lock couldn't be claimed until either
// 1) the version is where we want to be at or
// 2) the lock was claimed by us
func claimAutoUpgradeLock(ctx context.Context, obsvCtx *observation.Context, upgradestore UpgradeStore, toVersion oobmigration.Version) (stillNeedsUpgrade bool, err error) {
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

		claimed, err := upgradestore.ClaimAutoUpgrade(ctx, currentVersionStr, version.Version())
		if err != nil {
			return false, errors.Wrap(err, "autoupgradstore.ClaimAutoUpgrade")
		}

		if claimed {
			return true, nil
		}

		obsvCtx.Logger.Warn("unable to claim autoupgrade lock, sleeping...")

		time.Sleep(time.Second * 10)
	}
}

func serveInternalServer(obsvCtx *observation.Context) (context.CancelFunc, error) {
	middleware := httpapi.JsonMiddleware(&httpapi.ErrorHandler{
		Logger:       obsvCtx.Logger,
		WriteErrBody: true,
	})

	serveMux := http.NewServeMux()

	internalRouter := mux.NewRouter().PathPrefix("/.internal").Subrouter()
	internalRouter.StrictSlash(true)
	internalRouter.Path("/configuration").Methods("POST").Name(apirouter.Configuration)
	internalRouter.Get(apirouter.Configuration).Handler(middleware(func(w http.ResponseWriter, r *http.Request) error {
		configuration := conf.Unified{
			SiteConfiguration: schema.SiteConfiguration{},
			ServiceConnectionConfig: conftypes.ServiceConnections{
				PostgresDSN:          dbconn.MigrationInProgressSentinelDSN,
				CodeIntelPostgresDSN: dbconn.MigrationInProgressSentinelDSN,
				CodeInsightsDSN:      dbconn.MigrationInProgressSentinelDSN,
			},
		}
		b, _ := json.Marshal(configuration.SiteConfiguration)
		raw := conftypes.RawUnified{
			Site:               string(b),
			ServiceConnections: configuration.ServiceConnections(),
		}
		return json.NewEncoder(w).Encode(raw)
	}))

	serveMux.Handle("/.internal/", internalRouter)

	h := gcontext.ClearHandler(serveMux)
	h = healthCheckMiddleware(h)

	server := &http.Server{
		Handler:      h,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
	}
	listener, err := httpserver.NewListener(httpAddrInternal)
	if err != nil {
		return nil, err
	}
	confServer := httpserver.New(listener, server)

	goroutine.Go(func() {
		confServer.Start()
	})

	return confServer.Stop, nil
}

func serveExternalServer(db database.DB) (context.CancelFunc, error) {
	serveMux := http.NewServeMux()

	serveMux.Handle("/.assets/", http.StripPrefix("/.assets", secureHeadersMiddleware(assetsutil.NewAssetHandler(serveMux), crossOriginPolicyAssets)))
	serveMux.HandleFunc("/", makeUpgradeProgressHandler(db))
	h := gcontext.ClearHandler(serveMux)
	h = healthCheckMiddleware(h)

	server := &http.Server{
		Handler:      h,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
	}
	listener, err := httpserver.NewListener(httpAddr)
	if err != nil {
		return nil, err
	}
	confServer := httpserver.New(listener, server)

	goroutine.Go(func() {
		confServer.Start()
	})

	return confServer.Stop, nil
}

// we want to block until all named connections (which we make use of) besides 'frontend-autoupgrader' are no longer connected,
// so that:
// 1) we know old frontends are retired and not coming back (due to new frontends running health/ready server)
// 2) dependent services have picked up the magic DSN and restarted
// TODO: can we surface this in the UI?
func blockForDisconnects(ctx context.Context, logger log.Logger, db database.DB) error {
	for {
		query := sqlf.Sprintf(`SELECT DISTINCT(application_name) FROM pg_stat_activity
			WHERE application_name <> '' AND application_name <> %s AND application_name <> 'psql'`,
			appName)
		store := basestore.NewWithHandle(db.Handle())
		applications, err := basestore.ScanStrings(store.Query(ctx, query))
		if err != nil {
			return err
		}

		if len(applications) == 0 {
			return nil
		}

		logger.Warn("named postgres connections found, waiting for them to shutdown, manually shutdown any unexpected ones", log.Strings("applications", applications))

		time.Sleep(time.Second * 10)
	}
}
