package run

import (
	"context"
	"fmt"
	"regexp/syntax"
	"sort"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/google/zoekt"
	zoektquery "github.com/google/zoekt/query"
	"github.com/neelance/parallel"
	"github.com/opentracing/opentracing-go/ext"
	otlog "github.com/opentracing/opentracing-go/log"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/search"
	"github.com/sourcegraph/sourcegraph/internal/search/result"
	"github.com/sourcegraph/sourcegraph/internal/search/streaming"
	zoektutil "github.com/sourcegraph/sourcegraph/internal/search/zoekt"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/internal/trace/ot"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/internal/vcs/git"
)

var MockSearchSymbols func(ctx context.Context, args *search.TextParameters, limit int) (res []result.Match, stats *streaming.Stats, err error)

// SearchSymbols searches the given repos in parallel for symbols matching the given search query
// it can be used for both search suggestions and search results
//
// May return partial results and an error
func SearchSymbols(ctx context.Context, args *search.TextParameters, limit int, stream streaming.Sender) (err error) {
	if MockSearchSymbols != nil {
		results, stats, err := MockSearchSymbols(ctx, args, limit)
		stream.Send(streaming.SearchEvent{
			Results: results,
			Stats:   statsDeref(stats),
		})
		return err
	}

	repos, err := args.RepoPromise.Get(ctx)
	if err != nil {
		return err
	}

	tr, ctx := trace.New(ctx, "Search symbols", fmt.Sprintf("query: %+v, numRepoRevs: %d", args.PatternInfo, len(repos)))
	defer func() {
		tr.SetError(err)
		tr.Finish()
	}()

	if args.PatternInfo.Pattern == "" {
		return nil
	}

	ctx, stream, cancel := streaming.WithLimit(ctx, stream, limit)
	defer cancel()

	indexed, err := zoektutil.NewIndexedSearchRequest(ctx, args, zoektutil.SymbolRequest, stream)
	if err != nil {
		return err
	}

	run := parallel.NewRun(conf.SearchSymbolsParallelism())

	run.Acquire()
	goroutine.Go(func() {
		defer run.Release()

		err := indexed.Search(ctx, stream)
		if err != nil {
			tr.LogFields(otlog.Error(err))
			// Only record error if we haven't timed out.
			if ctx.Err() == nil {
				cancel()
				run.Error(err)
			}
		}
	})

	for _, repoRevs := range indexed.Unindexed {
		repoRevs := repoRevs
		if ctx.Err() != nil {
			break
		}
		if len(repoRevs.RevSpecs()) == 0 {
			continue
		}
		run.Acquire()
		goroutine.Go(func() {
			defer run.Release()

			matches, err := searchSymbolsInRepo(ctx, repoRevs, args.PatternInfo, limit)
			stats, err := handleRepoSearchResult(repoRevs, len(matches) > limit, false, err)
			stream.Send(streaming.SearchEvent{
				Results: matches,
				Stats:   stats,
			})
			if err != nil {
				tr.LogFields(otlog.String("repo", string(repoRevs.Repo.Name)), otlog.Error(err))
				// Only record error if we haven't timed out.
				if ctx.Err() == nil {
					cancel()
					run.Error(err)
				}
			}
		})
	}

	return run.Wait()
}

func searchSymbolsInRepo(ctx context.Context, repoRevs *search.RepositoryRevisions, patternInfo *search.TextPatternInfo, limit int) (res []result.Match, err error) {
	span, ctx := ot.StartSpanFromContext(ctx, "Search symbols in repo")
	defer func() {
		if err != nil {
			ext.Error.Set(span, true)
			span.LogFields(otlog.Error(err))
		}
		span.Finish()
	}()
	span.SetTag("repo", string(repoRevs.Repo.Name))

	inputRev := repoRevs.RevSpecs()[0]
	span.SetTag("rev", inputRev)
	// Do not trigger a repo-updater lookup (e.g.,
	// backend.{GitRepo,Repos.ResolveRev}) because that would slow this operation
	// down by a lot (if we're looping over many repos). This means that it'll fail if a
	// repo is not on gitserver.
	commitID, err := git.ResolveRevision(ctx, repoRevs.GitserverRepo(), inputRev, git.ResolveRevisionOptions{})
	if err != nil {
		return nil, err
	}
	span.SetTag("commit", string(commitID))

	symbols, err := backend.Symbols.ListTags(ctx, search.SymbolsParameters{
		Repo:            repoRevs.Repo.Name,
		CommitID:        commitID,
		Query:           patternInfo.Pattern,
		IsCaseSensitive: patternInfo.IsCaseSensitive,
		IsRegExp:        patternInfo.IsRegExp,
		IncludePatterns: patternInfo.IncludePatterns,
		ExcludePattern:  patternInfo.ExcludePattern,
		// Ask for limit + 1 so we can detect whether there are more results than the limit.
		First: limit + 1,
	})

	// All symbols are from the same repo, so we can just partition them by path
	// to build file matches
	symbolsByPath := make(map[string][]*result.Symbol)
	for _, symbol := range symbols {
		cur := symbolsByPath[symbol.Path]
		symbolsByPath[symbol.Path] = append(cur, &symbol)
	}

	// Create file matches from partitioned symbols
	matches := make([]result.Match, 0, len(symbolsByPath))
	for path, symbols := range symbolsByPath {
		file := result.File{
			Path:     path,
			Repo:     repoRevs.Repo,
			CommitID: commitID,
			InputRev: &inputRev,
		}

		symbolMatches := make([]*result.SymbolMatch, 0, len(symbols))
		for _, symbol := range symbols {
			symbolMatches = append(symbolMatches, &result.SymbolMatch{
				File:   &file,
				Symbol: *symbol,
			})
		}

		matches = append(matches, &result.FileMatch{
			Symbols: symbolMatches,
			File:    file,
		})
	}

	// Make the results deterministic
	sort.Sort(result.Matches(matches))
	return matches, err
}

