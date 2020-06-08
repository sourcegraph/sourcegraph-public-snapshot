package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/inconshreveable/log15"
	"github.com/opentracing/opentracing-go"
	"github.com/prometheus/client_golang/prometheus"
	indexabilityupdater "github.com/sourcegraph/sourcegraph/cmd/precise-code-intel-indexer/internal/indexability_updater"
	"github.com/sourcegraph/sourcegraph/cmd/precise-code-intel-indexer/internal/indexer"
	"github.com/sourcegraph/sourcegraph/cmd/precise-code-intel-indexer/internal/resetter"
	"github.com/sourcegraph/sourcegraph/cmd/precise-code-intel-indexer/internal/scheduler"
	"github.com/sourcegraph/sourcegraph/cmd/precise-code-intel-indexer/internal/server"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/db"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/gitserver"
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
		indexBatchSize                   = mustParseInt(rawIndexBatchSize, "PRECISE_CODE_INTEL_INDEX_BATCH_SIZE")
		indexMinimumTimeSinceLastEnqueue = mustParseInterval(rawIndexMinimumTimeSinceLastEnqueue, "PRECISE_CODE_INTEL_INDEX_MINIMUM_TIME_SINCE_LAST_ENQUEUE")
		indexMinimumSearchCount          = mustParseInt(rawIndexMinimumSearchCount, "PRECISE_CODE_INTEL_INDEX_MINIMUM_SEARCH_COUNT")
		indexMinimumPreciseCount         = mustParseInt(rawIndexMinimumPreciseCount, "PRECISE_CODE_INTEL_INDEX_MINIMUM_PRECISE_COUNT")
		indexMinimumSearchRatio          = mustParsePercent(rawIndexMinimumSearchRatio, "PRECISE_CODE_INTEL_INDEX_MINIMUM_SEARCH_RATIO")
	)

	observationContext := &observation.Context{
		Logger:     log15.Root(),
		Tracer:     &trace.Tracer{Tracer: opentracing.GlobalTracer()},
		Registerer: prometheus.DefaultRegisterer,
	}

	db := db.NewObserved(mustInitializeDatabase(), observationContext)
	MustRegisterQueueMonitor(observationContext.Registerer, db)
	resetterMetrics := resetter.NewResetterMetrics(prometheus.DefaultRegisterer)
	indexabilityUpdaterMetrics := indexabilityupdater.NewUpdaterMetrics(prometheus.DefaultRegisterer)
	schedulerMetrics := scheduler.NewSchedulerMetrics(prometheus.DefaultRegisterer)
	indexerMetrics := indexer.NewIndexerMetrics(prometheus.DefaultRegisterer)
	server := server.New()

	indexResetter := resetter.IndexResetter{
		DB:            db,
		ResetInterval: resetInterval,
		Metrics:       resetterMetrics,
	}

	indexabilityUpdater := indexabilityupdater.NewUpdater(
		db,
		gitserver.DefaultClient,
		indexabilityUpdaterInterval,
		indexabilityUpdaterMetrics,
	)

	scheduler := scheduler.NewScheduler(
		db,
		gitserver.DefaultClient,
		schedulerInterval,
		indexBatchSize,
		indexMinimumTimeSinceLastEnqueue,
		indexMinimumSearchCount,
		indexMinimumPreciseCount,
		float64(indexMinimumSearchRatio)/100,
		schedulerMetrics,
	)

	indexer := indexer.NewIndexer(
		db,
		gitserver.DefaultClient,
		frontendURL,
		indexerPollInterval,
		indexerMetrics,
	)

	go server.Start()
	go indexResetter.Run()
	go indexabilityUpdater.Start()
	go scheduler.Start()
	go indexer.Start()
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
	indexer.Stop()
	scheduler.Stop()
	indexabilityUpdater.Stop()
}

func mustInitializeDatabase() db.DB {
	postgresDSN := conf.Get().ServiceConnections.PostgresDSN
	conf.Watch(func() {
		if newDSN := conf.Get().ServiceConnections.PostgresDSN; postgresDSN != newDSN {
			log.Fatalf("detected repository DSN change, restarting to take effect: %s", newDSN)
		}
	})

	db, err := db.New(postgresDSN)
	if err != nil {
		log.Fatalf("failed to initialize db store: %s", err)
	}

	return db
}
