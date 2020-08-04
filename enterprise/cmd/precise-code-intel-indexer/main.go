package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/inconshreveable/log15"
	"github.com/opentracing/opentracing-go"
	"github.com/prometheus/client_golang/prometheus"
	indexmanager "github.com/sourcegraph/sourcegraph/enterprise/cmd/precise-code-intel-indexer/internal/index_manager"
	indexabilityupdater "github.com/sourcegraph/sourcegraph/enterprise/cmd/precise-code-intel-indexer/internal/indexability_updater"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/precise-code-intel-indexer/internal/indexer"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/precise-code-intel-indexer/internal/janitor"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/precise-code-intel-indexer/internal/resetter"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/precise-code-intel-indexer/internal/scheduler"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/precise-code-intel-indexer/internal/server"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/gitserver"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/store"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/debugserver"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/internal/tracer"
)

func main() {
	env.Lock()
	env.HandleHelpFlag()
	tracer.Init()

	var (
		frontendURL                      = mustGet(rawFrontendURL, "SRC_FRONTEND_INTERNAL")
		resetInterval                    = mustParseInterval(rawResetInterval, "PRECISE_CODE_INTEL_RESET_INTERVAL")
		indexerPollInterval              = mustParseInterval(rawIndexerPollInterval, "PRECISE_CODE_INTEL_INDEXER_POLL_INTERVAL")
		schedulerInterval                = mustParseInterval(rawSchedulerInterval, "PRECISE_CODE_INTEL_SCHEDULER_INTERVAL")
		indexabilityUpdaterInterval      = mustParseInterval(rawIndexabilityUpdaterInterval, "PRECISE_CODE_INTEL_INDEXABILITY_UPDATER_INTERVAL")
		janitorInterval                  = mustParseInterval(rawJanitorInterval, "PRECISE_CODE_INTEL_JANITOR_INTERVAL")
		indexBatchSize                   = mustParseInt(rawIndexBatchSize, "PRECISE_CODE_INTEL_INDEX_BATCH_SIZE")
		indexMinimumTimeSinceLastEnqueue = mustParseInterval(rawIndexMinimumTimeSinceLastEnqueue, "PRECISE_CODE_INTEL_INDEX_MINIMUM_TIME_SINCE_LAST_ENQUEUE")
		indexMinimumSearchCount          = mustParseInt(rawIndexMinimumSearchCount, "PRECISE_CODE_INTEL_INDEX_MINIMUM_SEARCH_COUNT")
		indexMinimumSearchRatio          = mustParsePercent(rawIndexMinimumSearchRatio, "PRECISE_CODE_INTEL_INDEX_MINIMUM_SEARCH_RATIO")
		indexMinimumPreciseCount         = mustParseInt(rawIndexMinimumPreciseCount, "PRECISE_CODE_INTEL_INDEX_MINIMUM_PRECISE_COUNT")
		disableJanitor                   = mustParseBool(rawDisableJanitor, "PRECISE_CODE_INTEL_DISABLE_JANITOR")
		maximumTransactions              = mustParseInt(rawMaxTransactions, "PRECISE_CODE_INTEL_MAXIMUM_TRANSACTIONS")
		requeueDelay                     = mustParseInterval(rawRequeueDelay, "PRECISE_CODE_INTEL_REQUEUE_DELAY")
		cleanupInterval                  = mustParseInterval(rawCleanupInterval, "PRECISE_CODE_INTEL_CLEANUP_INTERVAL")
		maximumMissedHeartbeats          = mustParseInt(rawMissedHeartbeats, "PRECISE_CODE_INTEL_MAXIMUM_MISSED_HEARTBEATS")
	)

	observationContext := &observation.Context{
		Logger:     log15.Root(),
		Tracer:     &trace.Tracer{Tracer: opentracing.GlobalTracer()},
		Registerer: prometheus.DefaultRegisterer,
	}

	s := store.NewObserved(mustInitializeStore(), observationContext)
	MustRegisterQueueMonitor(observationContext.Registerer, s)
	resetterMetrics := resetter.NewResetterMetrics(prometheus.DefaultRegisterer)
	indexabilityUpdaterMetrics := indexabilityupdater.NewUpdaterMetrics(prometheus.DefaultRegisterer)
	schedulerMetrics := scheduler.NewSchedulerMetrics(prometheus.DefaultRegisterer)
	indexerMetrics := indexer.NewIndexerMetrics(observationContext)
	indexManager := indexmanager.New(store.WorkerutilIndexStore(s), indexmanager.ManagerOptions{
		MaximumTransactions:   maximumTransactions,
		RequeueDelay:          requeueDelay,
		CleanupInterval:       cleanupInterval,
		UnreportedIndexMaxAge: cleanupInterval * time.Duration(maximumMissedHeartbeats),
		DeathThreshold:        cleanupInterval * time.Duration(maximumMissedHeartbeats),
	})
	server := server.New(indexManager)
	indexResetter := resetter.NewIndexResetter(s, resetInterval, resetterMetrics)

	indexabilityUpdater := indexabilityupdater.NewUpdater(
		s,
		gitserver.DefaultClient,
		indexabilityUpdaterInterval,
		indexabilityUpdaterMetrics,
	)

	scheduler := scheduler.NewScheduler(
		s,
		gitserver.DefaultClient,
		schedulerInterval,
		indexBatchSize,
		indexMinimumTimeSinceLastEnqueue,
		indexMinimumSearchCount,
		float64(indexMinimumSearchRatio)/100,
		indexMinimumPreciseCount,
		schedulerMetrics,
	)

	indexer := indexer.NewIndexer(
		s,
		gitserver.DefaultClient,
		frontendURL,
		indexerPollInterval,
		indexerMetrics,
	)

	janitorMetrics := janitor.NewJanitorMetrics(prometheus.DefaultRegisterer)
	janitor := janitor.New(s, janitorInterval, janitorMetrics)

	go server.Start()
	go indexResetter.Start()
	go indexabilityUpdater.Start()
	go scheduler.Start()
	go indexer.Start()
	go debugserver.Start()

	if !disableJanitor {
		go janitor.Run()
	} else {
		log15.Warn("Janitor process is disabled.")
	}

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
	indexResetter.Stop()
	indexer.Stop()
	scheduler.Stop()
	indexabilityUpdater.Stop()
	janitor.Stop()
}

func mustInitializeStore() store.Store {
	postgresDSN := conf.Get().ServiceConnections.PostgresDSN
	conf.Watch(func() {
		if newDSN := conf.Get().ServiceConnections.PostgresDSN; postgresDSN != newDSN {
			log.Fatalf("detected repository DSN change, restarting to take effect: %s", newDSN)
		}
	})

	store, err := store.New(postgresDSN)
	if err != nil {
		log.Fatalf("failed to initialize store: %s", err)
	}

	return store
}
