package shared

import (
	"context"
	"net/http"
	"time"

	"golang.org/x/sync/semaphore"

	"github.com/sourcegraph/go-ctags"
	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/symbols/fetcher"
	"github.com/sourcegraph/sourcegraph/cmd/symbols/gitserver"
	"github.com/sourcegraph/sourcegraph/cmd/symbols/internal/api"
	sqlite "github.com/sourcegraph/sourcegraph/cmd/symbols/internal/database"
	"github.com/sourcegraph/sourcegraph/cmd/symbols/internal/database/janitor"
	"github.com/sourcegraph/sourcegraph/cmd/symbols/internal/database/writer"
	symbolparser "github.com/sourcegraph/sourcegraph/cmd/symbols/parser"
	"github.com/sourcegraph/sourcegraph/cmd/symbols/types"
	"github.com/sourcegraph/sourcegraph/internal/conf/deploy"
	"github.com/sourcegraph/sourcegraph/internal/ctags_config"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/diskcache"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/search"
	"github.com/sourcegraph/sourcegraph/internal/search/result"
)

func LoadConfig() {
	RepositoryFetcherConfig = types.LoadRepositoryFetcherConfig(baseConfig)
	CtagsConfig = types.LoadCtagsConfig(baseConfig)
	config = types.LoadSqliteConfig(baseConfig, CtagsConfig, RepositoryFetcherConfig)
}

var config types.SqliteConfig

func SetupSqlite(observationCtx *observation.Context, db database.DB, gitserverClient gitserver.GitserverClient, repositoryFetcher fetcher.RepositoryFetcher) (types.SearchFunc, func(http.ResponseWriter, *http.Request), []goroutine.BackgroundRoutine, string, error) {
	logger := observationCtx.Logger.Scoped("sqlite.setup")

	if err := baseConfig.Validate(); err != nil {
		logger.Fatal("failed to load configuration", log.Error(err))
	}

	// Ensure we register our database driver before calling
	// anything that tries to open a SQLite database.
	sqlite.Init()

	if deploy.IsSingleBinary() && config.Ctags.UniversalCommand == "" {
		// app: ctags is not available
		searchFunc := func(ctx context.Context, params search.SymbolsParameters) (result.Symbols, error) {
			return nil, nil
		}
		return searchFunc, nil, []goroutine.BackgroundRoutine{}, "", nil
	}

	parserFactory := func(source ctags_config.ParserType) (ctags.Parser, error) {
		return symbolparser.SpawnCtags(logger, config.Ctags, source)
	}

	parserPool, err := symbolparser.NewParserPool(parserFactory, config.NumCtagsProcesses, parserTypesForDeployment())
	if err != nil {
		logger.Fatal("failed to create parser pool", log.Error(err))
	}

	cache := diskcache.NewStore(config.CacheDir, "symbols",
		diskcache.WithBackgroundTimeout(config.ProcessingTimeout),
		diskcache.WithobservationCtx(observationCtx),
	)

	parser := symbolparser.NewParser(observationCtx, parserPool, repositoryFetcher, config.RequestBufferSize, config.NumCtagsProcesses)
	databaseWriter := writer.NewDatabaseWriter(observationCtx, config.CacheDir, gitserverClient, parser, semaphore.NewWeighted(int64(config.MaxConcurrentlyIndexing)))
	cachedDatabaseWriter := writer.NewCachedDatabaseWriter(databaseWriter, cache)
	searchFunc := api.MakeSqliteSearchFunc(observationCtx, cachedDatabaseWriter, db)

	evictionInterval := time.Second * 10
	cacheSizeBytes := int64(config.CacheSizeMB) * 1000 * 1000
	cacheEvicter := janitor.NewCacheEvicter(evictionInterval, cache, cacheSizeBytes, janitor.NewMetrics(observationCtx))

	return searchFunc, nil, []goroutine.BackgroundRoutine{cacheEvicter}, config.Ctags.UniversalCommand, nil
}

func parserTypesForDeployment() []ctags_config.ParserType {
	if deploy.IsSingleBinary() {
		// ScipCtags is not available
		// TODO(burmudar): make it available
		return []ctags_config.ParserType{ctags_config.UniversalCtags}
	}

	return symbolparser.DefaultParserTypes
}
