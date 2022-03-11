package graphqlbackend

import (
	"context"

	"github.com/google/zoekt"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/envvar"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/endpoint"
	"github.com/sourcegraph/sourcegraph/internal/search"
	"github.com/sourcegraph/sourcegraph/internal/search/run"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/schema"
)

type SearchArgs struct {
	Version     string
	PatternType *string
	Query       string
}

type SearchImplementer interface {
	Results(context.Context) (*SearchResultsResolver, error)
	//lint:ignore U1000 is used by graphql via reflection
	Stats(context.Context) (*searchResultsStats, error)
}

// NewBatchSearchImplementer returns a SearchImplementer that provides search results and suggestions.
func NewBatchSearchImplementer(ctx context.Context, db database.DB, args *SearchArgs) (_ SearchImplementer, err error) {
	settings, err := DecodedViewerFinalSettings(ctx, db)
	if err != nil {
		return nil, err
	}

	inputs, err := run.NewSearchInputs(
		ctx,
		db,
		args.Version,
		args.PatternType,
		args.Query,
		search.Batch,
		settings,
		envvar.SourcegraphDotComMode(),
	)
	if err != nil {
		var queryErr *run.QueryError
		if errors.As(err, &queryErr) {
			return NewSearchAlertResolver(search.AlertForQuery(queryErr.Query, queryErr.Err)).wrapSearchImplementer(db), nil
		}
		return nil, err
	}

	return &searchResolver{
		db:           db,
		SearchInputs: inputs,
		zoekt:        search.Indexed(),
		searcherURLs: search.SearcherURLs(),
	}, nil
}

func (r *schemaResolver) Search(ctx context.Context, args *SearchArgs) (SearchImplementer, error) {
	return NewBatchSearchImplementer(ctx, r.db, args)
}

// searchResolver is a resolver for the GraphQL type `Search`
type searchResolver struct {
	SearchInputs *run.SearchInputs
	db           database.DB

	zoekt        zoekt.Streamer
	searcherURLs *endpoint.Map
}

var MockDecodedViewerFinalSettings *schema.Settings

// DecodedViewerFinalSettings returns the final (merged) settings for the viewer
func DecodedViewerFinalSettings(ctx context.Context, db database.DB) (_ *schema.Settings, err error) {
	tr, ctx := trace.New(ctx, "decodedViewerFinalSettings", "")
	defer func() {
		tr.SetError(err)
		tr.Finish()
	}()
	if MockDecodedViewerFinalSettings != nil {
		return MockDecodedViewerFinalSettings, nil
	}

	cascade, err := (&schemaResolver{db: db}).ViewerSettings(ctx)
	if err != nil {
		return nil, err
	}

	return cascade.finalTyped(ctx)
}
