package shared

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/prometheus/client_golang/prometheus"
	"go.opentelemetry.io/otel"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/symbols/fetcher"
	"github.com/sourcegraph/sourcegraph/cmd/symbols/gitserver"
	"github.com/sourcegraph/sourcegraph/cmd/symbols/internal/api"
	sqlite "github.com/sourcegraph/sourcegraph/cmd/symbols/internal/database"
	"github.com/sourcegraph/sourcegraph/cmd/symbols/types"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
	"github.com/sourcegraph/sourcegraph/internal/database"
	connections "github.com/sourcegraph/sourcegraph/internal/database/connections/live"
	"github.com/sourcegraph/sourcegraph/internal/debugserver"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/honey"
	"github.com/sourcegraph/sourcegraph/internal/hostname"
	"github.com/sourcegraph/sourcegraph/internal/httpserver"
	"github.com/sourcegraph/sourcegraph/internal/logging"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/profiler"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/internal/trace/ot"
	"github.com/sourcegraph/sourcegraph/internal/tracer"
	"github.com/sourcegraph/sourcegraph/internal/version"
)

const addr = ":3184"

type SetupFunc func(observationContext *observation.Context, db database.DB, gitserverClient gitserver.GitserverClient, repositoryFetcher fetcher.RepositoryFetcher) (types.SearchFunc, func(http.ResponseWriter, *http.Request), []goroutine.BackgroundRoutine, string, error)

func Main(setup SetupFunc) {
	// Initialization
	env.HandleHelpFlag()
	logging.Init()
	liblog := log.Init(log.Resource{
		Name:       env.MyName,
		Version:    version.Version(),
		InstanceID: hostname.Get(),
	}, log.NewSentrySinkWith(
		log.SentrySink{
			ClientOptions: sentry.ClientOptions{SampleRate: 0.2},
		},
	)) // Experimental: DevX is observing how sampling affects the errors signal
	defer liblog.Sync()

	conf.Init()
	go conf.Watch(liblog.Update(conf.GetLogSinks))
	tracer.Init(log.Scoped("tracer", "internal tracer package"), conf.DefaultClient())
	trace.Init()
	profiler.Init()

	routines := []goroutine.BackgroundRoutine{}

	// Initialize tracing/metrics
	logger := log.Scoped("service", "the symbols service")
	observationContext := &observation.Context{
		Logger:     logger,
		Tracer:     &trace.Tracer{TracerProvider: otel.GetTracerProvider()},
		Registerer: prometheus.DefaultRegisterer,
		HoneyDataset: &honey.Dataset{
			Name:       "codeintel-symbols",
			SampleRate: 5,
		},
	}

	// Allow to do a sanity check of sqlite.
	sanityCheck, err := strconv.ParseBool(env.Get("SANITY_CHECK", "false", "check that go-sqlite3 works then exit 0 if it's ok or 1 if not"))
	if err != nil {
		fmt.Printf("Invalid SANITY_CHECK value: %s\n", err.Error())
		os.Exit(1)
	}
	if sanityCheck {
		// Ensure we register our database driver before calling
		// anything that tries to open a SQLite database.
		sqlite.Init()

		fmt.Print("Running sanity check...")
		if err := sqlite.SanityCheck(); err != nil {
			fmt.Println("failed ❌", err)
			os.Exit(1)
		}

		fmt.Println("passed ✅")
		os.Exit(0)
	}

	// Initialize main DB connection.
	sqlDB := mustInitializeFrontendDB(logger, observationContext)
	db := database.NewDB(logger, sqlDB)

	// Run setup
	gitserverClient := gitserver.NewClient(db, observationContext)
	repositoryFetcherConfig := types.LoadRepositoryFetcherConfig(env.BaseConfig{})
	repositoryFetcher := fetcher.NewRepositoryFetcher(gitserverClient, repositoryFetcherConfig.MaxTotalPathsLength, int64(repositoryFetcherConfig.MaxFileSizeKb)*1000, observationContext)
	searchFunc, handleStatus, newRoutines, ctagsBinary, err := setup(observationContext, db, gitserverClient, repositoryFetcher)
	if err != nil {
		logger.Fatal("Failed to set up", log.Error(err))
	}
	routines = append(routines, newRoutines...)

	// Start debug server
	ready := make(chan struct{})
	go debugserver.NewServerRoutine(ready).Start()

	// Create HTTP server
	handler := api.NewHandler(searchFunc, gitserverClient.ReadFile, handleStatus, ctagsBinary)
	handler = trace.HTTPMiddleware(logger, handler, conf.DefaultClient())
	handler = ot.HTTPMiddleware(handler)
	handler = actor.HTTPMiddleware(handler)
	server := httpserver.NewFromAddr(addr, &http.Server{
		ReadTimeout:  75 * time.Second,
		WriteTimeout: 10 * time.Minute,
		Handler:      handler,
	})
	routines = append(routines, server)

	// Mark health server as ready and go!
	close(ready)
	goroutine.MonitorBackgroundRoutines(context.Background(), routines...)
}

func mustInitializeFrontendDB(logger log.Logger, observationContext *observation.Context) *sql.DB {
	dsn := conf.GetServiceConnectionValueAndRestartOnChange(func(serviceConnections conftypes.ServiceConnections) string {
		return serviceConnections.PostgresDSN
	})

	db, err := connections.EnsureNewFrontendDB(dsn, "symbols", observationContext)
	if err != nil {
		logger.Fatal("failed to connect to database", log.Error(err))
	}

	return db
}
