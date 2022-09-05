package api

import (
	"context"
	"strings"
	"time"

	"github.com/opentracing/opentracing-go/log"

	"github.com/sourcegraph/sourcegraph/lib/errors"

	"github.com/dustin/go-humanize"

	"github.com/sourcegraph/sourcegraph/cmd/symbols/internal/api/observability"
	"github.com/sourcegraph/sourcegraph/cmd/symbols/internal/database/store"
	"github.com/sourcegraph/sourcegraph/cmd/symbols/internal/database/writer"
	sharedobservability "github.com/sourcegraph/sourcegraph/cmd/symbols/observability"
	"github.com/sourcegraph/sourcegraph/cmd/symbols/types"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/search"
	"github.com/sourcegraph/sourcegraph/internal/search/result"
)

const searchTimeout = 60 * time.Second

func MakeSqliteSearchFunc(operations *sharedobservability.Operations, cachedDatabaseWriter writer.CachedDatabaseWriter, db database.DB) types.SearchFunc {
	return func(ctx context.Context, args search.SymbolsParameters) (results []result.Symbol, err error) {
		ctx, trace, endObservation := operations.Search.With(ctx, &err, observation.Args{LogFields: []log.Field{
			log.String("repo", string(args.Repo)),
			log.String("commitID", string(args.CommitID)),
			log.String("query", args.Query),
			log.Bool("isRegExp", args.IsRegExp),
			log.Bool("isCaseSensitive", args.IsCaseSensitive),
			log.Int("numIncludePatterns", len(args.IncludePatterns)),
			log.String("includePatterns", strings.Join(args.IncludePatterns, ":")),
			log.String("excludePattern", args.ExcludePattern),
			log.Int("first", args.First),
			log.Int("timeout", args.Timeout),
		}})
		defer func() {
			endObservation(1, observation.Args{
				MetricLabelValues: []string{observability.GetParseAmount(ctx)},
				LogFields:         []log.Field{log.String("parseAmount", observability.GetParseAmount(ctx))},
			})
		}()
		ctx = observability.SeedParseAmount(ctx)

		timeout := searchTimeout
		if args.Timeout > 0 && time.Duration(args.Timeout)*time.Second < timeout {
			timeout = time.Duration(args.Timeout) * time.Second
		}
		ctx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()
		defer func() {
			if ctx.Err() == nil || !errors.Is(ctx.Err(), context.DeadlineExceeded) {
				return
			}

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			info, err2 := db.GitserverRepos().GetByName(ctx, args.Repo)
			if err2 != nil {
				err = errors.New("Processing symbols using the SQLite backend is taking a while. If this repository is ~1GB+, enable [Rockskip](https://docs.sourcegraph.com/code_navigation/explanations/rockskip).")
				return
			}
			size := info.RepoSizeBytes

			help := ""
			if size > 1_000_000_000 {
				help = "Enable [Rockskip](https://docs.sourcegraph.com/code_navigation/explanations/rockskip)."
			} else if size > 100_000_000 {
				help = "If this persists, enable [Rockskip](https://docs.sourcegraph.com/code_navigation/explanations/rockskip)."
			} else {
				help = "If this persists, make sure the symbols service has an SSD, a few GHz of CPU, and a few GB of RAM."
			}

			err = errors.Newf("Processing symbols using the SQLite backend is taking a while on this %s repository. %s", humanize.Bytes(uint64(size)), help)
			return
		}()

		dbFile, err := cachedDatabaseWriter.GetOrCreateDatabaseFile(ctx, args)
		if err != nil {
			return nil, errors.Wrap(err, "databaseWriter.GetOrCreateDatabaseFile")
		}
		trace.Log(log.String("dbFile", dbFile))

		var res result.Symbols
		err = store.WithSQLiteStore(dbFile, func(db store.Store) (err error) {
			if res, err = db.Search(ctx, args); err != nil {
				return errors.Wrap(err, "store.Search")
			}

			return nil
		})

		return res, err
	}
}
