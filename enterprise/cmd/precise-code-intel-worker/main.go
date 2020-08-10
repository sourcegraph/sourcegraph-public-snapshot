package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/inconshreveable/log15"
	"github.com/opentracing/opentracing-go"
	"github.com/prometheus/client_golang/prometheus"
	commitupdater "github.com/sourcegraph/sourcegraph/enterprise/cmd/precise-code-intel-worker/internal/commit-updater"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/precise-code-intel-worker/internal/metrics"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/precise-code-intel-worker/internal/resetter"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/precise-code-intel-worker/internal/server"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/precise-code-intel-worker/internal/worker"
	bundles "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/bundles/client"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/commits"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/gitserver"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/store"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/db/basestore"
	"github.com/sourcegraph/sourcegraph/internal/db/dbconn"
	"github.com/sourcegraph/sourcegraph/internal/debugserver"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/sqliteutil"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/internal/tracer"
)

func main() {
	env.Lock()
	env.HandleHelpFlag()
	tracer.Init()

	sqliteutil.MustRegisterSqlite3WithPcre()

	var (
		bundleManagerURL      = mustGet(rawBundleManagerURL, "PRECISE_CODE_INTEL_BUNDLE_MANAGER_URL")
		workerPollInterval    = mustParseInterval(rawWorkerPollInterval, "PRECISE_CODE_INTEL_WORKER_POLL_INTERVAL")
		workerConcurrency     = mustParseInt(rawWorkerConcurrency, "PRECISE_CODE_INTEL_WORKER_CONCURRENCY")
		workerBudget          = mustParseInt64(rawWorkerBudget, "PRECISE_CODE_INTEL_WORKER_BUDGET")
		resetInterval         = mustParseInterval(rawResetInterval, "PRECISE_CODE_INTEL_RESET_INTERVAL")
		commitUpdaterInterval = mustParseInterval(rawCommitUpdaterInterval, "PRECISE_CODE_INTEL_COMMIT_UPDATER_INTERVAL")
	)

	observationContext := &observation.Context{
		Logger:     log15.Root(),
		Tracer:     &trace.Tracer{Tracer: opentracing.GlobalTracer()},
		Registerer: prometheus.DefaultRegisterer,
	}

	store := store.NewObserved(mustInitializeStore(), observationContext)
	MustRegisterQueueMonitor(observationContext.Registerer, store)
	workerMetrics := metrics.NewWorkerMetrics(observationContext)
	resetterMetrics := resetter.NewResetterMetrics(prometheus.DefaultRegisterer)
	server := server.New()
	uploadResetter := resetter.NewUploadResetter(store, resetInterval, resetterMetrics)
	commitUpdater := commitupdater.NewUpdater(
		store,
		commits.NewUpdater(store, gitserver.DefaultClient),
		commitupdater.UpdaterOptions{
			Interval: commitUpdaterInterval,
		},
	)
	worker := worker.NewWorker(
		store,
		bundles.New(bundleManagerURL),
		gitserver.DefaultClient,
		workerPollInterval,
		workerConcurrency,
		workerBudget,
		workerMetrics,
	)

	go server.Start()
	go uploadResetter.Start()
	go commitUpdater.Start()
	go worker.Start()
	go debugserver.Start()

	// Attempt to clean up after first shutdown signal
	signals := make(chan os.Signal, 2)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGHUP)
	<-signals

	go func() {
		// Insta-shutdown on a second signal
		<-signals
		os.Exit(0)
	}()

	server.Stop()
	uploadResetter.Stop()
	commitUpdater.Stop()
	worker.Stop()
}

func mustInitializeStore() store.Store {
	postgresDSN := conf.Get().ServiceConnections.PostgresDSN
	conf.Watch(func() {
		if newDSN := conf.Get().ServiceConnections.PostgresDSN; postgresDSN != newDSN {
			log.Fatalf("detected repository DSN change, restarting to take effect: %s", newDSN)
		}
	})

	if err := dbconn.ConnectToDB(postgresDSN); err != nil {
		log.Fatalf("failed to connect to database: %s", err)
	}

	return store.NewWithHandle(basestore.NewHandleWithDB(dbconn.Global))
}
