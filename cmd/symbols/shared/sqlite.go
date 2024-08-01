package shared

import (
	"net/http"
	"time"

	"golang.org/x/sync/semaphore"

	"github.com/sourcegraph/go-ctags"
	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/symbols/internal/api"
	sqlite "github.com/sourcegraph/sourcegraph/cmd/symbols/internal/database"
	"github.com/sourcegraph/sourcegraph/cmd/symbols/internal/database/janitor"
	"github.com/sourcegraph/sourcegraph/cmd/symbols/internal/database/writer"
	"github.com/sourcegraph/sourcegraph/cmd/symbols/internal/fetcher"
	"github.com/sourcegraph/sourcegraph/cmd/symbols/internal/gitserver"
	symbolparser "github.com/sourcegraph/sourcegraph/cmd/symbols/internal/parser"
	"github.com/sourcegraph/sourcegraph/cmd/symbols/internal/types"
	"github.com/sourcegraph/sourcegraph/internal/ctags_config"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/diskcache"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

func LoadConfig() {
	RepositoryFetcherConfig = types.LoadRepositoryFetcherConfig(baseConfig)
	CtagsConfig = types.LoadCtagsConfig(baseConfig)
	config = types.LoadSqliteConfig(baseConfig, CtagsConfig, RepositoryFetcherConfig)
}

var config types.SqliteConfig

func SetupSqlite(observationCtx *observation.Context, db database.DB, gitserverClient gitserver.GitserverClient, repositoryFetcher fetcher.RepositoryFetcher) (types.SearchFunc, func(http.ResponseWriter, *http.Request), []goroutine.BackgroundRoutine, error) {
	logger := observationCtx.Logger.Scoped("sqlite.setup")

	if err := baseConfig.Validate(); err != nil {
		logger.Fatal("failed to load configuration", log.Error(err))
	}

	// Ensure we register our database driver before calling
	// anything that tries to open a SQLite database.
	sqlite.Init()

	parserFactory := func(source ctags_config.ParserType) (ctags.Parser, error) {
		return symbolparser.SpawnCtags(logger, config.Ctags, source)
	}

	parserPool, err := symbolparser.NewParserPool(observationCtx, "src", parserFactory, config.NumCtagsProcesses, parserTypesForDeployment())
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

	return searchFunc, nil, []goroutine.BackgroundRoutine{cacheEvicter}, nil
}

func parserTypesForDeployment() []ctags_config.ParserType {
	return symbolparser.DefaultParserTypes
}
