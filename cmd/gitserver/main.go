// gitserver is the gitserver server.
package main // import "github.com/sourcegraph/sourcegraph/cmd/gitserver"

import (
	"container/list"
	"context"
	"database/sql"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/inconshreveable/log15"
	jsoniter "github.com/json-iterator/go"
	"github.com/opentracing/opentracing-go"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/sourcegraph/sourcegraph/cmd/gitserver/server"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	codeinteldbstore "github.com/sourcegraph/sourcegraph/internal/codeintel/stores/dbstore"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
	"github.com/sourcegraph/sourcegraph/internal/database"
	connections "github.com/sourcegraph/sourcegraph/internal/database/connections/live"
	"github.com/sourcegraph/sourcegraph/internal/debugserver"
	"github.com/sourcegraph/sourcegraph/internal/encryption/keyring"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/hostname"
	"github.com/sourcegraph/sourcegraph/internal/jsonc"
	"github.com/sourcegraph/sourcegraph/internal/logging"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/profiler"
	"github.com/sourcegraph/sourcegraph/internal/repos"
	"github.com/sourcegraph/sourcegraph/internal/sentry"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/internal/trace/ot"
	"github.com/sourcegraph/sourcegraph/internal/tracer"
	"github.com/sourcegraph/sourcegraph/schema"
)

var (
	reposDir                     = env.Get("SRC_REPOS_DIR", "/data/repos", "Root dir containing repos.")
	wantPctFree                  = env.MustGetInt("SRC_REPOS_DESIRED_PERCENT_FREE", 10, "Target percentage of free space on disk.")
	janitorInterval              = env.MustGetDuration("SRC_REPOS_JANITOR_INTERVAL", 1*time.Minute, "Interval between cleanup runs")
	syncRepoStateInterval        = env.MustGetDuration("SRC_REPOS_SYNC_STATE_INTERVAL", 10*time.Minute, "Interval between state syncs")
	syncRepoStateBatchSize       = env.MustGetInt("SRC_REPOS_SYNC_STATE_BATCH_SIZE", 500, "Number of upserts to perform per batch")
	syncRepoStateUpsertPerSecond = env.MustGetInt("SRC_REPOS_SYNC_STATE_UPSERT_PER_SEC", 500, "The number of upserted rows allowed per second across all gitserver instances")
)

