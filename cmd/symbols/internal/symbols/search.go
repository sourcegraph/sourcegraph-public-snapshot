package symbols

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/sourcegraph/sourcegraph/pkg/pathmatch"

	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	otlog "github.com/opentracing/opentracing-go/log"
	"github.com/sourcegraph/sourcegraph/pkg/symbols/protocol"
	"golang.org/x/net/trace"
	log15 "gopkg.in/inconshreveable/log15.v2"
)

// maxFileSize is the limit on file size in bytes. Only files smaller than this are processed.
const maxFileSize = 1 << 19 // 512KB

func (s *Service) handleSearch(w http.ResponseWriter, r *http.Request) {
	var args protocol.SearchArgs
	if err := json.NewDecoder(r.Body).Decode(&args); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	result, err := s.search(r.Context(), args)
	if err != nil {
		if err == context.Canceled && r.Context().Err() == context.Canceled {
			return // client went away
		}
		log15.Error("Symbol search failed", "args", args, "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(result); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (s *Service) search(ctx context.Context, args protocol.SearchArgs) (result *protocol.SearchResult, err error) {
	ctx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	log15.Debug("Symbol search", "repo", args.Repo, "query", args.Query)

	span, ctx := opentracing.StartSpanFromContext(ctx, "search")
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

	tr := trace.New("symbols.search", fmt.Sprintf("args:%+v", args))
	defer func() {
		if err != nil {
			tr.LazyPrintf("error: %v", err)
			tr.SetError()
		}
		tr.Finish()
	}()

	symbols, err := s.indexedSymbols(ctx, args.Repo, args.CommitID)
	if err != nil {
		return nil, err
	}

	const maxFirst = 500
	if args.First < 0 || args.First > maxFirst {
		args.First = maxFirst
	}

	result = &protocol.SearchResult{}
	if args.Query == "" && len(args.IncludePatterns) == 0 && args.ExcludePattern == "" {
		// No filters were provided, save iterating the symbols and return a slice
		if args.First != 0 && len(symbols) > args.First {
			symbols = symbols[:args.First]
		}
		result.Symbols = symbols
	} else {
		res, err := filterSymbols(ctx, symbols, args)
		if err != nil {
			return nil, err
		}
		result.Symbols = res
	}
	return result, nil
}

func filterSymbols(ctx context.Context, symbols []protocol.Symbol, args protocol.SearchArgs) (res []protocol.Symbol, err error) {
	span, _ := opentracing.StartSpanFromContext(ctx, "filterSymbols")
	defer func() {
		if err != nil {
			ext.Error.Set(span, true)
			span.LogFields(otlog.Error(err))
		}
		span.Finish()
	}()
	span.SetTag("before", len(symbols))

	query := args.Query
	if !args.IsRegExp {
		query = regexp.QuoteMeta(query)
	}
	if !args.IsCaseSensitive {
		query = "(?i:" + query + ")"
	}
	queryRegex, err := regexp.Compile(query)
	if err != nil {
		return nil, err
	}

	fileFilter, err := pathmatch.CompilePathPatterns(args.IncludePatterns, args.ExcludePattern, pathmatch.CompileOptions{
		CaseSensitive: args.IsCaseSensitive,
		RegExp:        args.IsRegExp,
	})
	if err != nil {
		return nil, err
	}

	pathFilter := func(s protocol.Symbol) bool {
		return fileFilter.MatchPath(s.Path)
	}

	nameFilter := func(s protocol.Symbol) bool {
		return queryRegex.MatchString(s.Name)
	}

	var filters []func(s protocol.Symbol) bool

	if strings.HasPrefix(args.Query, "^") {
		filters = []func(s protocol.Symbol) bool{nameFilter, pathFilter}
	} else {
		filters = []func(s protocol.Symbol) bool{pathFilter, nameFilter}
	}

SYMBOL:
	for _, symbol := range symbols {
		for _, filter := range filters {
			if !filter(symbol) {
				continue SYMBOL
			}
		}
		res = append(res, symbol)
		if args.First > 0 && len(res) == args.First {
			break
		}
	}

	span.SetTag("after", len(res))
	return res, nil
}
