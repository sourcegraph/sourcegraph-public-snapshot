package graphqlbackend

import (
	"context"

	"github.com/cockroachdb/errors"
	"github.com/google/zoekt"
	"github.com/hashicorp/go-multierror"

	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/endpoint"
	"github.com/sourcegraph/sourcegraph/internal/featureflag"
	"github.com/sourcegraph/sourcegraph/internal/search"
	"github.com/sourcegraph/sourcegraph/internal/search/query"
	"github.com/sourcegraph/sourcegraph/internal/search/run"
	"github.com/sourcegraph/sourcegraph/internal/search/searchcontexts"
	"github.com/sourcegraph/sourcegraph/internal/search/streaming"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/schema"
)

type SearchArgs struct {
	Version     string
	PatternType *string
	Query       string

	// Stream if non-nil will stream all SearchEvents.
	//
	// This is how our streaming and our batch interface co-exist. When this
	// is set, it exposes a way to stream out results as we collect them.
	//
	// TODO(keegan) This is not our final design. For example this doesn't
	// allow us to stream out things like dynamic filters or take into account
	// AND/OR. However, streaming is behind a feature flag for now, so this is
	// to make it visible in the browser.
	Stream streaming.Sender

	// For tests
	Settings *schema.Settings
}

type SearchImplementer interface {
	Results(context.Context) (*SearchResultsResolver, error)
	//lint:ignore U1000 is used by graphql via reflection
	Stats(context.Context) (*searchResultsStats, error)

	Inputs() run.SearchInputs
}

// NewSearchImplementer returns a SearchImplementer that provides search results and suggestions.
func NewSearchImplementer(ctx context.Context, db database.DB, args *SearchArgs) (_ SearchImplementer, err error) {
	tr, ctx := trace.New(ctx, "NewSearchImplementer", args.Query)
	defer func() {
		tr.SetError(err)
		tr.Finish()
	}()

	settings := args.Settings
	if settings == nil {
		var err error
		settings, err = decodedViewerFinalSettings(ctx, db)
		if err != nil {
			return nil, err
		}
	}

	searchType, err := detectSearchType(args.Version, args.PatternType)
	if err != nil {
		return nil, err
	}
	searchType = overrideSearchType(args.Query, searchType)

	if searchType == query.SearchTypeStructural && !conf.StructuralSearchEnabled() {
		return nil, errors.New("Structural search is disabled in the site configuration.")
	}

	var plan query.Plan
	plan, err = query.Pipeline(query.Init(args.Query, searchType))
	if err != nil {
		return alertForQuery(args.Query, err).wrapSearchImplementer(db), nil
	}
	tr.LazyPrintf("parsing done")

	if settings.ExperimentalFeatures != nil && getBoolPtr(settings.ExperimentalFeatures.SearchContextsQuery, false) {
		// Replace each context in the query with its repository query if any.
		plan, err = substituteSearchContexts(ctx, db, plan)
		if err != nil {
			return alertForQuery(args.Query, err).wrapSearchImplementer(db), nil
		}
		tr.LazyPrintf("context substitution done")
	}

	defaultLimit := defaultMaxSearchResults
	if args.Stream != nil {
		defaultLimit = defaultMaxSearchResultsStreaming
	}
	if searchType == query.SearchTypeStructural {
		// Set a lower max result count until structural search supports true streaming.
		defaultLimit = defaultMaxSearchResults
	}

	inputs := &run.SearchInputs{
		Plan:          plan,
		Query:         plan.ToParseTree(),
		OriginalQuery: args.Query,
		UserSettings:  settings,
		Features:      featureflag.FromContext(ctx),
		PatternType:   searchType,
		DefaultLimit:  defaultLimit,
	}

	tr.LazyPrintf("Parsed query: %s", inputs.Query)

	return &searchResolver{
		db:           db,
		SearchInputs: inputs,
		stream:       args.Stream,
		zoekt:        search.Indexed(),
		searcherURLs: search.SearcherURLs(),
	}, nil
}

