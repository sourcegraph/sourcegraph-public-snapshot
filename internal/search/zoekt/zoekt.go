package zoekt

import (
	"context"
	"regexp/syntax"
	"time"

	"github.com/google/zoekt"
	zoektquery "github.com/google/zoekt/query"
	"github.com/inconshreveable/log15"
	"github.com/opentracing/opentracing-go"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/search"
	"github.com/sourcegraph/sourcegraph/internal/trace/ot"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

const defaultMaxSearchResults = 30

var defaultTimeout = 20 * time.Second

func FileRe(pattern string, queryIsCaseSensitive bool) (zoektquery.Q, error) {
	return ParseRe(pattern, true, false, queryIsCaseSensitive)
}

func noOpAnyChar(re *syntax.Regexp) {
	if re.Op == syntax.OpAnyChar {
		re.Op = syntax.OpAnyCharNotNL
	}
	for _, s := range re.Sub {
		noOpAnyChar(s)
	}
}

func ParseRe(pattern string, filenameOnly bool, contentOnly bool, queryIsCaseSensitive bool) (zoektquery.Q, error) {
	// these are the flags used by zoekt, which differ to searcher.
	re, err := syntax.Parse(pattern, syntax.ClassNL|syntax.PerlX|syntax.UnicodeGroups)
	if err != nil {
		return nil, err
	}
	noOpAnyChar(re)
	// zoekt decides to use its literal optimization at the query parser
	// level, so we check if our regex can just be a literal.
	if re.Op == syntax.OpLiteral {
		return &zoektquery.Substring{
			Pattern:       string(re.Rune),
			CaseSensitive: queryIsCaseSensitive,
			Content:       contentOnly,
			FileName:      filenameOnly,
		}, nil
	}
	return &zoektquery.Regexp{
		Regexp:        re,
		CaseSensitive: queryIsCaseSensitive,
		Content:       contentOnly,
		FileName:      filenameOnly,
	}, nil
}

func getSpanContext(ctx context.Context) (shouldTrace bool, spanContext map[string]string) {
	if !ot.ShouldTrace(ctx) {
		return false, nil
	}

	spanContext = make(map[string]string)
	if err := ot.GetTracer(ctx).Inject(opentracing.SpanFromContext(ctx).Context(), opentracing.TextMap, opentracing.TextMapCarrier(spanContext)); err != nil {
		log15.Warn("Error injecting span context into map: %s", err)
		return true, nil
	}
	return true, spanContext
}

func SearchOpts(ctx context.Context, k int, query *search.TextPatternInfo) zoekt.SearchOptions {
	shouldTrace, spanContext := getSpanContext(ctx)
	searchOpts := zoekt.SearchOptions{
		Trace:                  shouldTrace,
		SpanContext:            spanContext,
		MaxWallTime:            defaultTimeout,
		ShardMaxMatchCount:     100 * k,
		TotalMaxMatchCount:     100 * k,
		ShardMaxImportantMatch: 15 * k,
		TotalMaxImportantMatch: 25 * k,
		MaxDocDisplayCount:     2 * defaultMaxSearchResults,
	}

	// We want zoekt to return more than FileMatchLimit results since we use
	// the extra results to populate reposLimitHit. Additionally the defaults
	// are very low, so we always want to return at least 2000.
	if query.FileMatchLimit > defaultMaxSearchResults {
		searchOpts.MaxDocDisplayCount = 2 * int(query.FileMatchLimit)
	}
	if searchOpts.MaxDocDisplayCount < 2000 {
		searchOpts.MaxDocDisplayCount = 2000
	}

	if userProbablyWantsToWaitLonger := query.FileMatchLimit > defaultMaxSearchResults; userProbablyWantsToWaitLonger {
		searchOpts.MaxWallTime *= time.Duration(3 * float64(query.FileMatchLimit) / float64(defaultMaxSearchResults))
	}

	return searchOpts
}

func ResultCountFactor(numRepos int, fileMatchLimit int32, globalSearch bool) (k int) {
	if globalSearch {
		// for globalSearch, numRepos = 0, but effectively we are searching over all
		// indexed repos, hence k should be 1
		k = 1
	} else {
		// If we're only searching a small number of repositories, return more
		// comprehensive results. This is arbitrary.
		switch {
		case numRepos <= 5:
			k = 100
		case numRepos <= 10:
			k = 10
		case numRepos <= 25:
			k = 8
		case numRepos <= 50:
			k = 5
		case numRepos <= 100:
			k = 3
		case numRepos <= 500:
			k = 2
		default:
			k = 1
		}
	}
	if fileMatchLimit > defaultMaxSearchResults {
		k = int(float64(k) * 3 * float64(fileMatchLimit) / float64(defaultMaxSearchResults))
	}
	return k
}

// RepoRevFunc is a function which maps repository names returned from Zoekt
// into the Sourcegraph's resolved repository revisions for the search.
type RepoRevFunc func(file *zoekt.FileMatch) (repo *types.RepoName, revs []string, ok bool)

// MatchLimiter is the logic which limits files based on limit. Additionally
// it calculates the set of repos with partial results. This information is
// not returned by zoekt, so if zoekt indicates a limit has been hit, we
// include all repos in partial.
type MatchLimiter struct {
	Limit int
}

// Slice will return the set of timed out repositories and the slice of files
// respecting the remaining limit.
func (m *MatchLimiter) Slice(files []zoekt.FileMatch, getRepoInputRev RepoRevFunc) (map[api.RepoID]struct{}, []zoekt.FileMatch) {
	partial, files := limitMatches(m.Limit, files, getRepoInputRev)
	m.Limit -= len(files)
	return partial, files
}

func limitMatches(limit int, files []zoekt.FileMatch, getRepoInputRev RepoRevFunc) (map[api.RepoID]struct{}, []zoekt.FileMatch) {
	if limit < 0 {
		limit = 0
	}

	if len(files) <= limit {
		return nil, files
	}

	resultFiles := files[:limit]
	partialFiles := files[limit:]

	partial := make(map[api.RepoID]struct{})
	last := ""
	for _, file := range partialFiles {
		// PERF: skip lookup if it is the same repo as the last result
		if file.Repository == last {
			continue
		}
		last = file.Repository

		if repo, _, ok := getRepoInputRev(&file); ok {
			partial[repo.ID] = struct{}{}
		}
	}

	return partial, resultFiles
}