// indexedSymbols checks to see if Zoekt has indexed symbols information for a
// repository at a specific commit. If it has it returns the branch name (for
// use when querying zoekt). Otherwise an empty string is returned.
func indexedSymbolsBranch(ctx context.Context, repository, commit string) string {
	z := search.Indexed()
	if !z.Enabled() {
		return ""
	}

	ctx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()
	set, err := z.ListAll(ctx)
	if err != nil {
		return ""
	}

	repo, ok := set[repository]
	if !ok || !repo.HasSymbols {
		return ""
	}

	for _, branch := range repo.Branches {
		if branch.Version == commit {
			return branch.Name
		}
	}

	return ""
}

func searchZoektSymbols(ctx context.Context, repoName types.RepoName, commitID api.CommitID, inputRev *string, branch string, queryString *string, first *int32, includePatterns *[]string) (res []*result.SymbolMatch, err error) {
	raw := *queryString
	if raw == "" {
		raw = ".*"
	}

	expr, err := syntax.Parse(raw, syntax.ClassNL|syntax.PerlX|syntax.UnicodeGroups)
	if err != nil {
		return
	}

	var query zoektquery.Q
	if expr.Op == syntax.OpLiteral {
		query = &zoektquery.Substring{
			Pattern: string(expr.Rune),
			Content: true,
		}
	} else {
		query = &zoektquery.Regexp{
			Regexp:  expr,
			Content: true,
		}
	}

	ands := []zoektquery.Q{
		&zoektquery.RepoBranches{Set: map[string][]string{
			string(repoName.Name): {branch},
		}},
		&zoektquery.Symbol{Expr: query},
	}
	for _, p := range *includePatterns {
		q, err := zoektutil.FileRe(p, true)
		if err != nil {
			return nil, err
		}
		ands = append(ands, q)
	}

	final := zoektquery.Simplify(zoektquery.NewAnd(ands...))
	match := limitOrDefault(first) + 1
	resp, err := search.Indexed().Client.Search(ctx, final, &zoekt.SearchOptions{
		Trace:                  ot.ShouldTrace(ctx),
		MaxWallTime:            3 * time.Second,
		ShardMaxMatchCount:     match * 25,
		TotalMaxMatchCount:     match * 25,
		ShardMaxImportantMatch: match * 25,
		TotalMaxImportantMatch: match * 25,
		MaxDocDisplayCount:     match,
	})
	if err != nil {
		return nil, err
	}

	for _, file := range resp.Files {
		newFile := &result.File{
			Repo:     repoName,
			CommitID: commitID,
			InputRev: inputRev,
			Path:     file.FileName,
		}

		for _, l := range file.LineMatches {
			if l.FileName {
				continue
			}

			for _, m := range l.LineFragments {
				if m.SymbolInfo == nil {
					continue
				}

				res = append(res, &result.SymbolMatch{
					Symbol: result.Symbol{
						Name:       m.SymbolInfo.Sym,
						Kind:       m.SymbolInfo.Kind,
						Parent:     m.SymbolInfo.Parent,
						ParentKind: m.SymbolInfo.ParentKind,
						Path:       file.FileName,
						Line:       l.LineNumber,
						Language:   file.Language,
					},
					File: newFile,
				})
			}
		}
	}
	return
}

func ComputeSymbols(ctx context.Context, repoName types.RepoName, commitID api.CommitID, inputRev *string, query *string, first *int32, includePatterns *[]string) (res []*result.SymbolMatch, err error) {
	// TODO(keegancsmith) we should be able to use indexedSearchRequest here
	// and remove indexedSymbolsBranch.
	if branch := indexedSymbolsBranch(ctx, string(repoName.Name), string(commitID)); branch != "" {
		return searchZoektSymbols(ctx, repoName, commitID, inputRev, branch, query, first, includePatterns)
	}

	ctx, done := context.WithTimeout(ctx, 5*time.Second)
	defer done()
	defer func() {
		if ctx.Err() != nil && len(res) == 0 {
			err = errors.New("processing symbols is taking longer than expected. Try again in a while")
		}
	}()
	var includePatternsSlice []string
	if includePatterns != nil {
		includePatternsSlice = *includePatterns
	}

	searchArgs := search.SymbolsParameters{
		CommitID:        commitID,
		First:           limitOrDefault(first) + 1, // add 1 so we can determine PageInfo.hasNextPage
		Repo:            repoName.Name,
		IncludePatterns: includePatternsSlice,
	}
	if query != nil {
		searchArgs.Query = *query
	}

	symbols, err := backend.Symbols.ListTags(ctx, searchArgs)
	if err != nil {
		return nil, err
	}

	fileWithPath := func(path string) *result.File {
		return &result.File{
			Path:     path,
			Repo:     repoName,
			InputRev: inputRev,
			CommitID: commitID,
		}
	}

	matches := make([]*result.SymbolMatch, 0, len(symbols))
	for _, symbol := range symbols {
		matches = append(matches, &result.SymbolMatch{
			Symbol: symbol,
			File:   fileWithPath(symbol.Path),
		})
	}
	return matches, err
}

const DefaultSymbolLimit = 100

func limitOrDefault(first *int32) int {
	if first == nil {
		return DefaultSymbolLimit
	}
	return int(*first)
}
