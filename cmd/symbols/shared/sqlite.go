package shared

import (
	"archive/tar"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/sourcegraph/sourcegraph/cmd/symbols/fetcher"
	"github.com/sourcegraph/sourcegraph/cmd/symbols/gitserver"
	"github.com/sourcegraph/sourcegraph/cmd/symbols/internal/api"
	sqlite "github.com/sourcegraph/sourcegraph/cmd/symbols/internal/database"
	"github.com/sourcegraph/sourcegraph/cmd/symbols/internal/database/janitor"
	"github.com/sourcegraph/sourcegraph/cmd/symbols/internal/database/writer"
	"github.com/sourcegraph/sourcegraph/cmd/symbols/internal/parser"
	"github.com/sourcegraph/sourcegraph/cmd/symbols/observability"
	"github.com/sourcegraph/sourcegraph/cmd/symbols/types"
	"github.com/sourcegraph/sourcegraph/internal/diskcache"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

func SetupSqlite(observationContext *observation.Context) (types.SearchFunc, []goroutine.BackgroundRoutine, string) {
	baseConfig := env.BaseConfig{}
	config := LoadSqliteConfig(baseConfig)
	if err := baseConfig.Validate(); err != nil {
		log.Fatalf("Failed to load configuration: %s", err)
	}

	// Ensure we register our database driver before calling
	// anything that tries to open a SQLite database.
	sqlite.Init()

	if config.SanityCheck {
		fmt.Print("Running sanity check...")
		if err := sqlite.SanityCheck(); err != nil {
			fmt.Println("failed ❌", err)
			os.Exit(1)
		}

		fmt.Println("passed ✅")
		os.Exit(0)
	}

	ctagsParserFactory := parser.NewCtagsParserFactory(
		config.Ctags.Command,
		config.Ctags.PatternLengthLimit,
		config.Ctags.LogErrors,
		config.Ctags.DebugLogs,
	)

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

		return true
	}

	parserPool, err := parser.NewParserPool(ctagsParserFactory, config.NumCtagsProcesses)
	if err != nil {
		log.Fatalf("Failed to create parser pool: %s", err)
	}

	cache := diskcache.NewStore(config.CacheDir, "symbols",
		diskcache.WithBackgroundTimeout(config.ProcessingTimeout),
		diskcache.WithObservationContext(observationContext),
	)

	repositoryFetcher := fetcher.NewRepositoryFetcher(gitserverClient, 15, config.RepositoryFetcher.MaxTotalPathsLength, observationContext, shouldRead)
	parser := parser.NewParser(parserPool, repositoryFetcher, config.RequestBufferSize, config.NumCtagsProcesses, observationContext)
	databaseWriter := writer.NewDatabaseWriter(config.CacheDir, gitserverClient, parser)
	cachedDatabaseWriter := writer.NewCachedDatabaseWriter(databaseWriter, cache)
	searchFunc := api.MakeSqliteSearchFunc(observability.NewOperations(observationContext), cachedDatabaseWriter)

	evictionInterval := time.Second * 10
	cacheSizeBytes := int64(config.CacheSizeMB) * 1000 * 1000
	cacheEvicter := janitor.NewCacheEvicter(evictionInterval, cache, cacheSizeBytes, janitor.NewMetrics(observationContext))

	return searchFunc, []goroutine.BackgroundRoutine{cacheEvicter}, config.Ctags.Command
}
