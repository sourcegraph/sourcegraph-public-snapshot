package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/dgraph-io/ristretto"
	"github.com/inconshreveable/log15"
	"github.com/opentracing/opentracing-go"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sourcegraph/sourcegraph/cmd/precise-code-intel-bundle-manager/internal/database"
	"github.com/sourcegraph/sourcegraph/cmd/precise-code-intel-bundle-manager/internal/janitor"
	"github.com/sourcegraph/sourcegraph/cmd/precise-code-intel-bundle-manager/internal/paths"
	"github.com/sourcegraph/sourcegraph/cmd/precise-code-intel-bundle-manager/internal/server"
	"github.com/sourcegraph/sourcegraph/internal/debugserver"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/metrics"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/sqliteutil"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/internal/tracer"
)

func main() {
	host := ""
	if env.InsecureDev {
		host = "127.0.0.1"
	}

	env.Lock()
	env.HandleHelpFlag()
	tracer.Init()

	sqliteutil.MustRegisterSqlite3WithPcre()

	var (
		bundleDir                = mustGet(rawBundleDir, "PRECISE_CODE_INTEL_BUNDLE_DIR")
		databaseCacheSize        = mustParseInt(rawDatabaseCacheSize, "PRECISE_CODE_INTEL_CONNECTION_CACHE_CAPACITY")
		documentDataCacheSize    = mustParseInt(rawDocumentDataCacheSize, "PRECISE_CODE_INTEL_DOCUMENT_CACHE_CAPACITY")
		resultChunkDataCacheSize = mustParseInt(rawResultChunkDataCacheSize, "PRECISE_CODE_INTEL_RESULT_CHUNK_CACHE_CAPACITY")
		desiredPercentFree       = mustParsePercent(rawDesiredPercentFree, "PRECISE_CODE_INTEL_DESIRED_PERCENT_FREE")
		janitorInterval          = mustParseInterval(rawJanitorInterval, "PRECISE_CODE_INTEL_JANITOR_INTERVAL")
		maxUploadAge             = mustParseInterval(rawMaxUploadAge, "PRECISE_CODE_INTEL_MAX_UPLOAD_AGE")
	)

	if err := paths.PrepDirectories(bundleDir); err != nil {
		log.Fatalf("failed to prepare directories: %s", err)
	}

	observationContext := &observation.Context{
		Logger:     log15.Root(),
		Tracer:     &trace.Tracer{Tracer: opentracing.GlobalTracer()},
		Registerer: prometheus.DefaultRegisterer,
	}

	metrics.MustRegisterDiskMonitor(bundleDir)
	janitorMetrics := janitor.NewJanitorMetrics(prometheus.DefaultRegisterer)
	databaseCache, documentDataCache, resultChunkDataCache := prepCaches(
		observationContext.Registerer,
		databaseCacheSize,
		documentDataCacheSize,
		resultChunkDataCacheSize,
	)

	server := server.Server{
		Host:                 host,
		Port:                 3187,
		BundleDir:            bundleDir,
		DatabaseCache:        databaseCache,
		DocumentDataCache:    documentDataCache,
		ResultChunkDataCache: resultChunkDataCache,
		ObservationContext:   observationContext,
	}
	go server.Start()

	janitor := janitor.Janitor{
		BundleDir:          bundleDir,
		DesiredPercentFree: desiredPercentFree,
		JanitorInterval:    janitorInterval,
		MaxUploadAge:       maxUploadAge,
		Metrics:            janitorMetrics,
	}
	go janitor.Run()

	go debugserver.Start()
	waitForSignal()
}

func prepCaches(r prometheus.Registerer, databaseCacheSize, documentDataCacheSize, resultChunkDataCacheSize int) (
	*database.DatabaseCache,
	*database.DocumentDataCache,
	*database.ResultChunkDataCache,
) {
	databaseCache, databaseCacheMetrics, err := database.NewDatabaseCache(int64(databaseCacheSize))
	if err != nil {
		log.Fatal(errors.Wrap(err, "failed to initialize database cache"))
	}

	documentDataCache, documentDataCacheMetrics, err := database.NewDocumentDataCache(int64(documentDataCacheSize))
	if err != nil {
		log.Fatal(errors.Wrap(err, "failed to initialize document cache"))
	}

	resultChunkDataCache, resultChunkDataCacheMetrics, err := database.NewResultChunkDataCache(int64(resultChunkDataCacheSize))
	if err != nil {
		log.Fatal(errors.Wrap(err, "failed to initialize result chunk cache"))
	}

	cacheMetrics := map[string]*ristretto.Metrics{
		"precise-code-intel-database":          databaseCacheMetrics,
		"precise-code-intel-document-data":     documentDataCacheMetrics,
		"precise-code-intel-result-chunk-data": resultChunkDataCacheMetrics,
	}
	for cacheName, metrics := range cacheMetrics {
		MustRegisterCacheMonitor(r, cacheName, metrics)
	}

	return databaseCache, documentDataCache, resultChunkDataCache
}

func waitForSignal() {
	signals := make(chan os.Signal, 2)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGHUP)

	for i := 0; i < 2; i++ {
		<-signals
	}

	os.Exit(0)
}
