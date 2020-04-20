package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/sourcegraph/sourcegraph/cmd/precise-code-intel-bundle-manager/internal/janitor"
	"github.com/sourcegraph/sourcegraph/cmd/precise-code-intel-bundle-manager/internal/paths"
	"github.com/sourcegraph/sourcegraph/cmd/precise-code-intel-bundle-manager/internal/server"
	"github.com/sourcegraph/sourcegraph/internal/debugserver"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/sqliteutil"
	"github.com/sourcegraph/sourcegraph/internal/tracer"
)

func main() {
	env.Lock()
	env.HandleHelpFlag()
	tracer.Init()

	sqliteutil.MustRegisterSqlite3WithPcre()

	var (
		bundleDir                = mustGet(rawBundleDir, "BUNDLE_DIR")
		databaseCacheSize        = mustParseInt(rawDatabaseCacheSize, "CONNECTION_CACHE_CAPACITY")
		documentDataCacheSize    = mustParseInt(rawDocumentDataCacheSize, "DOCUMENT_CACHE_CAPACITY")
		resultChunkDataCacheSize = mustParseInt(rawResultChunkDataCacheSize, "RESULT_CHUNK_CACHE_CAPACITY")
		desiredPercentFree       = mustParsePercent(rawDesiredPercentFree, "DESIRED_PERCENT_FREE")
		janitorInterval          = mustParseInterval(rawJanitorInterval, "JANITOR_INTERVAL")
		maxUnconvertedUploadAge  = mustParseInterval(rawMaxUnconvertedUploadAge, "MAX_UNCONVERTED_UPLOAD_AGE")
	)

	if err := paths.PrepDirectories(bundleDir); err != nil {
		log.Fatalf("failed to prepare directories: %s", err)
	}

	host := ""
	if env.InsecureDev {
		host = "127.0.0.1"
	}

	serverInst, err := server.New(server.ServerOpts{
		Host:                     host,
		Port:                     3187,
		BundleDir:                bundleDir,
		DatabaseCacheSize:        int64(databaseCacheSize),
		DocumentDataCacheSize:    int64(documentDataCacheSize),
		ResultChunkDataCacheSize: int64(resultChunkDataCacheSize),
	})
	if err != nil {
		log.Fatal(err)
	}

	janitorInst := janitor.NewJanitor(janitor.JanitorOpts{
		BundleDir:               bundleDir,
		DesiredPercentFree:      desiredPercentFree,
		JanitorInterval:         janitorInterval,
		MaxUnconvertedUploadAge: maxUnconvertedUploadAge,
	})

	go serverInst.Start()
	go janitorInst.Start()
	go debugserver.Start()
	waitForSignal()
}

func waitForSignal() {
	signals := make(chan os.Signal, 2)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGHUP)

	for i := 0; i < 2; i++ {
		<-signals
	}

	os.Exit(0)
}
