package main

import (
	"context"
	"database/sql"
	"net/http"
	"strings"
	"time"

	"github.com/sourcegraph/go-ctags"
	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/symbols/fetcher"
	symbolsGitserver "github.com/sourcegraph/sourcegraph/cmd/symbols/gitserver"
	symbolsParser "github.com/sourcegraph/sourcegraph/cmd/symbols/parser"
	"github.com/sourcegraph/sourcegraph/cmd/symbols/shared"
	"github.com/sourcegraph/sourcegraph/cmd/symbols/types"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/rockskip"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
	connections "github.com/sourcegraph/sourcegraph/internal/database/connections/live"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/search"
	"github.com/sourcegraph/sourcegraph/internal/search/result"
)

func main() {
	reposVar := env.Get("ROCKSKIP_REPOS", "", "comma separated list of repositories to index (e.g. `github.com/torvalds/linux,github.com/pallets/flask`)")
	repos := strings.Split(reposVar, ",")

	minRepoSizeMb := env.MustGetInt("ROCKSKIP_MIN_REPO_SIZE_MB", -1, "all repos that are at least this big will be indexed using Rockskip")

	repoToSize := map[string]int64{}

	if env.Get("USE_ROCKSKIP", "false", "use Rockskip to index the repos specified in ROCKSKIP_REPOS, or repos over ROCKSKIP_MIN_REPO_SIZE_MB in size") == "true" {
		shared.Main(func(observationContext *observation.Context, gitserverClient symbolsGitserver.GitserverClient, repositoryFetcher fetcher.RepositoryFetcher) (types.SearchFunc, func(http.ResponseWriter, *http.Request), []goroutine.BackgroundRoutine, string, error) {
			rockskipSearchFunc, rockskipHandleStatus, rockskipBackgroundRoutines, rockskipCtagsCommand, err := SetupRockskip(observationContext, gitserverClient, repositoryFetcher)
			if err != nil {
				return nil, nil, nil, "", err
			}

			// The blanks are the SQLite status endpoint (it's always nil) and the ctags command (same as
			// Rockskip's).
			sqliteSearchFunc, _, sqliteBackgroundRoutines, _, err := shared.SetupSqlite(observationContext, gitserverClient, repositoryFetcher)
			if err != nil {
				return nil, nil, nil, "", err
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
					if _, ok := repoToSize[string(args.Repo)]; ok {
						size = repoToSize[string(args.Repo)]
					} else {
						ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
						defer cancel()
						size, err = gitserverClient.GetRepoSize(ctx, args.Repo)
						if err != nil {
							return sqliteSearchFunc(ctx, args)
						}
						repoToSize[string(args.Repo)] = size
					}

					if size >= int64(minRepoSizeMb)*1000*1000 {
						return rockskipSearchFunc(ctx, args)
					} else {
						return sqliteSearchFunc(ctx, args)
					}
				}

				return sqliteSearchFunc(ctx, args)
			}

			return searchFunc, rockskipHandleStatus, append(rockskipBackgroundRoutines, sqliteBackgroundRoutines...), rockskipCtagsCommand, nil
		})
	} else {
		shared.Main(shared.SetupSqlite)
	}
}

func SetupRockskip(observationContext *observation.Context, gitserverClient symbolsGitserver.GitserverClient, repositoryFetcher fetcher.RepositoryFetcher) (types.SearchFunc, func(http.ResponseWriter, *http.Request), []goroutine.BackgroundRoutine, string, error) {
	logger := log.Scoped("rockskip", "rockskip-based symbols")

	baseConfig := env.BaseConfig{}
	config := LoadRockskipConfig(baseConfig)
	if err := baseConfig.Validate(); err != nil {
		logger.Fatal("failed to load configuration", log.Error(err))
	}

	codeintelDB := mustInitializeCodeIntelDB(logger)
	createParser := func() (ctags.Parser, error) {
		return symbolsParser.SpawnCtags(log.Scoped("parser", "ctags parser"), config.Ctags)
	}
	server, err := rockskip.NewService(codeintelDB, gitserverClient, repositoryFetcher, createParser, config.MaxConcurrentlyIndexing, config.MaxRepos, config.LogQueries, config.IndexRequestsQueueSize, config.SymbolsCacheSize, config.PathSymbolsCacheSize)
	if err != nil {
		return nil, nil, nil, config.Ctags.Command, err
	}

	return server.Search, server.HandleStatus, nil, config.Ctags.Command, nil
}

type RockskipConfig struct {
	Ctags                   types.CtagsConfig
	RepositoryFetcher       types.RepositoryFetcherConfig
	MaxRepos                int
	LogQueries              bool
	IndexRequestsQueueSize  int
	MaxConcurrentlyIndexing int
	SymbolsCacheSize        int
	PathSymbolsCacheSize    int
}

func LoadRockskipConfig(baseConfig env.BaseConfig) RockskipConfig {
	return RockskipConfig{
		Ctags:                   types.LoadCtagsConfig(baseConfig),
		RepositoryFetcher:       types.LoadRepositoryFetcherConfig(baseConfig),
		MaxRepos:                baseConfig.GetInt("MAX_REPOS", "1000", "maximum number of repositories to store in Postgres, with LRU eviction"),
		LogQueries:              baseConfig.GetBool("LOG_QUERIES", "false", "print search queries to stdout"),
		IndexRequestsQueueSize:  baseConfig.GetInt("INDEX_REQUESTS_QUEUE_SIZE", "1000", "how many index requests can be queued at once, at which point new requests will be rejected"),
		MaxConcurrentlyIndexing: baseConfig.GetInt("MAX_CONCURRENTLY_INDEXING", "4", "maximum number of repositories being indexed at a time (also limits ctags processes)"),
		SymbolsCacheSize:        baseConfig.GetInt("SYMBOLS_CACHE_SIZE", "100000", "how many tuples of (path, symbol name, int ID) to cache in memory"),
		PathSymbolsCacheSize:    baseConfig.GetInt("PATH_SYMBOLS_CACHE_SIZE", "10000", "how many sets of symbols for files to cache in memory"),
	}
}

func mustInitializeCodeIntelDB(logger log.Logger) *sql.DB {
	dsn := conf.GetServiceConnectionValueAndRestartOnChange(func(serviceConnections conftypes.ServiceConnections) string {
		return serviceConnections.CodeIntelPostgresDSN
	})
	var (
		db  *sql.DB
		err error
	)
	db, err = connections.EnsureNewCodeIntelDB(dsn, "symbols", &observation.TestContext)
	if err != nil {
		logger.Fatal("failed to connect to codeintel database", log.Error(err))
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
