package run

import (
	"context"
	"fmt"

	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/featureflag"
	"github.com/sourcegraph/sourcegraph/internal/search"
	"github.com/sourcegraph/sourcegraph/internal/search/limits"
	"github.com/sourcegraph/sourcegraph/internal/search/query"
	"github.com/sourcegraph/sourcegraph/internal/search/searchcontexts"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/schema"
)

// SearchInputs contains fields we set before kicking off search.
type SearchInputs struct {
	Plan                query.Plan // the comprehensive query plan
	Query               query.Q    // the current basic query being evaluated, one part of query.Plan
	OriginalQuery       string     // the raw string of the original search query
	PatternType         query.SearchType
	UserSettings        *schema.Settings
	OnSourcegraphDotCom bool
	Features            featureflag.FlagSet
	Protocol            search.Protocol
}

// MaxResults computes the limit for the query.
func (inputs SearchInputs) MaxResults() int {
	return inputs.Query.MaxResults(inputs.DefaultLimit())
}

// DefaultLimit is the default limit to use if not specified in query.
func (inputs SearchInputs) DefaultLimit() int {
	if inputs.Protocol == search.Batch || inputs.PatternType == query.SearchTypeStructural {
		return limits.DefaultMaxSearchResults
	}
	return limits.DefaultMaxSearchResultsStreaming
}

func NewSearchInputs(
	ctx context.Context,
	db database.DB,
	version string,
	patternType *string,
	searchQuery string,
	protocol search.Protocol,
	settings *schema.Settings,
	sourcegraphDotComMode bool,
) (_ *SearchInputs, err error) {
	tr, ctx := trace.New(ctx, "NewSearchInputs", searchQuery)
	defer func() {
		tr.SetError(err)
		tr.Finish()
	}()

	searchType, err := detectSearchType(version, patternType)
	if err != nil {
		return nil, err
	}
	searchType = overrideSearchType(searchQuery, searchType)

	if searchType == query.SearchTypeStructural && !conf.StructuralSearchEnabled() {
		return nil, errors.New("Structural search is disabled in the site configuration.")
	}

	// Beta: create a step to replace each context in the query with its repository query if any.
	searchContextsQueryEnabled := settings.ExperimentalFeatures != nil && getBoolPtr(settings.ExperimentalFeatures.SearchContextsQuery, true)
	substituteContextsStep := query.SubstituteSearchContexts(func(context string) (string, error) {
		sc, err := searchcontexts.ResolveSearchContextSpec(ctx, db, context)
		if err != nil {
			return "", err
		}
		tr.LazyPrintf("substitute query %s for context %s", sc.Query, context)
		return sc.Query, nil
	})

	var plan query.Plan
	plan, err = query.Pipeline(
		query.Init(searchQuery, searchType),
		query.With(searchContextsQueryEnabled, substituteContextsStep),
	)
	if err != nil {
		return nil, &QueryError{Query: searchQuery, Err: err}
	}
	tr.LazyPrintf("parsing done")

	inputs := &SearchInputs{
		Plan:                plan,
		Query:               plan.ToQ(),
		OriginalQuery:       searchQuery,
		UserSettings:        settings,
		OnSourcegraphDotCom: sourcegraphDotComMode,
		Features:            featureflag.FromContext(ctx),
		PatternType:         searchType,
		Protocol:            protocol,
	}

	tr.LazyPrintf("Parsed query: %s", inputs.Query)

	return inputs, nil
}

type QueryError struct {
	Query string
	Err   error
}

func (e *QueryError) Error() string {
	return fmt.Sprintf("invalid query %q: %s", e.Query, e.Err)
}

// detectSearchType returns the search type to perform. The search type derives
// from three sources: the version and patternType parameters passed to the
// search endpoint (literal search is the default in V2), and the `patternType:`
// filter in the input query string which overrides the searchType, if present.
func detectSearchType(version string, patternType *string) (query.SearchType, error) {
	var searchType query.SearchType
	if patternType != nil {
		switch *patternType {
		case "literal":
			searchType = query.SearchTypeLiteralDefault
		case "regexp":
			searchType = query.SearchTypeRegex
		case "structural":
			searchType = query.SearchTypeStructural
		case "lucky":
			searchType = query.SearchTypeLucky
		default:
			return -1, errors.Errorf("unrecognized patternType %q", *patternType)
		}
	} else {
		switch version {
		case "V1":
			searchType = query.SearchTypeRegex
		case "V2":
			searchType = query.SearchTypeLiteralDefault
		default:
			return -1, errors.Errorf("unrecognized version: want \"V1\" or \"V2\", got %q", version)
		}
	}
	return searchType, nil
}

func overrideSearchType(input string, searchType query.SearchType) query.SearchType {
	q, err := query.Parse(input, query.SearchTypeLiteralDefault)
	q = query.LowercaseFieldNames(q)
	if err != nil {
		// If parsing fails, return the default search type. Any actual
		// parse errors will be raised by subsequent parser calls.
		return searchType
	}
	query.VisitField(q, "patterntype", func(value string, _ bool, _ query.Annotation) {
		switch value {
		case "regex", "regexp":
			searchType = query.SearchTypeRegex
		case "literal":
			searchType = query.SearchTypeLiteralDefault
		case "structural":
			searchType = query.SearchTypeStructural
		case "lucky":
			searchType = query.SearchTypeLucky
		}
	})
	return searchType
}

func getBoolPtr(b *bool, def bool) bool {
	if b == nil {
		return def
	}
	return *b
}