func (r *schemaResolver) Search(ctx context.Context, args *SearchArgs) (SearchImplementer, error) {
	return NewSearchImplementer(ctx, r.db, args)
}

// detectSearchType returns the search type to perform ("regexp", or
// "literal"). The search type derives from three sources: the version and
// patternType parameters passed to the search endpoint (literal search is the
// default in V2), and the `patternType:` filter in the input query string which
// overrides the searchType, if present.
func detectSearchType(version string, patternType *string) (query.SearchType, error) {
	var searchType query.SearchType
	if patternType != nil {
		switch *patternType {
		case "literal":
			searchType = query.SearchTypeLiteral
		case "regexp":
			searchType = query.SearchTypeRegex
		case "structural":
			searchType = query.SearchTypeStructural
		default:
			return -1, errors.Errorf("unrecognized patternType: %v", patternType)
		}
	} else {
		switch version {
		case "V1":
			searchType = query.SearchTypeRegex
		case "V2":
			searchType = query.SearchTypeLiteral
		default:
			return -1, errors.Errorf("unrecognized version want \"V1\" or \"V2\": %v", version)
		}
	}
	return searchType, nil
}

func overrideSearchType(input string, searchType query.SearchType) query.SearchType {
	q, err := query.Parse(input, query.SearchTypeLiteral)
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
			searchType = query.SearchTypeLiteral
		case "structural":
			searchType = query.SearchTypeStructural
		}
	})
	return searchType
}

func substituteSearchContexts(ctx context.Context, db database.DB, plan query.Plan) (query.Plan, error) {
	errs := new(multierror.Error)
	dnf := query.Dnf(query.MapParameter(plan.ToParseTree(), func(field, value string, negated bool, a query.Annotation) query.Node {
		p := query.Parameter{
			Value:   value,
			Field:   field,
			Negated: negated,
		}

		if field != query.FieldContext {
			return p
		}

		sc, err := searchcontexts.ResolveSearchContextSpec(ctx, db, value)
		if err != nil {
			errs = multierror.Append(errs, err)
			return p
		}

		if sc.Query == "" {
			return p
		}

		contextQuery, err := query.Pipeline(query.Init(sc.Query, query.SearchTypeRegex))
		if err != nil {
			errs = multierror.Append(errs, err)
			return p
		}

		return contextQuery.ToParseTree()[0]
	}))

	if err := errs.ErrorOrNil(); err != nil {
		return nil, err
	}

	return query.ToPlan(dnf)
}

func getBoolPtr(b *bool, def bool) bool {
	if b == nil {
		return def
	}
	return *b
}

// searchResolver is a resolver for the GraphQL type `Search`
type searchResolver struct {
	*run.SearchInputs
	db database.DB

	// stream if non-nil will send all search events we receive down it.
	stream streaming.Sender

	zoekt        zoekt.Streamer
	searcherURLs *endpoint.Map
}

func (r *searchResolver) Inputs() run.SearchInputs {
	return *r.SearchInputs
}

// rawQuery returns the original query string input.
func (r *searchResolver) rawQuery() string {
	return r.OriginalQuery
}

// protocol returns what type of search we are doing (batch, stream,
// paginated).
func (r *searchResolver) protocol() search.Protocol {
	if r.stream != nil {
		return search.Streaming
	}
	return search.Batch
}

const (
	defaultMaxSearchResults          = 30
	defaultMaxSearchResultsStreaming = 500
)

var mockDecodedViewerFinalSettings *schema.Settings

// decodedViewerFinalSettings returns the final (merged) settings for the viewer
func decodedViewerFinalSettings(ctx context.Context, db database.DB) (_ *schema.Settings, err error) {
	tr, ctx := trace.New(ctx, "decodedViewerFinalSettings", "")
	defer func() {
		tr.SetError(err)
		tr.Finish()
	}()
	if mockDecodedViewerFinalSettings != nil {
		return mockDecodedViewerFinalSettings, nil
	}

	cascade, err := (&schemaResolver{db: db}).ViewerSettings(ctx)
	if err != nil {
		return nil, err
	}

	return cascade.finalTyped(ctx)
}
