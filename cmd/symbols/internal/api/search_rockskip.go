package api

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"sync"

	"github.com/sourcegraph/go-ctags"

	"github.com/sourcegraph/sourcegraph/cmd/symbols/internal/types"
	rockskip "github.com/sourcegraph/sourcegraph/enterprise/cmd/rockskip"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
	connections "github.com/sourcegraph/sourcegraph/internal/database/connections/live"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/search/result"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func MakeRockskipSearchFunc(operations *operations, ctagsConfig types.CtagsConfig, maxRepos int) (types.SearchFunc, error) {
	var repoToMutexMutex = sync.Mutex{}
	var repoToMutex = map[string]*sync.Mutex{}

	parser := mustCreateCtagsParser(ctagsConfig)

	db := mustInitializeCodeIntelDB()

	return func(ctx context.Context, args types.SearchArgs) (results *[]result.Symbol, err error) {
		// _, _, endObservation := operations.search.WithAndLogger(ctx, &err, observation.Args{LogFields: []otlog.Field{
		// 	otlog.String("repo", string(args.Repo)),
		// 	otlog.String("commitID", string(args.CommitID)),
		// 	otlog.String("query", args.Query),
		// 	otlog.Bool("isRegExp", args.IsRegExp),
		// 	otlog.Bool("isCaseSensitive", args.IsCaseSensitive),
		// 	otlog.Int("numIncludePatterns", len(args.IncludePatterns)),
		// 	otlog.String("includePatterns", strings.Join(args.IncludePatterns, ":")),
		// 	otlog.String("excludePattern", args.ExcludePattern),
		// 	otlog.Int("first", args.First),
		// }})
		// defer func() {
		// 	endObservation(1, observation.Args{})
		// }()

		repoToMutexMutex.Lock()
		mutex, ok := repoToMutex[string(args.Repo)]
		if !ok {
			mutex = &sync.Mutex{}
			repoToMutex[string(args.Repo)] = mutex
		}
		mutex.Lock()
		repoToMutexMutex.Unlock()
		defer mutex.Unlock()

		var parse rockskip.ParseSymbolsFunc = func(path string, bytes []byte) (symbols []rockskip.Symbol, err error) {
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

		err = rockskip.Index(NewGitserver(string(args.Repo)), db, parse, string(args.Repo), string(args.CommitID), maxRepos)
		if err != nil {
			return nil, errors.Wrap(err, "rockskip.Index")
		}

		var query *string
		if args.Query != "" {
			query = &args.Query
		}
		blobs, err := rockskip.Search(db, string(args.Repo), string(args.CommitID), query)
		if err != nil {
			return nil, errors.Wrap(err, "rockskip.Search")
		}

		res := []result.Symbol{}
		for _, blob := range blobs {
			for _, symbol := range blob.Symbols {
				res = append(res, result.Symbol{
					Name:   symbol.Name,
					Path:   blob.Path,
					Line:   symbol.Line,
					Kind:   symbol.Kind,
					Parent: symbol.Parent,
				})
			}
		}

		return &res, err
	}, nil
}

func mustInitializeCodeIntelDB() *sql.DB {
	dsn := conf.GetServiceConnectionValueAndRestartOnChange(func(serviceConnections conftypes.ServiceConnections) string {
		return serviceConnections.CodeIntelPostgresDSN
	})
	var (
		db  *sql.DB
		err error
	)
	if os.Getenv("NEW_MIGRATIONS") == "" {
		// CURRENTLY DEPRECATING
		db, err = connections.NewCodeIntelDB(dsn, "symbols", true, &observation.TestContext)
	} else {
		db, err = connections.EnsureNewCodeIntelDB(dsn, "symbols", &observation.TestContext)
	}
	if err != nil {
		log.Fatalf("Failed to connect to codeintel database: %s", err)
	}

	return db
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

	return parser
}

type Gitserver struct {
	repo string
}

func NewGitserver(repo string) rockskip.Git {
	return Gitserver{
		repo: repo,
	}
}

func (g Gitserver) LogReverse(commit string, n int) ([]rockskip.LogEntry, error) {
	command := gitserver.DefaultClient.Command("git", rockskip.LogReverseArgs(n, commit)...)
	command.Repo = api.RepoName(g.repo)
	stdout, err := command.Output(context.Background())
	if err != nil {
		return nil, err
	}
	return rockskip.ParseLogReverse(bytes.NewReader(stdout))
}

func (g Gitserver) RevListEach(commit string, onCommit func(commit string) (shouldContinue bool, err error)) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	command := gitserver.DefaultClient.Command("git", rockskip.RevListArgs(commit)...)
	command.Repo = api.RepoName(g.repo)
	stdout, err := gitserver.StdoutReader(ctx, command)
	if err != nil {
		return err
	}
	defer stdout.Close()

	return rockskip.RevListEach(stdout, onCommit)
}

func (g Gitserver) CatFile(commit string, path string) ([]byte, error) {
	command := gitserver.DefaultClient.Command("git", "cat-file", "blob", fmt.Sprintf("%s:%s", commit, path))
	command.Repo = api.RepoName(g.repo)
	return command.Output(context.Background())
}
