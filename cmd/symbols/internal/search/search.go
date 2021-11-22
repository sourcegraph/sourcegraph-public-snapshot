package search

import (
	"context"
	"fmt"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/diskcache"

	"github.com/inconshreveable/log15"
	"github.com/opentracing/opentracing-go/ext"
	otlog "github.com/opentracing/opentracing-go/log"
	nettrace "golang.org/x/net/trace"

	sqlite "github.com/sourcegraph/sourcegraph/cmd/symbols/internal/database"
	"github.com/sourcegraph/sourcegraph/cmd/symbols/internal/database/store"
	"github.com/sourcegraph/sourcegraph/cmd/symbols/internal/types"
	"github.com/sourcegraph/sourcegraph/internal/trace/ot"

	"github.com/sourcegraph/sourcegraph/internal/search/result"
)

type Searcher interface {
	Search(ctx context.Context, args types.SearchArgs) (*result.Symbols, error)
}

type searcher struct {
	cache          *diskcache.Store
	databaseWriter sqlite.DatabaseWriter
}

func NewSearcher(
	cache *diskcache.Store,
	databaseWriter sqlite.DatabaseWriter,
) Searcher {
	return &searcher{
		cache:          cache,
		databaseWriter: databaseWriter,
	}
}

func (s *searcher) Search(ctx context.Context, args types.SearchArgs) (*result.Symbols, error) {
	var err error
	ctx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	log15.Debug("Symbol search", "repo", args.Repo, "query", args.Query)
	span, ctx := ot.StartSpanFromContext(ctx, "search")
	span.SetTag("repo", args.Repo)
	span.SetTag("commitID", args.CommitID)
	span.SetTag("query", args.Query)
	span.SetTag("first", args.First)
	defer func() {
		if err != nil {
			ext.Error.Set(span, true)
			span.LogFields(otlog.Error(err))
		}
		span.Finish()
	}()

	tr := nettrace.New("symbols.search", fmt.Sprintf("args:%+v", args))
	defer func() {
		if err != nil {
			tr.LazyPrintf("error: %v", err)
			tr.SetError()
		}
		tr.Finish()
	}()

	dbFile, err := getDBFile(ctx, s.cache, args, s.databaseWriter)
	if err != nil {
		return nil, err
	}

	var result []result.Symbol
	if err := store.WithSQLiteStore(dbFile, func(db store.Store) error {
		result, err = filterSymbols(ctx, db, args)
		return err
	}); err != nil {
		return nil, err
	}

	return &result, nil
}

// The version of the symbols database schema. This is included in the database
// filenames to prevent a newer version of the symbols service from attempting
// to read from a database created by an older (and likely incompatible) symbols
// service. Increment this when you change the database schema.
const symbolsDBVersion = 4

// getDBFile returns the path to the sqlite3 database for the repo@commit
// specified in `args`. If the database doesn't already exist in the disk cache,
// it will create a new one and write all the symbols into it.
func getDBFile(ctx context.Context, cache *diskcache.Store, args types.SearchArgs, databaseWriter sqlite.DatabaseWriter) (string, error) {
	diskcacheFile, err := cache.OpenWithPath(ctx, []string{string(args.Repo), fmt.Sprintf("%s-%d", args.CommitID, symbolsDBVersion)}, func(fetcherCtx context.Context, tempDBFile string) error {
		return databaseWriter.WriteDBFile(fetcherCtx, args, tempDBFile)
	})
	if err != nil {
		return "", err
	}
	defer diskcacheFile.File.Close()

	return diskcacheFile.File.Name(), err
}

func filterSymbols(ctx context.Context, db store.Store, args types.SearchArgs) (res []result.Symbol, err error) {
	span, _ := ot.StartSpanFromContext(ctx, "filterSymbols")
	defer func() {
		if err != nil {
			ext.Error.Set(span, true)
			span.LogFields(otlog.Error(err))
		}
		span.Finish()
	}()

	const maxFirst = 500
	if args.First < 0 || args.First > maxFirst {
		args.First = maxFirst
	}

	res, err = db.Search(ctx, args)
	if err != nil {
		return nil, err
	}

	span.SetTag("hits", len(res))
	return res, nil
}
