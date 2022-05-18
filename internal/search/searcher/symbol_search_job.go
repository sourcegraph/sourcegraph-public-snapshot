package searcher

import (
	"context"
	"sort"

	"github.com/neelance/parallel"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/opentracing/opentracing-go/log"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/search"
	"github.com/sourcegraph/sourcegraph/internal/search/job"
	"github.com/sourcegraph/sourcegraph/internal/search/result"
	"github.com/sourcegraph/sourcegraph/internal/search/streaming"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/internal/trace/ot"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

type SymbolSearcherJob struct {
	PatternInfo *search.TextPatternInfo
	Repos       []*search.RepositoryRevisions // the set of repositories to search with searcher.
	Limit       int
}

// Run calls the searcher service to search symbols.
func (s *SymbolSearcherJob) Run(ctx context.Context, clients job.RuntimeClients, stream streaming.Sender) (alert *search.Alert, err error) {
	tr, ctx, stream, finish := job.StartSpan(ctx, stream, s)
	defer func() { finish(alert, err) }()
	tr.TagFields(trace.LazyFields(s.Tags))

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	run := parallel.NewRun(conf.SearchSymbolsParallelism())

	for _, repoRevs := range s.Repos {
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

			matches, err := searchInRepo(ctx, clients.DB, repoRevs, s.PatternInfo, s.Limit)
			status, limitHit, err := search.HandleRepoSearchResult(repoRevs, len(matches) > s.Limit, false, err)
			stream.Send(streaming.SearchEvent{
				Results: matches,
				Stats: streaming.Stats{
					Status:     status,
					IsLimitHit: limitHit,
				},
			})
			if err != nil {
				tr.LogFields(log.String("repo", string(repoRevs.Repo.Name)), log.Error(err))
				// Only record error if we haven't timed out.
				if ctx.Err() == nil {
					cancel()
					run.Error(err)
				}
			}
		})
	}

	return nil, run.Wait()
}

func (s *SymbolSearcherJob) Name() string {
	return "SymbolSearcherJob"
}

func (s *SymbolSearcherJob) Tags() []log.Field {
	return []log.Field{
		trace.Stringer("patternInfo", s.PatternInfo),
		log.Int("numRepos", len(s.Repos)),
		log.Int("limit", s.Limit),
	}
}

func searchInRepo(ctx context.Context, db database.DB, repoRevs *search.RepositoryRevisions, patternInfo *search.TextPatternInfo, limit int) (res []result.Match, err error) {
	span, ctx := ot.StartSpanFromContext(ctx, "Search symbols in repo")
	defer func() {
		if err != nil {
			ext.Error.Set(span, true)
			span.LogFields(log.Error(err))
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
	commitID, err := gitserver.NewClient(db).ResolveRevision(ctx, repoRevs.GitserverRepo(), inputRev, gitserver.ResolveRevisionOptions{})
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
