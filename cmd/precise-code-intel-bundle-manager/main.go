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
	"github.com/sourcegraph/sourcegraph/cmd/precise-code-intel-bundle-manager/internal/readers"
	"github.com/sourcegraph/sourcegraph/cmd/precise-code-intel-bundle-manager/internal/server"
	sqlitereader "github.com/sourcegraph/sourcegraph/internal/codeintel/bundles/persistence/sqlite"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/db"
	"github.com/sourcegraph/sourcegraph/internal/conf"
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
		bundleDir          = mustGet(rawBundleDir, "PRECISE_CODE_INTEL_BUNDLE_DIR")
		databaseCacheSize  = mustParseInt(rawDatabaseCacheSize, "PRECISE_CODE_INTEL_CONNECTION_CACHE_CAPACITY")
		readerCacheSize    = mustParseInt(rawReaderCacheSize, "PRECISE_CODE_INTEL_READER_CACHE_CAPACITY")
		desiredPercentFree = mustParsePercent(rawDesiredPercentFree, "PRECISE_CODE_INTEL_DESIRED_PERCENT_FREE")
		janitorInterval    = mustParseInterval(rawJanitorInterval, "PRECISE_CODE_INTEL_JANITOR_INTERVAL")
		maxUploadAge       = mustParseInterval(rawMaxUploadAge, "PRECISE_CODE_INTEL_MAX_UPLOAD_AGE")
		maxUploadPartAge   = mustParseInterval(rawMaxUploadPartAge, "PRECISE_CODE_INTEL_MAX_UPLOAD_PART_AGE")
		maxDatabasePartAge = mustParseInterval(rawMaxDatabasePartAge, "PRECISE_CODE_INTEL_MAX_DATABASE_PART_AGE")
	)

	observationContext := &observation.Context{
		Logger:     log15.Root(),
		Tracer:     &trace.Tracer{Tracer: opentracing.GlobalTracer()},
		Registerer: prometheus.DefaultRegisterer,
	}

	databaseCache, readerCache := prepCaches(observationContext.Registerer, databaseCacheSize, readerCacheSize)

	if err := paths.PrepDirectories(bundleDir); err != nil {
		log.Fatalf("failed to prepare directories: %s", err)
	}

	if err := paths.Migrate(bundleDir); err != nil {
		log.Fatalf("failed to migrate paths: %s", err)
	}

	if err := readers.Migrate(bundleDir, readerCache); err != nil {
		log.Fatalf("failed to migrate readers: %s", err)
	}

	db := db.NewObserved(mustInitializeDatabase(), observationContext)
	metrics.MustRegisterDiskMonitor(bundleDir)

	server := server.New(bundleDir, databaseCache, readerCache, observationContext)
	janitorMetrics := janitor.NewJanitorMetrics(prometheus.DefaultRegisterer)
	janitor := janitor.New(db, bundleDir, desiredPercentFree, janitorInterval, maxUploadAge, maxUploadPartAge, maxDatabasePartAge, janitorMetrics)

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

func prepCaches(r prometheus.Registerer, databaseCacheSize, readerCacheSize int) (*database.DatabaseCache, sqlitereader.Cache) {
	databaseCache, _, err := database.NewDatabaseCache(int64(databaseCacheSize))
	if err != nil {
		log.Fatal(errors.Wrap(err, "failed to initialize database cache"))
	}

	readerCache, err := sqlitereader.NewCache(int64(readerCacheSize))
	if err != nil {
		log.Fatal(errors.Wrap(err, "failed to initialize reader cache"))
	}

	return databaseCache, readerCache
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
