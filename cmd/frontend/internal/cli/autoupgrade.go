package cli

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	gcontext "github.com/gorilla/context"
	"github.com/gorilla/mux"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/app/assetsutil"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/httpapi"
	apirouter "github.com/sourcegraph/sourcegraph/cmd/frontend/internal/httpapi/router"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
	"github.com/sourcegraph/sourcegraph/internal/database"
	connections "github.com/sourcegraph/sourcegraph/internal/database/connections/live"
	"github.com/sourcegraph/sourcegraph/internal/database/migration"
	"github.com/sourcegraph/sourcegraph/internal/database/migration/multiversion"
	"github.com/sourcegraph/sourcegraph/internal/database/migration/runner"
	"github.com/sourcegraph/sourcegraph/internal/database/migration/schemas"
	"github.com/sourcegraph/sourcegraph/internal/database/migration/store"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/httpserver"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/oobmigration"
	"github.com/sourcegraph/sourcegraph/internal/oobmigration/migrations"
	"github.com/sourcegraph/sourcegraph/internal/version"
	"github.com/sourcegraph/sourcegraph/internal/version/upgradestore"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/output"
	"github.com/sourcegraph/sourcegraph/schema"
)

var buffer strings.Builder // :)

var shouldAutoUpgade = env.MustGetBool("SRC_AUTOUPGRADE", false, "If you forgot to set intent to autoupgrade before shutting down the instance, set this env var.")

func tryAutoUpgrade(ctx context.Context, obsvCtx *observation.Context, db database.DB, hook store.RegisterMigratorsUsingConfAndStoreFactoryFunc) (err error) {
	upgradestore := upgradestore.New(db)

	_, doAutoUpgrade, err := upgradestore.GetAutoUpgrade(ctx)
	if err != nil {
		return errors.Wrap(err, "autoupgradestore.GetAutoUpgrade")
	}
	if !doAutoUpgrade && !shouldAutoUpgade {
		return nil
	}

	stopFunc, err := serveConfigurationServer(obsvCtx)
	if err != nil {
		return err
	}
	defer stopFunc()

	stopFunc, err = serveUpgradeUI()
	if err != nil {
		return err
	}
	defer stopFunc()

	toVersion, ok := oobmigration.NewVersionFromString(version.Version())
	if !ok {
		return nil
	}

	return performAutoUpgrade(ctx, obsvCtx, db, toVersion, hook)
}

func performAutoUpgrade(ctx context.Context, obsvCtx *observation.Context, db database.DB, toVersion oobmigration.Version, hook store.RegisterMigratorsUsingConfAndStoreFactoryFunc) error {
	upgradestore := upgradestore.New(db)

	var (
		currentVersion oobmigration.Version
		success        bool
	)

	if err := upgradestore.EnsureUpgradeTable(ctx); err != nil {
		return errors.Wrap(err, "autoupgradestore.EnsureUpgradeTable")
	}

	// try to claim
	for {
		obsvCtx.Logger.Info("attempting to claim autoupgrade lock")

		currentVersionStr, _, err := upgradestore.GetServiceVersion(ctx)
		if err != nil {
			return errors.Wrap(err, "autoupgradestore.GetServiceVersion")
		}

		var ok bool
		currentVersion, ok = oobmigration.NewVersionFromString(currentVersionStr)
		if !ok {
			return errors.Newf("VERSION STRING BAD %s", currentVersion)
		}

		if cmp := oobmigration.CompareVersions(currentVersion, toVersion); cmp == oobmigration.VersionOrderAfter || cmp == oobmigration.VersionOrderEqual {
			obsvCtx.Logger.Info("installation is up-to-date, nothing to do!")
			return nil
		}

		claimed, err := upgradestore.ClaimAutoUpgrade(ctx, currentVersionStr, version.Version())
		if err != nil {
			return errors.Wrap(err, "autoupgradstore.ClaimAutoUpgrade")
		}

		if claimed {
			defer func() {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
				defer cancel()
				if err := upgradestore.SetUpgradeStatus(ctx, success); err != nil {
					fmt.Println("whoopsy", err)
				}
			}()
			break
		}

		obsvCtx.Logger.Warn("unable to claim autoupgrade lock, sleeping...")

		time.Sleep(time.Second * 10)
	}

	if err := runMigration(ctx, obsvCtx, currentVersion, toVersion, db, hook); err != nil {
		return err
	}

	if err := upgradestore.SetAutoUpgrade(ctx, false); err != nil {
		return errors.Wrap(err, "autoupgradestore.SetAutoUpgrade")
	}

	schemas := []string{"frontend", "codeintel", "codeinsights"}
	for i, fn := range []func(_ *observation.Context, dsn string, appName string) (*sql.DB, error){
		connections.MigrateNewFrontendDB, connections.MigrateNewCodeIntelDB, connections.MigrateNewCodeInsightsDB,
	} {
		sqlDB, err := fn(obsvCtx, "", "frontend")
		if err != nil {
			return errors.Wrapf(err, "failed to perform last-mile migration for %s schema", schemas[i])
		}
		sqlDB.Close()
	}

	success = true

	obsvCtx.Logger.Info("MIGRATION SUCCESSFUL")

	return nil // errors.New("MIGRATION SUCCEEDED, RESTARTING")
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

	tee := io.MultiWriter(&buffer, os.Stdout)
	out := output.NewOutput(tee, output.OutputOpts{})

	runnerFactory := func(schemaNames []string, schemas []*schemas.Schema) (*runner.Runner, error) {
		return migration.NewRunnerWithSchemas(
			obsvCtx,
			out,
			"frontend-autoupgrader", schemaNames, schemas,
		)
	}

	return multiversion.RunMigration(
		ctx,
		db,
		runnerFactory,
		plan,
		runner.ApplyPrivilegedMigrations,
		nil,
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

func serveConfigurationServer(obsvCtx *observation.Context) (context.CancelFunc, error) {
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
				PostgresDSN:          "lol",
				CodeIntelPostgresDSN: "lol",
				CodeInsightsDSN:      "lol",
			},
		}
		return json.NewEncoder(w).Encode(configuration)
	}))

	serveMux.Handle("/.internal/", internalRouter)
	serveMux.Handle("/.assets/", http.StripPrefix("/.assets", secureHeadersMiddleware(assetsutil.NewAssetHandler(serveMux), crossOriginPolicyAssets)))

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

func serveUpgradeUI() (context.CancelFunc, error) {
	serveMux := http.NewServeMux()

	serveMux.Handle("/.assets/", http.StripPrefix("/.assets", secureHeadersMiddleware(assetsutil.NewAssetHandler(serveMux), crossOriginPolicyAssets)))
	serveMux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `<body>
			<h1>MIGRATION IN PROGRESS</h1>
		</body>`)
	})
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
