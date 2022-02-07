package api

import (
	"context"
	"strings"
	"time"

	"github.com/opentracing/opentracing-go/log"

	"github.com/sourcegraph/sourcegraph/cmd/symbols/internal/api/observability"
	"github.com/sourcegraph/sourcegraph/cmd/symbols/internal/database/store"
	"github.com/sourcegraph/sourcegraph/cmd/symbols/internal/types"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/search/result"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

const searchTimeout = 60 * time.Second

func (h *apiHandler) handleSearchInternal(ctx context.Context, args types.SearchArgs) (_ *result.Symbols, err error) {
	ctx, trace, endObservation := h.operations.search.WithAndLogger(ctx, &err, observation.Args{LogFields: []log.Field{
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

	dbFile, err := h.cachedDatabaseWriter.GetOrCreateDatabaseFile(ctx, args)
	if err != nil {
		return nil, errors.Wrap(err, "databaseWriter.GetOrCreateDatabaseFile")
	}
	trace.Log(log.String("dbFile", dbFile))

	var results result.Symbols
	err = store.WithSQLiteStore(dbFile, func(db store.Store) (err error) {
		if results, err = db.Search(ctx, args); err != nil {
			return errors.Wrap(err, "store.Search")
		}

		return nil
	})

	return &results, err
}
