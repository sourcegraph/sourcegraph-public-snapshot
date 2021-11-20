package search

import (
	"context"
	"fmt"
	"regexp/syntax"
	"strings"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/diskcache"

	"github.com/inconshreveable/log15"
	"github.com/jmoiron/sqlx"
	"github.com/keegancsmith/sqlf"
	"github.com/opentracing/opentracing-go/ext"
	otlog "github.com/opentracing/opentracing-go/log"
	nettrace "golang.org/x/net/trace"

	"github.com/sourcegraph/sourcegraph/cmd/symbols/internal/parser"
	"github.com/sourcegraph/sourcegraph/cmd/symbols/internal/sqlite"
	"github.com/sourcegraph/sourcegraph/cmd/symbols/internal/types"
	"github.com/sourcegraph/sourcegraph/internal/trace/ot"

	"github.com/sourcegraph/sourcegraph/internal/search/result"
)

type Searcher interface {
	Search(ctx context.Context, args types.SearchArgs) (*result.Symbols, error)
}

type searcher struct {
	gitserverClient sqlite.GitserverClient
	parser          parser.Parser
	cache           *diskcache.Store
	databaseWriter  sqlite.DatabaseWriter
}

func NewSearcher(
	gitserverClient sqlite.GitserverClient,
	parser parser.Parser,
	cache *diskcache.Store,
	databaseWriter sqlite.DatabaseWriter,
) Searcher {
	return &searcher{
		gitserverClient: gitserverClient,
		parser:          parser,
		cache:           cache,
		databaseWriter:  databaseWriter,
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

	dbFile, err := getDBFile(ctx, s.gitserverClient, s.parser, s.cache, args, s.databaseWriter)
	if err != nil {
		return nil, err
	}
	db, err := sqlx.Open("sqlite3_with_regexp", dbFile)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	result, err := filterSymbols(ctx, db, args)
	if err != nil {
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
func getDBFile(ctx context.Context, gitserverClient sqlite.GitserverClient, parser parser.Parser, cache *diskcache.Store, args types.SearchArgs, databaseWriter sqlite.DatabaseWriter) (string, error) {
	diskcacheFile, err := cache.OpenWithPath(ctx, []string{string(args.Repo), fmt.Sprintf("%s-%d", args.CommitID, symbolsDBVersion)}, func(fetcherCtx context.Context, tempDBFile string) error {
		return databaseWriter.WriteDBFile(fetcherCtx, args, tempDBFile)
	})
	if err != nil {
		return "", err
	}
	defer diskcacheFile.File.Close()

	return diskcacheFile.File.Name(), err
}

func filterSymbols(ctx context.Context, db *sqlx.DB, args types.SearchArgs) (res []result.Symbol, err error) {
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

	makeCondition := func(column string, regex string) []*sqlf.Query {
		conditions := []*sqlf.Query{}

		if regex == "" {
			return conditions
		}

		if isExact, symbolName, err := isLiteralEquality(regex); isExact && err == nil {
			// It looks like the user is asking for exact matches, so use `=` to
			// get the speed boost from the index on the column.
			if args.IsCaseSensitive {
				conditions = append(conditions, sqlf.Sprintf(column+" = %s", symbolName))
			} else {
				conditions = append(conditions, sqlf.Sprintf(column+"lowercase = %s", strings.ToLower(symbolName)))
			}
		} else {
			if !args.IsCaseSensitive {
				regex = "(?i:" + regex + ")"
			}
			conditions = append(conditions, sqlf.Sprintf(column+" REGEXP %s", regex))
		}

		return conditions
	}

	negateAll := func(oldConditions []*sqlf.Query) []*sqlf.Query {
		newConditions := []*sqlf.Query{}

		for _, oldCondition := range oldConditions {
			newConditions = append(newConditions, sqlf.Sprintf("NOT %s", oldCondition))
		}

		return newConditions
	}

	var conditions []*sqlf.Query
	conditions = append(conditions, makeCondition("name", args.Query)...)
	for _, includePattern := range args.IncludePatterns {
		conditions = append(conditions, makeCondition("path", includePattern)...)
	}
	conditions = append(conditions, negateAll(makeCondition("path", args.ExcludePattern))...)

	var sqlQuery *sqlf.Query
	if len(conditions) == 0 {
		sqlQuery = sqlf.Sprintf("SELECT * FROM symbols LIMIT %s", args.First)
	} else {
		sqlQuery = sqlf.Sprintf("SELECT * FROM symbols WHERE %s LIMIT %s", sqlf.Join(conditions, "AND"), args.First)
	}

	var symbolsInDB []types.SymbolInDB
	err = db.Select(&symbolsInDB, sqlQuery.Query(sqlf.PostgresBindVar), sqlQuery.Args()...)
	if err != nil {
		return nil, err
	}

	for _, symbolInDB := range symbolsInDB {
		res = append(res, types.SymbolInDBToSymbol(symbolInDB))
	}

	span.SetTag("hits", len(res))
	return res, nil
}

// isLiteralEquality checks if the given regex matches literal strings exactly.
// Returns whether or not the regex is exact, along with the literal string if
// so.
func isLiteralEquality(expr string) (ok bool, lit string, err error) {
	r, err := syntax.Parse(expr, syntax.Perl)
	if err != nil {
		return false, "", err
	}
	// Want a Concat of size 3 which is [Begin, Literal, End]
	if r.Op != syntax.OpConcat || len(r.Sub) != 3 || // size 3 concat
		!(r.Sub[0].Op == syntax.OpBeginLine || r.Sub[0].Op == syntax.OpBeginText) || // Starts with ^
		!(r.Sub[2].Op == syntax.OpEndLine || r.Sub[2].Op == syntax.OpEndText) || // Ends with $
		r.Sub[1].Op != syntax.OpLiteral { // is a literal
		return false, "", nil
	}
	return true, string(r.Sub[1].Rune), nil
}
