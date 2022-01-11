package api

import (
	"context"
	"database/sql"
	"log"
	"os"
	"strings"
	"sync"

	otlog "github.com/opentracing/opentracing-go/log"

	"github.com/sourcegraph/go-ctags"

	"github.com/sourcegraph/sourcegraph/cmd/symbols/internal/types"
	rockskip "github.com/sourcegraph/sourcegraph/enterprise/cmd/rockskip"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
	connections "github.com/sourcegraph/sourcegraph/internal/database/connections/live"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/search/result"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func MakeRockskipSearchFunc(operations *operations, ctagsConfig types.CtagsConfig) types.SearchFunc {
	var repoToMutexMutex = sync.Mutex{}
	var repoToMutex = map[string]*sync.Mutex{}

	parser := mustCreateCtagsParser(ctagsConfig)

	return func(ctx context.Context, args types.SearchArgs) (results *[]result.Symbol, err error) {
		_, _, endObservation := operations.search.WithAndLogger(ctx, &err, observation.Args{LogFields: []otlog.Field{
			otlog.String("repo", string(args.Repo)),
			otlog.String("commitID", string(args.CommitID)),
			otlog.String("query", args.Query),
			otlog.Bool("isRegExp", args.IsRegExp),
			otlog.Bool("isCaseSensitive", args.IsCaseSensitive),
			otlog.Int("numIncludePatterns", len(args.IncludePatterns)),
			otlog.String("includePatterns", strings.Join(args.IncludePatterns, ":")),
			otlog.String("excludePattern", args.ExcludePattern),
			otlog.Int("first", args.First),
		}})
		defer func() {
			endObservation(1, observation.Args{})
		}()

		repoToMutexMutex.Lock()
		mutex, ok := repoToMutex[string(args.Repo)]
		if !ok {
			mutex = &sync.Mutex{}
			repoToMutex[string(args.Repo)] = mutex
		}
		mutex.Lock()
		repoToMutexMutex.Unlock()
		defer mutex.Unlock()

		git, err := NewGitserver()
		if err != nil {
			return nil, errors.Wrap(err, "rockskip.NewPostgresDB")
		}

		db, err := rockskip.NewPostgresDB(mustInitializeCodeIntelDB())
		if err != nil {
			return nil, errors.Wrap(err, "rockskip.NewPostgresDB")
		}

		var parse rockskip.ParseSymbolsFunc = func(path string, bytes []byte) (symbols []string, err error) {
			entries, err := parser.Parse(path, bytes)
			if err != nil {
				return nil, err
			}

			symbols = []string{}
			for _, entry := range entries {
				symbols = append(symbols, entry.Name)
			}

			return symbols, nil
		}

		err = rockskip.Index(git, db, parse, string(args.CommitID))
		if err != nil {
			return nil, errors.Wrap(err, "rockskip.Index")
		}

		blobs, err := rockskip.Search(db, string(args.CommitID))
		if err != nil {
			return nil, errors.Wrap(err, "rockskip.Search")
		}

		res := []result.Symbol{}
		for _, blob := range blobs {
			for _, symbol := range blob.Symbols {
				res = append(res, result.Symbol{
					Name:        symbol,
					Path:        blob.Path,
					Line:        0,
					Kind:        "",
					Language:    "",
					Parent:      "",
					ParentKind:  "",
					Signature:   "",
					Pattern:     "",
					FileLimited: false,
				})
			}
		}

		return &res, err
	}
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

func NewGitserver() (rockskip.Git, error) {
	return nil, nil
}
