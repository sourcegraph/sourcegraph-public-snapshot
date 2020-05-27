package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/inconshreveable/log15"
	"github.com/opentracing/opentracing-go"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sourcegraph/sourcegraph/cmd/precise-code-intel-indexer/internal/indexer"
	indexscheduler "github.com/sourcegraph/sourcegraph/cmd/precise-code-intel-indexer/internal/scheduler/index"
	indexabilityscheduler "github.com/sourcegraph/sourcegraph/cmd/precise-code-intel-indexer/internal/scheduler/indexability"
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
		indexerPollInterval              = mustParseInterval(rawIndexerPollInterval, "PRECISE_CODE_INTEL_INDEXER_POLL_INTERVAL")
		indexSchedulerInterval           = mustParseInterval(rawIndexSchedulerInterval, "PRECISE_CODE_INTEL_INDEX_SCHEDULER_INTERVAL")
		indexabilitySchedulerInterval    = mustParseInterval(rawIndexabilitySchedulerInterval, "PRECISE_CODE_INTEL_INDEXABILITY_SCHEDULER_INTERVAL")
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
	indexerMetrics := indexer.NewIndexerMetrics(prometheus.DefaultRegisterer)
	indexabilitySchedulerMetrics := indexabilityscheduler.NewSchedulerMetrics(prometheus.DefaultRegisterer)
	indexSchedulerMetrics := indexscheduler.NewSchedulerMetrics(prometheus.DefaultRegisterer)
	server := server.New()

	indexabilityScheduler := indexabilityscheduler.NewScheduler(
		db,
		gitserver.DefaultClient,
		indexabilitySchedulerInterval,
		indexabilitySchedulerMetrics,
	)

	indexScheduler := indexscheduler.NewScheduler(
		db,
		gitserver.DefaultClient,
		indexSchedulerInterval,
		indexBatchSize,
		indexMinimumTimeSinceLastEnqueue,
		indexMinimumSearchCount,
		indexMinimumPreciseCount,
		float64(indexMinimumSearchRatio)/100,
		indexSchedulerMetrics,
	)

	indexer := indexer.NewIndexer(
		db,
		gitserver.DefaultClient,
		frontendURL,
		indexerPollInterval,
		indexerMetrics,
	)

	go server.Start()
	go indexabilityScheduler.Start()
	go indexScheduler.Start()
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
	indexScheduler.Stop()
	indexabilityScheduler.Stop()
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
