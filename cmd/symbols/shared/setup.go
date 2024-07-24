package shared

import (
	"context"
	"database/sql"
	"net/http"
	"strings"
	"sync"

	"github.com/sourcegraph/go-ctags"
	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/symbols/internal/fetcher"
	symbolsGitserver "github.com/sourcegraph/sourcegraph/cmd/symbols/internal/gitserver"
	symbolparser "github.com/sourcegraph/sourcegraph/cmd/symbols/internal/parser"
	"github.com/sourcegraph/sourcegraph/cmd/symbols/internal/rockskip"
	"github.com/sourcegraph/sourcegraph/cmd/symbols/internal/types"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
	"github.com/sourcegraph/sourcegraph/internal/ctags_config"
	"github.com/sourcegraph/sourcegraph/internal/database"
	connections "github.com/sourcegraph/sourcegraph/internal/database/connections/live"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/search"
	"github.com/sourcegraph/sourcegraph/internal/search/result"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

var (
	useRockskip = env.MustGetBool("USE_ROCKSKIP", false, "use Rockskip to index the repos specified in ROCKSKIP_REPOS, or repos over ROCKSKIP_MIN_REPO_SIZE_MB in size")

	reposVar = env.Get("ROCKSKIP_REPOS", "", "comma separated list of repositories to index (e.g. `github.com/torvalds/linux,github.com/pallets/flask`)")
	repos    = strings.Split(reposVar, ",")

	minRepoSizeMb = env.MustGetInt("ROCKSKIP_MIN_REPO_SIZE_MB", -1, "all repos that are at least this big will be indexed using Rockskip")
)

func CreateSetup(config rockskipConfig) SetupFunc {
	var repoToSize sync.Map

	if useRockskip {
		return func(observationCtx *observation.Context, db database.DB, gitserverClient symbolsGitserver.GitserverClient, repositoryFetcher fetcher.RepositoryFetcher) (types.SearchFunc, func(http.ResponseWriter, *http.Request), []goroutine.BackgroundRoutine, error) {
			rockskipSearchFunc, rockskipHandleStatus, err := setupRockskip(observationCtx, config, gitserverClient, repositoryFetcher)
			if err != nil {
				return nil, nil, nil, err
			}

			// The blanks are the SQLite status endpoint (it's always nil) and the ctags command (same as
			// Rockskip's).
			sqliteSearchFunc, _, sqliteBackgroundRoutines, err := SetupSqlite(observationCtx, db, gitserverClient, repositoryFetcher)
			if err != nil {
				return nil, nil, nil, err
			}

			searchFunc := func(ctx context.Context, args search.SymbolsParameters) (results result.Symbols, err error) {
				if reposVar != "" {
					if sliceContains(repos, string(args.Repo)) {
						return rockskipSearchFunc(ctx, args)
					} else {
						return sqliteSearchFunc(ctx, args)
					}
				}

				if minRepoSizeMb != -1 {
					var size int64
					if value, ok := repoToSize.Load(args.Repo); ok {
						size = value.(int64)
					} else {
						info, err := db.GitserverRepos().GetByName(ctx, args.Repo)
						if err != nil {
							return sqliteSearchFunc(ctx, args)
						}
						size = info.RepoSizeBytes
						repoToSize.Store(args.Repo, size)
					}

					if size >= int64(minRepoSizeMb)*1000*1000 {
						return rockskipSearchFunc(ctx, args)
					} else {
						return sqliteSearchFunc(ctx, args)
					}
				}

				return sqliteSearchFunc(ctx, args)
			}

			return searchFunc, rockskipHandleStatus, sqliteBackgroundRoutines, nil
		}
	} else {
		return SetupSqlite
	}
}

type rockskipConfig struct {
	env.BaseConfig
	Ctags                   types.CtagsConfig
	NumCtagsProcesses       int
	RequestBufferSize       int
	RepositoryFetcher       types.RepositoryFetcherConfig
	MaxRepos                int
	LogQueries              bool
	IndexRequestsQueueSize  int
	MaxConcurrentlyIndexing int
	SymbolsCacheSize        int
	PathSymbolsCacheSize    int
	SearchLastIndexedCommit bool
}

func (c *rockskipConfig) Load() {
	// TODO(sqs): TODO(single-binary): load rockskip config from here
}

func loadRockskipConfig(baseConfig env.BaseConfig, ctags types.CtagsConfig, numCtagsProcesses int, requestBufferSize int, repositoryFetcher types.RepositoryFetcherConfig) rockskipConfig {
	return rockskipConfig{
		Ctags:                   ctags,
		NumCtagsProcesses:       numCtagsProcesses,
		RequestBufferSize:       requestBufferSize,
		RepositoryFetcher:       repositoryFetcher,
		MaxRepos:                baseConfig.GetInt("MAX_REPOS", "1000", "maximum number of repositories to store in Postgres, with LRU eviction"),
		LogQueries:              baseConfig.GetBool("LOG_QUERIES", "false", "print search queries to stdout"),
		IndexRequestsQueueSize:  baseConfig.GetInt("INDEX_REQUESTS_QUEUE_SIZE", "1000", "how many index requests can be queued at once, at which point new requests will be rejected"),
		MaxConcurrentlyIndexing: baseConfig.GetInt("ROCKSKIP_MAX_CONCURRENTLY_INDEXING", "4", "maximum number of repositories being indexed at a time (also limits ctags processes)"),
		SymbolsCacheSize:        baseConfig.GetInt("SYMBOLS_CACHE_SIZE", "100000", "how many tuples of (path, symbol name, int ID) to cache in memory while indexing a specific repo"),
		PathSymbolsCacheSize:    baseConfig.GetInt("PATH_SYMBOLS_CACHE_SIZE", "10000", "how many sets of symbols for files to cache in memory while indexing a specific repo"),
		SearchLastIndexedCommit: baseConfig.GetBool("SEARCH_LAST_INDEXED_COMMIT", "false", "falls back to searching the most recently indexed commit if the requested commit is not indexed"),
	}
}

func setupRockskip(observationCtx *observation.Context, config rockskipConfig, gitserverClient symbolsGitserver.GitserverClient, repositoryFetcher fetcher.RepositoryFetcher) (types.SearchFunc, func(http.ResponseWriter, *http.Request), error) {
	observationCtx = observation.ContextWithLogger(observationCtx.Logger.Scoped("rockskip"), observationCtx)

	codeintelDB := mustInitializeCodeIntelDB(observationCtx)

	logger := log.Scoped("parser")
	parserFactory := func(source ctags_config.ParserType) (ctags.Parser, error) {
		return symbolparser.SpawnCtags(logger, config.Ctags, source)
	}
	symbolParserPool, err := symbolparser.NewParserPool(observationCtx, "src_rockskip_service", parserFactory, config.NumCtagsProcesses, parserTypesForDeployment())
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to create symbol parser pool")
	}

	server, err := rockskip.NewService(observationCtx, codeintelDB, gitserverClient, repositoryFetcher, symbolParserPool, config.MaxConcurrentlyIndexing, config.MaxRepos, config.LogQueries, config.IndexRequestsQueueSize, config.SymbolsCacheSize, config.PathSymbolsCacheSize, config.SearchLastIndexedCommit)
	if err != nil {
		return nil, nil, err
	}

	return server.Search, server.HandleStatus, nil
}

func mustInitializeCodeIntelDB(observationCtx *observation.Context) *sql.DB {
	dsn := conf.GetServiceConnectionValueAndRestartOnChange(func(serviceConnections conftypes.ServiceConnections) string {
		return serviceConnections.CodeIntelPostgresDSN
	})
	var (
		db  *sql.DB
		err error
	)
	db, err = connections.EnsureNewCodeIntelDB(observationCtx, dsn, "symbols")
	if err != nil {
		observationCtx.Logger.Fatal("failed to connect to codeintel database", log.Error(err))
	}

	return db
}

func sliceContains(slice []string, s string) bool {
	for _, v := range slice {
		if v == s {
			return true
		}
	}
	return false
}
