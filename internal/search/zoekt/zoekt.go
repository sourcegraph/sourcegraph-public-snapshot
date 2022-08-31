package zoekt

import (
	"context"
	"regexp/syntax" //nolint:depguard // zoekt requires this pkg
	"time"

	"github.com/inconshreveable/log15"
	"github.com/opentracing/opentracing-go"
	"github.com/sourcegraph/zoekt"
	zoektquery "github.com/sourcegraph/zoekt/query"

	"github.com/sourcegraph/sourcegraph/internal/search/filter"
	"github.com/sourcegraph/sourcegraph/internal/search/limits"
	"github.com/sourcegraph/sourcegraph/internal/trace/ot"
	"github.com/sourcegraph/sourcegraph/internal/trace/policy"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

var defaultTimeout = 20 * time.Second

func FileRe(pattern string, queryIsCaseSensitive bool) (zoektquery.Q, error) {
	return parseRe(pattern, true, false, queryIsCaseSensitive)
}

func noOpAnyChar(re *syntax.Regexp) {
	if re.Op == syntax.OpAnyChar {
		re.Op = syntax.OpAnyCharNotNL
	}
	for _, s := range re.Sub {
		noOpAnyChar(s)
	}
}

const regexpFlags = syntax.ClassNL | syntax.PerlX | syntax.UnicodeGroups

func parseRe(pattern string, filenameOnly bool, contentOnly bool, queryIsCaseSensitive bool) (zoektquery.Q, error) {
	// these are the flags used by zoekt, which differ to searcher.
	re, err := syntax.Parse(pattern, regexpFlags)
	if err != nil {
		return nil, err
	}
	noOpAnyChar(re)

	// OptimizeRegexp currently only converts capture groups into non-capture
	// groups (faster for stdlib regexp to execute).
	re = zoektquery.OptimizeRegexp(re, regexpFlags)

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
	if !policy.ShouldTrace(ctx) {
		return false, nil
	}

	spanContext = make(map[string]string)
	if err := ot.GetTracer(ctx).Inject(opentracing.SpanFromContext(ctx).Context(), opentracing.TextMap, opentracing.TextMapCarrier(spanContext)); err != nil {
		log15.Warn("Error injecting span context into map: %s", err)
		return true, nil
	}
	return true, spanContext
}

func SearchOpts(ctx context.Context, k int, fileMatchLimit int32, selector filter.SelectPath) zoekt.SearchOptions {
	shouldTrace, spanContext := getSpanContext(ctx)
	searchOpts := zoekt.SearchOptions{
		Trace:        shouldTrace,
		SpanContext:  spanContext,
		MaxWallTime:  defaultTimeout,
		ChunkMatches: true,
	}

	if userProbablyWantsToWaitLonger := fileMatchLimit > limits.DefaultMaxSearchResults; userProbablyWantsToWaitLonger {
		searchOpts.MaxWallTime *= time.Duration(3 * float64(fileMatchLimit) / float64(limits.DefaultMaxSearchResults))
	}

	if selector.Root() == filter.Repository {
		searchOpts.ShardRepoMaxMatchCount = 1
	} else {
		searchOpts.ShardMaxMatchCount = 100 * k
		searchOpts.TotalMaxMatchCount = 100 * k
		searchOpts.ShardMaxImportantMatch = 15 * k
		searchOpts.TotalMaxImportantMatch = 25 * k
		// Ask for 2000 more results so we have results to populate
		// RepoStatusLimitHit.
		searchOpts.MaxDocDisplayCount = int(fileMatchLimit) + 2000
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
	if fileMatchLimit > limits.DefaultMaxSearchResults {
		k = int(float64(k) * 3 * float64(fileMatchLimit) / float64(limits.DefaultMaxSearchResults))
	}
	return k
}

// repoRevFunc is a function which maps repository names returned from Zoekt
// into the Sourcegraph's resolved repository revisions for the search.
type repoRevFunc func(file *zoekt.FileMatch) (repo types.MinimalRepo, revs []string)
