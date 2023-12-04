package searcher

import (
	"context"
	"sort"

	"github.com/sourcegraph/conc/pool"
	"go.opentelemetry.io/otel/attribute"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/search"
	"github.com/sourcegraph/sourcegraph/internal/search/job"
	"github.com/sourcegraph/sourcegraph/internal/search/result"
	"github.com/sourcegraph/sourcegraph/internal/search/streaming"
	"github.com/sourcegraph/sourcegraph/internal/symbols"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

type SymbolSearchJob struct {
	PatternInfo *search.TextPatternInfo
	Repos       []*search.RepositoryRevisions // the set of repositories to search with searcher.
	Limit       int
}

// Run calls the searcher service to search symbols.
func (s *SymbolSearchJob) Run(ctx context.Context, clients job.RuntimeClients, stream streaming.Sender) (alert *search.Alert, err error) {
	tr, ctx, stream, finish := job.StartSpan(ctx, stream, s)
	defer func() { finish(alert, err) }()

	p := pool.New().
		WithContext(ctx).
		WithCancelOnError().
		WithFirstError().
		WithMaxGoroutines(conf.SearchSymbolsParallelism())

	for _, repoRevs := range s.Repos {
		repoRevs := repoRevs
		if ctx.Err() != nil {
			break
		}
		if len(repoRevs.Revs) == 0 {
			continue
		}

		p.Go(func(ctx context.Context) error {
			matches, err := searchInRepo(ctx, clients.Gitserver, repoRevs, s.PatternInfo, s.Limit)
			status, limitHit, err := search.HandleRepoSearchResult(repoRevs.Repo.ID, repoRevs.Revs, len(matches) > s.Limit, false, err)
			stream.Send(streaming.SearchEvent{
				Results: matches,
				Stats: streaming.Stats{
					Status:     status,
					IsLimitHit: limitHit,
				},
			})
			if err != nil {
				tr.SetAttributes(repoRevs.Repo.Name.Attr(), trace.Error(err))
			}
			return err
		})
	}

	return nil, p.Wait()
}

func (s *SymbolSearchJob) Name() string {
	return "SearcherSymbolSearchJob"
}

func (s *SymbolSearchJob) Attributes(v job.Verbosity) (res []attribute.KeyValue) {
	switch v {
	case job.VerbosityMax:
		fallthrough
	case job.VerbosityBasic:
		res = append(res, trace.Scoped("patternInfo", s.PatternInfo.Fields()...)...)
		res = append(res,
			attribute.Int("numRepos", len(s.Repos)),
			attribute.Int("limit", s.Limit),
		)
	}
	return res
}

func (s *SymbolSearchJob) Children() []job.Describer       { return nil }
func (s *SymbolSearchJob) MapChildren(job.MapFunc) job.Job { return s }

func searchInRepo(ctx context.Context, gitserverClient gitserver.Client, repoRevs *search.RepositoryRevisions, patternInfo *search.TextPatternInfo, limit int) (res []result.Match, err error) {
	inputRev := repoRevs.Revs[0]
	tr, ctx := trace.New(ctx, "symbols.searchInRepo",
		repoRevs.Repo.Name.Attr(),
		attribute.String("rev", inputRev))
	defer tr.EndWithErr(&err)

	// Do not trigger a repo-updater lookup (e.g.,
	// backend.{GitRepo,Repos.ResolveRev}) because that would slow this operation
	// down by a lot (if we're looping over many repos). This means that it'll fail if a
	// repo is not on gitserver.
	commitID, err := gitserverClient.ResolveRevision(ctx, repoRevs.GitserverRepo(), inputRev, gitserver.ResolveRevisionOptions{})
	if err != nil {
		return nil, err
	}
	tr.SetAttributes(commitID.Attr())

	symbols, err := symbols.DefaultClient.Search(ctx, search.SymbolsParameters{
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
	if err != nil {
		return nil, err
	}

	for i := range symbols {
		symbols[i].Line += 1 // callers expect 1-indexed lines
	}

	// All symbols are from the same repo, so we can just partition them by path
	// to build file matches
	return symbolsToMatches(symbols, repoRevs.Repo, commitID, inputRev), err
}

func symbolsToMatches(symbols []result.Symbol, repo types.MinimalRepo, commitID api.CommitID, inputRev string) result.Matches {
	symbolsByPath := make(map[string][]result.Symbol)
	for _, symbol := range symbols {
		cur := symbolsByPath[symbol.Path]
		symbolsByPath[symbol.Path] = append(cur, symbol)
	}

	// Create file matches from partitioned symbols
	matches := make(result.Matches, 0, len(symbolsByPath))
	for path, symbols := range symbolsByPath {
		file := result.File{
			Path:     path,
			Repo:     repo,
			CommitID: commitID,
			InputRev: &inputRev,
		}

		symbolMatches := make([]*result.SymbolMatch, 0, len(symbols))
		for _, symbol := range symbols {
			symbolMatches = append(symbolMatches, &result.SymbolMatch{
				File:   &file,
				Symbol: symbol,
			})
		}

		matches = append(matches, &result.FileMatch{
			Symbols: symbolMatches,
			File:    file,
		})
	}

	// Make the results deterministic
	sort.Sort(matches)
	return matches
}
