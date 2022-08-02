package client

import (
	"context"
	"fmt"

	"github.com/google/zoekt"
	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/endpoint"
	"github.com/sourcegraph/sourcegraph/internal/featureflag"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/search"
	"github.com/sourcegraph/sourcegraph/internal/search/job"
	"github.com/sourcegraph/sourcegraph/internal/search/job/jobutil"
	"github.com/sourcegraph/sourcegraph/internal/search/query"
	"github.com/sourcegraph/sourcegraph/internal/search/searchcontexts"
	"github.com/sourcegraph/sourcegraph/internal/search/streaming"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/schema"
)

type SearchClient interface {
	Plan(
		ctx context.Context,
		version string,
		patternType *string,
		searchQuery string,
		protocol search.Protocol,
		settings *schema.Settings,
		sourcegraphDotComMode bool,
	) (*search.Inputs, error)

	Execute(
		ctx context.Context,
		stream streaming.Sender,
		inputs *search.Inputs,
	) (_ *search.Alert, err error)

	JobClients() job.RuntimeClients
}

func NewSearchClient(logger log.Logger, db database.DB, zoektStreamer zoekt.Streamer, searcherURLs *endpoint.Map) SearchClient {
	return &searchClient{
		logger:       logger,
		db:           db,
		zoekt:        zoektStreamer,
		searcherURLs: searcherURLs,
	}
}

type searchClient struct {
	logger       log.Logger
	db           database.DB
	zoekt        zoekt.Streamer
	searcherURLs *endpoint.Map
}

func (s *searchClient) Plan(
	ctx context.Context,
	version string,
	patternType *string,
	searchQuery string,
	protocol search.Protocol,
	settings *schema.Settings,
	sourcegraphDotComMode bool,
) (_ *search.Inputs, err error) {
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
		sc, err := searchcontexts.ResolveSearchContextSpec(ctx, s.db, context)
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

	inputs := &search.Inputs{
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

func (s *searchClient) Execute(
	ctx context.Context,
	stream streaming.Sender,
	inputs *search.Inputs,
) (_ *search.Alert, err error) {
	tr, ctx := trace.New(ctx, "Execute", "")
	defer func() {
		tr.SetError(err)
		tr.Finish()
	}()

	planJob, err := jobutil.NewPlanJob(inputs, inputs.Plan)
	if err != nil {
		return nil, err
	}

	return planJob.Run(ctx, s.JobClients(), stream)
}

func (s *searchClient) JobClients() job.RuntimeClients {
	return job.RuntimeClients{
		Logger:       s.logger,
		DB:           s.db,
		Zoekt:        s.zoekt,
		SearcherURLs: s.searcherURLs,
		Gitserver:    gitserver.NewClient(s.db),
	}
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
// search endpoint and the `patternType:` filter in the input query string which
// overrides the searchType, if present.
func detectSearchType(version string, patternType *string) (query.SearchType, error) {
	var searchType query.SearchType
	if patternType != nil {
		switch *patternType {
		case "standard":
			searchType = query.SearchTypeStandard
		case "literal":
			searchType = query.SearchTypeLiteral
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
			searchType = query.SearchTypeLiteral
		case "V3":
			searchType = query.SearchTypeStandard
		default:
			return -1, errors.Errorf("unrecognized version: want \"V1\", \"V2\", or \"V3\", got %q", version)
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
		case "standard":
			searchType = query.SearchTypeStandard
		case "regex", "regexp":
			searchType = query.SearchTypeRegex
		case "literal":
			searchType = query.SearchTypeLiteral
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
