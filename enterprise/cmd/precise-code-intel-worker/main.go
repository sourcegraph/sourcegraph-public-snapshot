package main

import (
	"context"
	"database/sql"
	"net/http"
	"time"

	smithyhttp "github.com/aws/smithy-go/transport/http"
	"github.com/opentracing/opentracing-go"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/enterprise/cmd/precise-code-intel-worker/internal/worker"
	eiauthz "github.com/sourcegraph/sourcegraph/enterprise/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/stores"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/stores/dbstore"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/stores/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/stores/lsifstore"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/stores/lsifuploadstore"
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
	"github.com/sourcegraph/sourcegraph/internal/sentry"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/internal/tracer"
	"github.com/sourcegraph/sourcegraph/internal/uploadstore"
	"github.com/sourcegraph/sourcegraph/internal/version"
	"github.com/sourcegraph/sourcegraph/internal/workerutil"
	"github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

const addr = ":3188"

func main() {
	config := &Config{}
	config.Load()

	env.Lock()
	env.HandleHelpFlag()
	conf.Init()
	logging.Init()
	liblog := log.Init(log.Resource{
		Name:       env.MyName,
		Version:    version.Version(),
		InstanceID: hostname.Get(),
	})
	defer liblog.Sync()
	go conf.Watch(liblog.Update(conf.GetLogSinks))
	tracer.Init(conf.DefaultClient())
	sentry.Init(conf.DefaultClient())
	trace.Init()
	profiler.Init()

	logger := log.Scoped("worker", "The precise-code-intel-worker service converts LSIF upload file into Postgres data.")

	if err := config.Validate(); err != nil {
		logger.Error("Failed for load config", log.Error(err))
	}

	// Initialize tracing/metrics
	observationContext := &observation.Context{
		Logger:     log.Scoped("worker", "the precise codeintel worker"),
		Tracer:     &trace.Tracer{Tracer: opentracing.GlobalTracer()},
		Registerer: prometheus.DefaultRegisterer,
		HoneyDataset: &honey.Dataset{
			Name: "codeintel-worker",
		},
	}

	// Start debug server
	ready := make(chan struct{})
	go debugserver.NewServerRoutine(ready).Start()

	if err := keyring.Init(context.Background()); err != nil {
		logger.Fatal("Failed to intialise keyring", log.Error(err))
	}

	// Connect to databases
	db := database.NewDB(mustInitializeDB())
	codeIntelDB := mustInitializeCodeIntelDB()

	// Migrations may take a while, but after they're done we'll immediately
	// spin up a server and can accept traffic. Inform external clients we'll
	// be ready for traffic.
	close(ready)

	// Initialize stores
	dbStore := dbstore.NewWithDB(db, observationContext)
	workerStore := dbstore.WorkerutilUploadStore(dbStore, makeObservationContext(observationContext, false))
	lsifStore := lsifstore.NewStore(codeIntelDB, conf.Get(), observationContext)
	gitserverClient := gitserver.New(db, dbStore, observationContext)

	uploadStore, err := lsifuploadstore.New(context.Background(), config.LSIFUploadStoreConfig, observationContext)
	if err != nil {
		logger.Fatal("Failed to create upload store", log.Error(err))
	}
	if err := initializeUploadStore(context.Background(), uploadStore); err != nil {
		logger.Fatal("Failed to initialize upload store", log.Error(err))
	}

	// Initialize sub-repo permissions client
	authz.DefaultSubRepoPermsChecker, err = authz.NewSubRepoPermsClient(db.SubRepoPerms())
	if err != nil {
		logger.Fatal("Failed to create sub-repo client", log.Error(err))
	}

	// Initialize metrics
	dbworker.InitPrometheusMetric(observationContext, workerStore, "codeintel", "upload", nil)

	// Initialize worker
	worker := worker.NewWorker(
		&worker.DBStoreShim{Store: dbStore},
		workerStore,
		&worker.LSIFStoreShim{Store: lsifStore},
		uploadStore,
		gitserverClient,
		config.WorkerPollInterval,
		config.WorkerConcurrency,
		config.WorkerBudget,
		config.MaximumRuntimePerJob,
		makeWorkerMetrics(observationContext),
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

func mustInitializeDB() *sql.DB {
	dsn := conf.GetServiceConnectionValueAndRestartOnChange(func(serviceConnections conftypes.ServiceConnections) string {
		return serviceConnections.PostgresDSN
	})
	logger := log.Scoped("init db", "Initialize fontend database")
	sqlDB, err := connections.EnsureNewFrontendDB(dsn, "precise-code-intel-worker", &observation.TestContext)
	if err != nil {
		logger.Fatal("Failed to connect to frontend database", log.Error(err))
	}

	//
	// START FLAILING

	ctx := context.Background()
	db := database.NewDB(sqlDB)
	go func() {
		for range time.NewTicker(eiauthz.RefreshInterval()).C {
			allowAccessByDefault, authzProviders, _, _ := eiauthz.ProvidersFromConfig(ctx, conf.Get(), db.ExternalServices(), db)
			authz.SetProviders(allowAccessByDefault, authzProviders)
		}
	}()

	// END FLAILING
	//

	return sqlDB
}

func mustInitializeCodeIntelDB() stores.CodeIntelDB {
	dsn := conf.GetServiceConnectionValueAndRestartOnChange(func(serviceConnections conftypes.ServiceConnections) string {
		return serviceConnections.CodeIntelPostgresDSN
	})
	logger := log.Scoped("init db", "Initialize codeintel database.")
	db, err := connections.EnsureNewCodeIntelDB(dsn, "precise-code-intel-worker", &observation.TestContext)
	if err != nil {
		logger.Fatal("Failed to connect to codeintel database", log.Error(err))
	}

	return stores.NewCodeIntelDB(db)
}

func makeObservationContext(observationContext *observation.Context, withHoney bool) *observation.Context {
	if withHoney {
		return observationContext
	}
	ctx := *observationContext
	ctx.HoneyDataset = nil
	return &ctx
}

func makeWorkerMetrics(observationContext *observation.Context) workerutil.WorkerMetrics {
	return workerutil.NewMetrics(observationContext, "codeintel_upload_processor")
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