func main() {
	ctx := context.Background()

	env.Lock()
	env.HandleHelpFlag()

	if err := profiler.Init(); err != nil {
		log.Fatalf("failed to start profiler: %v", err)
	}

	conf.Init()
	logging.Init()
	tracer.Init(conf.DefaultClient())
	sentry.Init(conf.DefaultClient())
	trace.Init()

	if reposDir == "" {
		log.Fatal("git-server: SRC_REPOS_DIR is required")
	}
	if err := os.MkdirAll(reposDir, os.ModePerm); err != nil {
		log.Fatalf("failed to create SRC_REPOS_DIR: %s", err)
	}

	wantPctFree2, err := getPercent(wantPctFree)
	if err != nil {
		log.Fatalf("SRC_REPOS_DESIRED_PERCENT_FREE is out of range: %v", err)
	}

	sqlDB, err := getDB()
	if err != nil {
		log.Fatalf("failed to initialize database stores: %v", err)
	}
	db := database.NewDB(sqlDB)

	repoStore := database.Repos(db)
	codeintelDB := codeinteldbstore.NewWithDB(db, &observation.Context{
		Logger:     log15.Root(),
		Tracer:     &trace.Tracer{Tracer: opentracing.GlobalTracer()},
		Registerer: prometheus.DefaultRegisterer,
	}, nil)
	externalServiceStore := database.ExternalServices(db)

	err = keyring.Init(ctx)
	if err != nil {
		log.Fatalf("failed to initialise keyring: %s", err)
	}

	gitserver := server.Server{
		ReposDir:           reposDir,
		DesiredPercentFree: wantPctFree2,
		GetRemoteURLFunc: func(ctx context.Context, repo api.RepoName) (string, error) {
			r, err := repoStore.GetByName(ctx, repo)
			if err != nil {
				return "", err
			}

			for _, info := range r.Sources {
				// build the clone url using the external service config instead of using
				// the source CloneURL field
				svc, err := externalServiceStore.GetByID(ctx, info.ExternalServiceID())
				if err != nil {
					return "", err
				}

				return repos.CloneURL(svc.Kind, svc.Config, r)
			}
			return "", errors.Errorf("no sources for %q", repo)
		},
		GetVCSSyncer: func(ctx context.Context, repo api.RepoName) (server.VCSSyncer, error) {
			return getVCSSyncer(ctx, externalServiceStore, repoStore, codeintelDB, repo)
		},
		Hostname:   hostname.Get(),
		DB:         db,
		CloneQueue: server.NewCloneQueue(list.New()),
	}
	gitserver.RegisterMetrics()

	if tmpDir, err := gitserver.SetupAndClearTmp(); err != nil {
		log.Fatalf("failed to setup temporary directory: %s", err)
	} else if err := os.Setenv("TMP_DIR", tmpDir); err != nil {
		// Additionally, set TMP_DIR so other temporary files we may accidentally
		// create are on the faster RepoDir mount.
		log.Fatalf("Setting TMP_DIR: %s", err)
	}

	// Create Handler now since it also initializes state
	// TODO: Why do we set server state as a side effect of creating our handler?
	handler := gitserver.Handler()
	handler = actor.HTTPMiddleware(handler)
	handler = ot.HTTPMiddleware(trace.HTTPMiddleware(handler, conf.DefaultClient()))

	// Ready immediately
	ready := make(chan struct{})
	close(ready)
	go debugserver.NewServerRoutine(ready).Start()
	go gitserver.Janitor(janitorInterval)
	go gitserver.SyncRepoState(syncRepoStateInterval, syncRepoStateBatchSize, syncRepoStateUpsertPerSecond)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	gitserver.StartClonePipeline(ctx)

	port := "3178"
	host := ""
	if env.InsecureDev {
		host = "127.0.0.1"
	}
	addr := net.JoinHostPort(host, port)
	srv := &http.Server{
		Addr:    addr,
		Handler: handler,
	}
	log15.Info("git-server: listening", "addr", srv.Addr)

	go func() {
		err := srv.ListenAndServe()
		if err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	// Listen for shutdown signals. When we receive one attempt to clean up,
	// but do an insta-shutdown if we receive more than one signal.
	c := make(chan os.Signal, 2)
	signal.Notify(c, syscall.SIGINT, syscall.SIGHUP, syscall.SIGTERM)

	// Once we receive one of the signals from above, continues with the shutdown
	// process.
	<-c
	go func() {
		// If a second signal is received, exit immediately.
		<-c
		os.Exit(0)
	}()

	// Wait for at most for the configured shutdown timeout.
	ctx, cancel = context.WithTimeout(ctx, goroutine.GracefulShutdownTimeout)
	defer cancel()
	// Stop accepting requests.
	if err := srv.Shutdown(ctx); err != nil {
		log15.Error("shutting down http server", "error", err)
	}

	// The most important thing this does is kill all our clones. If we just
	// shutdown they will be orphaned and continue running.
	gitserver.Stop()
}

func configureFusionClient(conn schema.PerforceConnection) server.FusionConfig {
	// Set up default settings first
	fc := server.FusionConfig{
		Enabled:             false,
		Client:              conn.P4Client,
		LookAhead:           2000,
		NetworkThreads:      12,
		NetworkThreadsFetch: 12,
		PrintBatch:          10,
		Refresh:             100,
		Retries:             10,
		MaxChanges:          -1,
		IncludeBinaries:     false,
	}

	if conn.FusionClient == nil {
		return fc
	}

	// Required
	fc.Enabled = conn.FusionClient.Enabled
	fc.LookAhead = conn.FusionClient.LookAhead

	// Optional
	if conn.FusionClient.NetworkThreads > 0 {
		fc.NetworkThreads = conn.FusionClient.NetworkThreads
	}
	if conn.FusionClient.NetworkThreadsFetch > 0 {
		fc.NetworkThreadsFetch = conn.FusionClient.NetworkThreadsFetch
	}
	if conn.FusionClient.PrintBatch > 0 {
		fc.PrintBatch = conn.FusionClient.PrintBatch
	}
	if conn.FusionClient.Refresh > 0 {
		fc.Refresh = conn.FusionClient.Refresh
	}
	if conn.FusionClient.Retries > 0 {
		fc.Retries = conn.FusionClient.Retries
	}
	if conn.FusionClient.MaxChanges > 0 {
		fc.MaxChanges = conn.FusionClient.MaxChanges
	}
	fc.IncludeBinaries = conn.FusionClient.IncludeBinaries

	return fc
}

func getPercent(p int) (int, error) {
	if p < 0 {
		return 0, errors.Errorf("negative value given for percentage: %d", p)
	}
	if p > 100 {
		return 0, errors.Errorf("excessively high value given for percentage: %d", p)
	}
	return p, nil
}

// getDB initializes a connection to the database and returns a dbutil.DB
func getDB() (*sql.DB, error) {
	// Gitserver is an internal actor. We rely on the frontend to do authz checks for
	// user requests.
	//
	// This call to SetProviders is here so that calls to GetProviders don't block.
	authz.SetProviders(true, []authz.Provider{})

	dsn := conf.GetServiceConnectionValueAndRestartOnChange(func(serviceConnections conftypes.ServiceConnections) string {
		return serviceConnections.PostgresDSN
	})
	var (
		sqlDB *sql.DB
		err   error
	)
	if os.Getenv("NEW_MIGRATIONS") == "" {
		// CURRENTLY DEPRECATING
		sqlDB, err = connections.NewFrontendDB(dsn, "gitserver", false, &observation.TestContext)
	} else {
		sqlDB, err = connections.EnsureNewFrontendDB(dsn, "gitserver", &observation.TestContext)
	}
	return sqlDB, err
}

func getVCSSyncer(ctx context.Context, externalServiceStore database.ExternalServiceStore, repoStore database.RepoStore,
	codeintelDB *codeinteldbstore.Store, repo api.RepoName) (server.VCSSyncer, error) {
	// We need an internal actor in case we are trying to access a private repo. We
	// only need access in order to find out the type of code host we're using, so
	// it's safe.
	r, err := repoStore.GetByName(actor.WithInternalActor(ctx), repo)
	if err != nil {
		return nil, errors.Wrap(err, "get repository")
	}

	extractOptions := func(connection interface{}) error {
		for _, info := range r.Sources {
			extSvc, err := externalServiceStore.GetByID(ctx, info.ExternalServiceID())
			if err != nil {
				return errors.Wrap(err, "get external service")
			}
			normalized, err := jsonc.Parse(extSvc.Config)
			if err != nil {
				return errors.Wrap(err, "normalize JSON")
			}
			if err = jsoniter.Unmarshal(normalized, connection); err != nil {
				return errors.Wrap(err, "unmarshal JSON")
			}
			return nil
		}
		return errors.Errorf("unexpected empty Sources map in %v", r)
	}

	switch r.ExternalRepo.ServiceType {
	case extsvc.TypePerforce:
		var c schema.PerforceConnection
		if err := extractOptions(&c); err != nil {
			return nil, err
		}
		return &server.PerforceDepotSyncer{
			MaxChanges:   int(c.MaxChanges),
			Client:       c.P4Client,
			FusionConfig: configureFusionClient(c),
		}, nil
	case extsvc.TypeJVMPackages:
		var c schema.JVMPackagesConnection
		if err := extractOptions(&c); err != nil {
			return nil, err
		}
		return &server.JVMPackagesSyncer{Config: &c, DBStore: codeintelDB}, nil
	case extsvc.TypeNPMPackages:
		var c schema.NPMPackagesConnection
		if err := extractOptions(&c); err != nil {
			return nil, err
		}
		return &server.NPMPackagesSyncer{Config: &c, DBStore: codeintelDB}, nil
	}
	return &server.GitRepoSyncer{}, nil
}
