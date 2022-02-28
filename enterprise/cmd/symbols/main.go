package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"os"

	"github.com/sourcegraph/go-ctags"

	"github.com/sourcegraph/sourcegraph/cmd/symbols/fetcher"
	symbolsGitserver "github.com/sourcegraph/sourcegraph/cmd/symbols/gitserver"
	symbolsParser "github.com/sourcegraph/sourcegraph/cmd/symbols/parser"
	"github.com/sourcegraph/sourcegraph/cmd/symbols/shared"
	"github.com/sourcegraph/sourcegraph/cmd/symbols/types"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/rockskip"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
	connections "github.com/sourcegraph/sourcegraph/internal/database/connections/live"
	"github.com/sourcegraph/sourcegraph/internal/env"
	gitserver "github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func main() {
	if env.Get("USE_ROCKSKIP", "false", "use Rockskip instead of SQLite") == "true" {
		shared.Main(SetupRockskip)
	} else {
		shared.Main(shared.SetupSqlite)
	}
}

func SetupRockskip(observationContext *observation.Context) (types.SearchFunc, func(http.ResponseWriter, *http.Request), []goroutine.BackgroundRoutine, string, error) {
	baseConfig := env.BaseConfig{}
	config := LoadRockskipConfig(baseConfig)
	if err := baseConfig.Validate(); err != nil {
		log.Fatalf("Failed to load configuration: %s", err)
	}

	db := mustInitializeCodeIntelDB()
	gitserverClient := symbolsGitserver.NewClient(observationContext)
	repositoryFetcher := fetcher.NewRepositoryFetcher(gitserverClient, config.RepositoryFetcher.MaxTotalPathsLength, observationContext)
	git := NewGitserver(repositoryFetcher)
	createParser := func() rockskip.ParseSymbolsFunc { return createParserWithConfig(config.Ctags) }
	server, err := rockskip.NewServer(db, git, createParser, config.MaxConcurrentlyIndexing, config.MaxRepos, config.LogQueries, config.IndexRequestsQueueSize, config.SymbolsCacheSize, config.PathSymbolsCacheSize)
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
		SymbolsCacheSize:        baseConfig.GetInt("SYMBOLS_CACHE_SIZE", "1000000", "how many tuples of (path, symbol name, int ID) to cache in memory"),
		PathSymbolsCacheSize:    baseConfig.GetInt("PATH_SYMBOLS_CACHE_SIZE", "100000", "how many sets of symbols for files to cache in memory"),
	}
}

func createParserWithConfig(config types.CtagsConfig) rockskip.ParseSymbolsFunc {
	parser := mustCreateCtagsParser(config)

	return func(path string, bytes []byte) (symbols []rockskip.Symbol, err error) {
		entries, err := parser.Parse(path, bytes)
		if err != nil {
			return nil, err
		}

		symbols = []rockskip.Symbol{}
		for _, entry := range entries {
			symbols = append(symbols, rockskip.Symbol{
				Name:   entry.Name,
				Parent: entry.Parent,
				Kind:   entry.Kind,
				Line:   entry.Line,
			})
		}

		return symbols, nil
	}
}

func mustCreateCtagsParser(ctagsConfig types.CtagsConfig) ctags.Parser {
	options := ctags.Options{
		Bin:                ctagsConfig.Command,
		PatternLengthLimit: ctagsConfig.PatternLengthLimit,
	}
	if ctagsConfig.LogErrors {
		options.Info = log.New(os.Stderr, "ctags: ", log.LstdFlags)
	}
	if ctagsConfig.DebugLogs {
		options.Debug = log.New(os.Stderr, "DBUG ctags: ", log.LstdFlags)
	}

	parser, err := ctags.New(options)
	if err != nil {
		log.Fatalf("Failed to create new ctags parser: %s", err)
	}

	return symbolsParser.NewFilteringParser(parser, ctagsConfig.MaxFileSize, ctagsConfig.MaxSymbols)
}

func mustInitializeCodeIntelDB() *sql.DB {
	dsn := conf.GetServiceConnectionValueAndRestartOnChange(func(serviceConnections conftypes.ServiceConnections) string {
		return serviceConnections.CodeIntelPostgresDSN
	})
	var (
		db  *sql.DB
		err error
	)
	db, err = connections.EnsureNewCodeIntelDB(dsn, "symbols", &observation.TestContext)
	if err != nil {
		log.Fatalf("Failed to connect to codeintel database: %s", err)
	}

	return db
}

type Gitserver struct {
	repositoryFetcher fetcher.RepositoryFetcher
}

func NewGitserver(repositoryFetcher fetcher.RepositoryFetcher) Gitserver {
	return Gitserver{repositoryFetcher: repositoryFetcher}
}

func (g Gitserver) LogReverseEach(repo string, commit string, n int, onLogEntry func(entry rockskip.LogEntry) error) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	command := gitserver.DefaultClient.Command("git", rockskip.LogReverseArgs(n, commit)...)
	command.Repo = api.RepoName(repo)
	// We run a single `git log` command and stream the output while the repo is being processed, which
	// can take much longer than 1 minute (the default timeout).
	command.DisableTimeout()
	stdout, err := gitserver.StdoutReader(ctx, command)
	if err != nil {
		return err
	}
	defer stdout.Close()

	return errors.Wrap(rockskip.ParseLogReverseEach(stdout, onLogEntry), "ParseLogReverseEach")
}

func (g Gitserver) RevListEach(repo string, commit string, onCommit func(commit string) (shouldContinue bool, err error)) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	command := gitserver.DefaultClient.Command("git", rockskip.RevListArgs(commit)...)
	command.Repo = api.RepoName(repo)
	command.DisableTimeout()
	stdout, err := gitserver.StdoutReader(ctx, command)
	if err != nil {
		return err
	}
	defer stdout.Close()

	return rockskip.RevListEach(stdout, onCommit)
}

func (g Gitserver) ArchiveEach(repo string, commit string, paths []string, onFile func(path string, contents []byte) error) error {
	if len(paths) == 0 {
		return nil
	}

	args := types.SearchArgs{Repo: api.RepoName(repo), CommitID: api.CommitID(commit)}
	parseRequestOrErrors := g.repositoryFetcher.FetchRepositoryArchive(context.TODO(), args, paths)
	defer func() {
		// Ensure the channel is drained
		for range parseRequestOrErrors {
		}
	}()

	for parseRequestOrError := range parseRequestOrErrors {
		if parseRequestOrError.Err != nil {
			return errors.Wrap(parseRequestOrError.Err, "FetchRepositoryArchive")
		}

		err := onFile(parseRequestOrError.ParseRequest.Path, parseRequestOrError.ParseRequest.Data)
		if err != nil {
			return err
		}
	}

	return nil
}
