// Command symbols is a service that serves code symbols (functions, variables, etc.) from a repository at a
// specific commit.
package main

import (
	"archive/tar"
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"path"
	"time"

	"github.com/inconshreveable/log15"
	"github.com/opentracing/opentracing-go"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/sourcegraph/sourcegraph/cmd/symbols/internal/api"
	sqlite "github.com/sourcegraph/sourcegraph/cmd/symbols/internal/database"
	"github.com/sourcegraph/sourcegraph/cmd/symbols/internal/database/janitor"
	"github.com/sourcegraph/sourcegraph/cmd/symbols/internal/database/writer"
	"github.com/sourcegraph/sourcegraph/cmd/symbols/internal/fetcher"
	"github.com/sourcegraph/sourcegraph/cmd/symbols/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/cmd/symbols/internal/parser"
	"github.com/sourcegraph/sourcegraph/cmd/symbols/internal/types"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/debugserver"
	"github.com/sourcegraph/sourcegraph/internal/diskcache"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/honey"
	"github.com/sourcegraph/sourcegraph/internal/httpserver"
	"github.com/sourcegraph/sourcegraph/internal/logging"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/profiler"
	"github.com/sourcegraph/sourcegraph/internal/sentry"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/internal/trace/ot"
	"github.com/sourcegraph/sourcegraph/internal/tracer"
)

const addr = ":3184"

func main() {
	config.Load()

	// Set up Google Cloud Profiler when running in Cloud
	if err := profiler.Init(); err != nil {
		log.Fatalf("Failed to start profiler: %v", err)
	}

	env.Lock()
	env.HandleHelpFlag()
	conf.Init()
	logging.Init()
	tracer.Init(conf.DefaultClient())
	sentry.Init(conf.DefaultClient())
	trace.Init()

	if err := config.Validate(); err != nil {
		log.Fatalf("Failed to load configuration: %s", err)
	}

	// Ensure we register our database driver before calling
	// anything that tries to open a SQLite database.
	sqlite.Init()

	if config.sanityCheck {
		fmt.Print("Running sanity check...")
		if err := sqlite.SanityCheck(); err != nil {
			fmt.Println("failed ❌", err)
			os.Exit(1)
		}

		fmt.Println("passed ✅")
		os.Exit(0)
	}

	// Initialize tracing/metrics
	observationContext := &observation.Context{
		Logger:     log15.Root(),
		Tracer:     &trace.Tracer{Tracer: opentracing.GlobalTracer()},
		Registerer: prometheus.DefaultRegisterer,
		HoneyDataset: &honey.Dataset{
			Name:       "codeintel-symbols",
			SampleRate: 5,
		},
	}

	// Start debug server
	ready := make(chan struct{})
	go debugserver.NewServerRoutine(ready).Start()

	ctagsParserFactory := parser.NewCtagsParserFactory(
		config.ctags.Command,
		config.ctags.PatternLengthLimit,
		config.ctags.LogErrors,
		config.ctags.DebugLogs,
	)

	cache := diskcache.NewStore(config.cacheDir, "symbols",
		diskcache.WithBackgroundTimeout(config.processingTimeout),
		diskcache.WithObservationContext(observationContext),
	)

	parserPool, err := parser.NewParserPool(ctagsParserFactory, config.numCtagsProcesses)
	if err != nil {
		log.Fatalf("Failed to create parser pool: %s", err)
	}

	var searchFunc types.SearchFunc
	if config.useRockskip {
		searchFunc, err = api.MakeRockskipSearchFunc(observationContext, config.ctags, config.maxRepos)
		if err != nil {
			log.Fatalf("Failed to create rockskip search function: %s", err)
		}
	} else {
		gitserverClient := gitserver.NewClient(observationContext)

		shouldRead := func(tarHeader *tar.Header) bool {
			// We do not search large files over 512KiB
			if tarHeader.Size > 524288 {
				return false
			}

			// We only care about files
			if tarHeader.Typeflag != tar.TypeReg && tarHeader.Typeflag != tar.TypeRegA {
				return false
			}

			// JSON files are symbol-less
			if path.Ext(tarHeader.Name) == ".json" {
				return false
			}

			return true
		}

		repositoryFetcher := fetcher.NewRepositoryFetcher(gitserverClient, 15, config.maxTotalPathsLength, observationContext, shouldRead)
		parser := parser.NewParser(parserPool, repositoryFetcher, config.requestBufferSize, config.numCtagsProcesses, observationContext)
		databaseWriter := writer.NewDatabaseWriter(config.cacheDir, gitserverClient, parser)
		cachedDatabaseWriter := writer.NewCachedDatabaseWriter(databaseWriter, cache)
		searchFunc = api.MakeSqliteSearchFunc(api.NewOperations(observationContext), cachedDatabaseWriter)
	}
	apiHandler := api.NewHandler(searchFunc, config.ctags.Command)

	server := httpserver.NewFromAddr(addr, &http.Server{
		ReadTimeout:  75 * time.Second,
		WriteTimeout: 10 * time.Minute,
		Handler:      actor.HTTPMiddleware(ot.HTTPMiddleware(trace.HTTPMiddleware(apiHandler, conf.DefaultClient()))),
	})

	evictionInterval := time.Second * 10
	cacheSizeBytes := int64(config.cacheSizeMB) * 1000 * 1000
	cacheEvicter := janitor.NewCacheEvicter(evictionInterval, cache, cacheSizeBytes, janitor.NewMetrics(observationContext))

	// Mark health server as ready and go!
	close(ready)
	goroutine.MonitorBackgroundRoutines(context.Background(), server, cacheEvicter)
}
