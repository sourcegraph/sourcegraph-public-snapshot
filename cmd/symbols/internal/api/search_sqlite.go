package api

import (
	"context"
	"strings"
	"time"

	"github.com/opentracing/opentracing-go/log"

	"github.com/sourcegraph/sourcegraph/lib/errors"

	"github.com/sourcegraph/sourcegraph/cmd/symbols/internal/api/observability"
	"github.com/sourcegraph/sourcegraph/cmd/symbols/internal/database/store"
	"github.com/sourcegraph/sourcegraph/cmd/symbols/internal/database/writer"
	sharedobservability "github.com/sourcegraph/sourcegraph/cmd/symbols/observability"
	"github.com/sourcegraph/sourcegraph/cmd/symbols/types"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/search/result"
)

const searchTimeout = 60 * time.Second

func MakeSqliteSearchFunc(operations *sharedobservability.Operations, cachedDatabaseWriter writer.CachedDatabaseWriter) types.SearchFunc {
	return func(ctx context.Context, args types.SearchArgs) (results []result.Symbol, err error) {
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
		}})
		defer func() {
			endObservation(1, observation.Args{
				MetricLabelValues: []string{observability.GetParseAmount(ctx)},
				LogFields:         []log.Field{log.String("parseAmount", observability.GetParseAmount(ctx))},
			})
		}()
		ctx = observability.SeedParseAmount(ctx)

		ctx, cancel := context.WithTimeout(ctx, searchTimeout)
		defer cancel()

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
