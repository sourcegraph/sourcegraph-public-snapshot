package shared

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/symbols/internal/api"
	sqlite "github.com/sourcegraph/sourcegraph/cmd/symbols/internal/database"
	"github.com/sourcegraph/sourcegraph/cmd/symbols/internal/fetcher"
	symbolsgitserver "github.com/sourcegraph/sourcegraph/cmd/symbols/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/cmd/symbols/internal/types"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
	"github.com/sourcegraph/sourcegraph/internal/database"
	connections "github.com/sourcegraph/sourcegraph/internal/database/connections/live"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/honey"
	"github.com/sourcegraph/sourcegraph/internal/httpserver"
	"github.com/sourcegraph/sourcegraph/internal/instrumentation"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/service"
	"github.com/sourcegraph/sourcegraph/internal/tenant"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	internaltypes "github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

var sanityCheck, _ = strconv.ParseBool(env.Get("SANITY_CHECK", "false", "check that go-sqlite3 works then exit 0 if it's ok or 1 if not"))

var (
	baseConfig              = env.BaseConfig{}
	RepositoryFetcherConfig types.RepositoryFetcherConfig
	CtagsConfig             types.CtagsConfig
)

const addr = ":3184"

type SetupFunc func(observationCtx *observation.Context, db database.DB, gitserverClient symbolsgitserver.GitserverClient, repositoryFetcher fetcher.RepositoryFetcher) (types.SearchFunc, func(http.ResponseWriter, *http.Request), []goroutine.BackgroundRoutine, error)

func Main(ctx context.Context, observationCtx *observation.Context, ready service.ReadyFunc, setup SetupFunc) error {
	logger := observationCtx.Logger

	routines := []goroutine.BackgroundRoutine{}

	// Initialize tracing/metrics
	observationCtx = observation.NewContext(logger, observation.Honeycomb(&honey.Dataset{
		Name:       "codeintel-symbols",
		SampleRate: 200,
	}))

	// Allow to do a sanity check of sqlite.
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
	sqlDB := mustInitializeFrontendDB(observationCtx)
	db := database.NewDB(logger, sqlDB)

	// Run setup
	gitserverClient := symbolsgitserver.NewClient(observationCtx, gitserver.NewClient("symbols"))
	repositoryFetcher := fetcher.NewRepositoryFetcher(observationCtx, gitserverClient, RepositoryFetcherConfig.MaxTotalPathsLength, int64(RepositoryFetcherConfig.MaxFileSizeKb)*1000)
	searchFunc, handleStatus, newRoutines, err := setup(observationCtx, db, gitserverClient, repositoryFetcher)
	if err != nil {
		return errors.Wrap(err, "failed to set up")
	}
	routines = append(routines, newRoutines...)

	// Create HTTP server
	handler := api.NewHandler(searchFunc, func(ctx context.Context, rcp internaltypes.RepoCommitPath) ([]byte, error) {
		r, err := gitserverClient.NewFileReader(ctx, rcp)
		if err != nil {
			return nil, err
		}
		defer r.Close()
		return io.ReadAll(r)
	}, handleStatus)

	handler = handlePanic(logger, handler)
	handler = trace.HTTPMiddleware(logger, handler)
	handler = instrumentation.HTTPMiddleware("", handler)
	handler = actor.HTTPMiddleware(logger, handler)
	handler = tenant.InternalHTTPMiddleware(logger, handler)
	server := httpserver.NewFromAddr(addr, &http.Server{
		ReadTimeout:  75 * time.Second,
		WriteTimeout: 10 * time.Minute,
		Handler:      handler,
	})
	routines = append(routines, server)

	// Mark health server as ready and go!
	ready()
	return goroutine.MonitorBackgroundRoutines(ctx, routines...)
}

func mustInitializeFrontendDB(observationCtx *observation.Context) *sql.DB {
	dsn := conf.GetServiceConnectionValueAndRestartOnChange(func(serviceConnections conftypes.ServiceConnections) string {
		return serviceConnections.PostgresDSN
	})

	db, err := connections.EnsureNewFrontendDB(observationCtx, dsn, "symbols")
	if err != nil {
		observationCtx.Logger.Fatal("failed to connect to database", log.Error(err))
	}

	return db
}

func handlePanic(logger log.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				err := fmt.Sprintf("%v", rec)
				http.Error(w, fmt.Sprintf("%v", rec), http.StatusInternalServerError)
				logger.Error("recovered from panic", log.String("err", err))
			}
		}()

		next.ServeHTTP(w, r)
	})
}
