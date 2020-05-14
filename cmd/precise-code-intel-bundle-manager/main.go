package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

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
	env.Lock()
	env.HandleHelpFlag()
	tracer.Init()

	sqliteutil.MustRegisterSqlite3WithPcre()

	var (
		bundleDir            = mustGet(rawBundleDir, "PRECISE_CODE_INTEL_BUNDLE_DIR")
		databaseCacheSize    = mustParseInt(rawDatabaseCacheSize, "PRECISE_CODE_INTEL_CONNECTION_CACHE_CAPACITY")
		documentCacheSize    = mustParseInt(rawDocumentCacheSize, "PRECISE_CODE_INTEL_DOCUMENT_CACHE_CAPACITY")
		resultChunkCacheSize = mustParseInt(rawResultChunkCacheSize, "PRECISE_CODE_INTEL_RESULT_CHUNK_CACHE_CAPACITY")
		desiredPercentFree   = mustParsePercent(rawDesiredPercentFree, "PRECISE_CODE_INTEL_DESIRED_PERCENT_FREE")
		janitorInterval      = mustParseInterval(rawJanitorInterval, "PRECISE_CODE_INTEL_JANITOR_INTERVAL")
		maxUploadAge         = mustParseInterval(rawMaxUploadAge, "PRECISE_CODE_INTEL_MAX_UPLOAD_AGE")
	)

	if err := paths.PrepDirectories(bundleDir); err != nil {
		log.Fatalf("failed to prepare directories: %s", err)
	}

	observationContext := &observation.Context{
		Logger:     log15.Root(),
		Tracer:     &trace.Tracer{Tracer: opentracing.GlobalTracer()},
		Registerer: prometheus.DefaultRegisterer,
	}

	databaseCache, documentCache, resultChunkCache := prepCaches(
		observationContext.Registerer,
		databaseCacheSize,
		documentCacheSize,
		resultChunkCacheSize,
	)

	metrics.MustRegisterDiskMonitor(bundleDir)
	janitorMetrics := janitor.NewJanitorMetrics(prometheus.DefaultRegisterer)
	server := server.New(bundleDir, databaseCache, documentCache, resultChunkCache, observationContext)
	janitor := janitor.New(bundleDir, desiredPercentFree, janitorInterval, maxUploadAge, janitorMetrics)

	go server.Start()
	go janitor.Run()
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
	janitor.Stop()
}

func prepCaches(r prometheus.Registerer, databaseCacheSize, documentCacheSize, resultChunkCacheSize int) (
	*database.DatabaseCache,
	*database.DocumentCache,
	*database.ResultChunkCache,
) {
	databaseCache, databaseCacheMetrics, err := database.NewDatabaseCache(int64(databaseCacheSize))
	if err != nil {
		log.Fatal(errors.Wrap(err, "failed to initialize database cache"))
	}

	documentCache, documentCacheMetrics, err := database.NewDocumentCache(int64(documentCacheSize))
	if err != nil {
		log.Fatal(errors.Wrap(err, "failed to initialize document cache"))
	}

	resultChunkCache, resultChunkCacheMetrics, err := database.NewResultChunkCache(int64(resultChunkCacheSize))
	if err != nil {
		log.Fatal(errors.Wrap(err, "failed to initialize result chunk cache"))
	}

	MustRegisterCacheMonitor(r, "precise-code-intel-database", databaseCacheSize, databaseCacheMetrics)
	MustRegisterCacheMonitor(r, "precise-code-intel-document", documentCacheSize, documentCacheMetrics)
	MustRegisterCacheMonitor(r, "precise-code-intel-result-chunk", resultChunkCacheSize, resultChunkCacheMetrics)

	return databaseCache, documentCache, resultChunkCache
}
