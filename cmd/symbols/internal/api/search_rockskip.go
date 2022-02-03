package api

import (
	"archive/tar"
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/sourcegraph/go-ctags"

	"github.com/sourcegraph/sourcegraph/cmd/symbols/internal/fetcher"
	symbolsGitserver "github.com/sourcegraph/sourcegraph/cmd/symbols/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/cmd/symbols/internal/types"
	rockskip "github.com/sourcegraph/sourcegraph/enterprise/cmd/rockskip"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
	connections "github.com/sourcegraph/sourcegraph/internal/database/connections/live"
	gitserver "github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/search/result"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func MakeRockskipSearchFunc(observationContext *observation.Context, ctagsConfig types.CtagsConfig, maxRepos int) (types.SearchFunc, error) {
	parser := mustCreateCtagsParser(ctagsConfig)

	operations := NewOperations(observationContext)
	// TODO use operations
	_ = operations

	gitserverClient := symbolsGitserver.NewClient(observationContext)

	shouldRead := func(tarHeader *tar.Header) bool { return true }
	f := fetcher.NewRepositoryFetcher(gitserverClient, 16, observationContext, shouldRead)

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

		fmt.Println(".")
		fmt.Println("🔵 Rockskip search", args.Repo, args.CommitID, args.Query)
		defer func() {
			if results == nil {
				fmt.Println("🔴 Rockskip search failed")
			} else {
				for _, result := range *results {
					fmt.Println("  -", result.Path+":"+fmt.Sprint(result.Line), result.Name)
				}
				fmt.Println("🔴 Rockskip search", len(*results))
				fmt.Println(".")
			}
		}()

		tasklog := rockskip.NewTaskLog()
		ctx, cancel := context.WithCancel(context.Background())
		go func() {
			for {
				select {
				case <-ctx.Done():
				case <-time.After(1 * time.Second):
				}

				tasklog.Print()

				if ctx.Err() != nil {
					break
				}
			}
		}()

		err = rockskip.Index(NewGitserver(f, string(args.Repo)), db, tasklog, parse, string(args.Repo), string(args.CommitID), maxRepos)
		cancel()
		if err != nil {
			return nil, errors.Wrap(err, "rockskip.Index")
		}

		var query *string
		if args.Query != "" {
			query = &args.Query
		}
		tasklog2 := rockskip.NewTaskLog()
		blobs, err := rockskip.Search(db, tasklog2, string(args.Repo), string(args.CommitID), query)
		tasklog2.Print()
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
	f    fetcher.RepositoryFetcher
	repo string
}

func NewGitserver(f fetcher.RepositoryFetcher, repo string) rockskip.Git {
	return Gitserver{
		f:    f,
		repo: repo,
	}
}

func (g Gitserver) LogReverseEach(commit string, n int, onLogEntry func(entry rockskip.LogEntry) error) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	command := gitserver.DefaultClient.Command("git", rockskip.LogReverseArgs(n, commit)...)
	command.Repo = api.RepoName(g.repo)
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

func (g Gitserver) ArchiveEach(commit string, paths []string, onFile func(path string, contents []byte) error) error {
	if len(paths) == 0 {
		return nil
	}

	args := types.SearchArgs{Repo: api.RepoName(g.repo), CommitID: api.CommitID(commit)}
	parseRequestOrErrors := g.f.FetchRepositoryArchive(context.TODO(), args, paths)
	defer func() {
		// Ensure the channel is drained
		for range parseRequestOrErrors {
		}
	}()

	for parseRequestOrError := range parseRequestOrErrors {
		if parseRequestOrError.Err != nil {
			return parseRequestOrError.Err
		}

		err := onFile(parseRequestOrError.ParseRequest.Path, parseRequestOrError.ParseRequest.Data)
		if err != nil {
			return err
		}
	}

	return nil
}
