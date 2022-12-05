package shared

import (
	"context"
	"database/sql"
	"net/http"
	"time"

	smithyhttp "github.com/aws/smithy-go/transport/http"
	"github.com/sourcegraph/log"

	eiauthz "github.com/sourcegraph/sourcegraph/enterprise/internal/authz"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel"
	codeintelshared "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/shared"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/shared/lsifuploadstore"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/uploads"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
	"github.com/sourcegraph/sourcegraph/internal/database"
	connections "github.com/sourcegraph/sourcegraph/internal/database/connections/live"
	"github.com/sourcegraph/sourcegraph/internal/debugserver"
	"github.com/sourcegraph/sourcegraph/internal/encryption/keyring"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/honey"
	"github.com/sourcegraph/sourcegraph/internal/hostname"
	"github.com/sourcegraph/sourcegraph/internal/httpserver"
	"github.com/sourcegraph/sourcegraph/internal/logging"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/profiler"
	"github.com/sourcegraph/sourcegraph/internal/tracer"
	"github.com/sourcegraph/sourcegraph/internal/uploadstore"
	"github.com/sourcegraph/sourcegraph/internal/version"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

const addr = ":3188"

func Main() {
	config := &Config{}
	config.Load()

	env.Lock()
	env.HandleHelpFlag()
	logging.Init()
	liblog := log.Init(log.Resource{
		Name:       env.MyName,
		Version:    version.Version(),
		InstanceID: hostname.Get(),
	})
	defer liblog.Sync()

	// Initialize tracing/metrics
	logger := log.Scoped("codeintel-worker", "The precise-code-intel-worker service converts LSIF upload file into Postgres data.")
	observationCtx := observation.NewContext(logger, observation.Honeycomb(&honey.Dataset{
		Name: "codeintel-worker",
	}))

	conf.Init()
	go conf.Watch(liblog.Update(conf.GetLogSinks))
	tracer.Init(logger.Scoped("tracer", "internal tracer package"), conf.DefaultClient())
	profiler.Init()

	if err := config.Validate(); err != nil {
		logger.Error("Failed for load config", log.Error(err))
	}

	// Start debug server
	ready := make(chan struct{})
	go debugserver.NewServerRoutine(ready).Start()

	if err := keyring.Init(context.Background()); err != nil {
		logger.Fatal("Failed to intialise keyring", log.Error(err))
	}

	// Connect to databases
	db := database.NewDB(logger, mustInitializeDB(observationCtx))
	codeIntelDB := mustInitializeCodeIntelDB(observationCtx)

	// Migrations may take a while, but after they're done we'll immediately
	// spin up a server and can accept traffic. Inform external clients we'll
	// be ready for traffic.
	close(ready)

	// Initialize sub-repo permissions client
	var err error
	authz.DefaultSubRepoPermsChecker, err = authz.NewSubRepoPermsClient(db.SubRepoPerms())
	if err != nil {
		logger.Fatal("Failed to create sub-repo client", log.Error(err))
	}

	services, err := codeintel.NewServices(codeintel.ServiceDependencies{
		DB:             db,
		CodeIntelDB:    codeIntelDB,
		ObservationCtx: observationCtx,
	})
	if err != nil {
		logger.Fatal("Failed to create codeintel services", log.Error(err))
	}

	// Initialize stores
	uploadStore, err := lsifuploadstore.New(context.Background(), observationCtx, config.LSIFUploadStoreConfig)
	if err != nil {
		logger.Fatal("Failed to create upload store", log.Error(err))
	}
	if err := initializeUploadStore(context.Background(), uploadStore); err != nil {
		logger.Fatal("Failed to initialize upload store", log.Error(err))
	}

	// Initialize worker
	worker := uploads.NewUploadProcessorJob(
		observationCtx,
		services.UploadsService,
		db,
		uploadStore,
		config.WorkerConcurrency,
		config.WorkerBudget,
		config.WorkerPollInterval,
		config.MaximumRuntimePerJob,
	)

	// Initialize health server
	server := httpserver.NewFromAddr(addr, &http.Server{
		ReadTimeout:  75 * time.Second,
		WriteTimeout: 10 * time.Minute,
		Handler:      httpserver.NewHandler(nil),
	})

	// Go!
	goroutine.MonitorBackgroundRoutines(context.Background(), worker, server)
}

func mustInitializeDB(observationCtx *observation.Context) *sql.DB {
	dsn := conf.GetServiceConnectionValueAndRestartOnChange(func(serviceConnections conftypes.ServiceConnections) string {
		return serviceConnections.PostgresDSN
	})
	sqlDB, err := connections.EnsureNewFrontendDB(observationCtx, dsn, "precise-code-intel-worker")
	if err != nil {
		log.Scoped("init db", "Initialize fontend database").Fatal("Failed to connect to frontend database", log.Error(err))
	}

	//
	// START FLAILING

	ctx := context.Background()
	db := database.NewDB(observationCtx.Logger, sqlDB)
	go func() {
		for range time.NewTicker(eiauthz.RefreshInterval()).C {
			allowAccessByDefault, authzProviders, _, _, _ := eiauthz.ProvidersFromConfig(ctx, conf.Get(), db.ExternalServices(), db)
			authz.SetProviders(allowAccessByDefault, authzProviders)
		}
	}()

	// END FLAILING
	//

	return sqlDB
}

func mustInitializeCodeIntelDB(observationCtx *observation.Context) codeintelshared.CodeIntelDB {
	dsn := conf.GetServiceConnectionValueAndRestartOnChange(func(serviceConnections conftypes.ServiceConnections) string {
		return serviceConnections.CodeIntelPostgresDSN
	})
	db, err := connections.EnsureNewCodeIntelDB(observationCtx, dsn, "precise-code-intel-worker")
	if err != nil {
		log.Scoped("init db", "Initialize codeintel database.").Fatal("Failed to connect to codeintel database", log.Error(err))
	}

	return codeintelshared.NewCodeIntelDB(observationCtx.Logger, db)
}

func initializeUploadStore(ctx context.Context, uploadStore uploadstore.Store) error {
	for {
		if err := uploadStore.Init(ctx); err == nil || !isRequestError(err) {
			return err
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(250 * time.Millisecond):
		}
	}
}

func isRequestError(err error) bool {
	return errors.HasType(err, &smithyhttp.RequestSendError{})
}
