package zoekt

import (
	"context"
	"regexp/syntax" //nolint:depguard // zoekt requires this pkg
	"time"

	"github.com/opentracing/opentracing-go"
	"github.com/sourcegraph/log"
	"github.com/sourcegraph/zoekt"
	zoektquery "github.com/sourcegraph/zoekt/query"

	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/search"
	"github.com/sourcegraph/sourcegraph/internal/search/filter"
	"github.com/sourcegraph/sourcegraph/internal/trace/ot"
	"github.com/sourcegraph/sourcegraph/internal/trace/policy"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

var defaultTimeout = 20 * time.Second

func FileRe(pattern string, queryIsCaseSensitive bool) (zoektquery.Q, error) {
	return parseRe(pattern, true, false, queryIsCaseSensitive)
}

const regexpFlags = syntax.ClassNL | syntax.PerlX | syntax.UnicodeGroups

func parseRe(pattern string, filenameOnly bool, contentOnly bool, queryIsCaseSensitive bool) (zoektquery.Q, error) {
	// these are the flags used by zoekt, which differ to searcher.
	re, err := syntax.Parse(pattern, regexpFlags)
	if err != nil {
		return nil, err
	}

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

func getSpanContext(ctx context.Context, logger log.Logger) (shouldTrace bool, spanContext map[string]string) {
	if !policy.ShouldTrace(ctx) {
		return false, nil
	}

	spanContext = make(map[string]string)
	if span := opentracing.SpanFromContext(ctx); span != nil {
		if err := ot.GetTracer(ctx).Inject(span.Context(), opentracing.TextMap, opentracing.TextMapCarrier(spanContext)); err != nil { //nolint:staticcheck // Drop once we get rid of OpenTracing
			logger.Warn("Error injecting span context into map", log.Error(err))
			return true, nil
		}
	}
	return true, spanContext
}

// Options represents the inputs from Sourcegraph that we use to compute
// zoekt.SearchOptions.
type Options struct {
	Selector filter.SelectPath

	// FileMatchLimit is how many results the user wants.
	FileMatchLimit int32

	// NumRepos is the number of repos we are searching over. This number is
	// used as a heuristics to scale the amount of work we will do.
	NumRepos int

	// GlobalSearch is true if we are doing a search were we skip computing
	// NumRepos and instead rely on zoekt.
	GlobalSearch bool

	// Features are feature flags that can affect behaviour of searcher.
	Features search.Features
}

func (o *Options) ToSearch(ctx context.Context, logger log.Logger) *zoekt.SearchOptions {
	shouldTrace, spanContext := getSpanContext(ctx, logger)
	searchOpts := &zoekt.SearchOptions{
		Trace:        shouldTrace,
		SpanContext:  spanContext,
		MaxWallTime:  defaultTimeout,
		ChunkMatches: true,
	}

	// If we're searching repos, ignore the other options and only check one file per repo
	if o.Selector.Root() == filter.Repository {
		searchOpts.ShardRepoMaxMatchCount = 1
		return searchOpts
	}

	if o.Features.Debug {
		searchOpts.DebugScore = true
	}

	if o.Features.Ranking {
		// This enables our stream based ranking, where we wait a certain amount
		// of time to collect results before ranking.
		searchOpts.FlushWallTime = conf.SearchFlushWallTime()

		// This enables the use of document ranks in scoring, if they are available.
		searchOpts.UseDocumentRanks = true
		searchOpts.DocumentRanksWeight = conf.SearchDocumentRanksWeight()
	}

	// These are reasonable default amounts of work to do per shard and
	// replica respectively.
	searchOpts.ShardMaxMatchCount = 10_000
	searchOpts.TotalMaxMatchCount = 100_000

	// Tell each zoekt replica to not send back more than limit results.
	limit := int(o.FileMatchLimit)
	searchOpts.MaxDocDisplayCount = limit

	// If we are searching for large limits, raise the amount of work we
	// are willing to do per shard and zoekt replica respectively.
	if limit > searchOpts.ShardMaxMatchCount {
		searchOpts.ShardMaxMatchCount = limit
	}
	if limit > searchOpts.TotalMaxMatchCount {
		searchOpts.TotalMaxMatchCount = limit
	}

	return searchOpts
}

// repoRevFunc is a function which maps repository names returned from Zoekt
// into the Sourcegraph's resolved repository revisions for the search.
type repoRevFunc func(file *zoekt.FileMatch) (repo types.MinimalRepo, revs []string)
