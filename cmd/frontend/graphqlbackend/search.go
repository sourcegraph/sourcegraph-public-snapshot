package graphqlbackend

import (
	"context"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/envvar"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/search"
	"github.com/sourcegraph/sourcegraph/internal/search/client"
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
func NewBatchSearchImplementer(ctx context.Context, logger log.Logger, db database.DB, args *SearchArgs) (_ SearchImplementer, err error) {
	settings, err := DecodedViewerFinalSettings(ctx, db)
	if err != nil {
		return nil, err
	}

	cli := client.NewSearchClient(logger, db, search.Indexed(), search.SearcherURLs())
	inputs, err := cli.Plan(
		ctx,
		args.Version,
		args.PatternType,
		args.Query,
		search.Batch,
		settings,
		envvar.SourcegraphDotComMode(),
	)
	if err != nil {
		var queryErr *client.QueryError
		if errors.As(err, &queryErr) {
			return NewSearchAlertResolver(search.AlertForQuery(queryErr.Query, queryErr.Err)).wrapSearchImplementer(db), nil
		}
		return nil, err
	}

	return &searchResolver{
		logger:       logger.Scoped("BatchSearchSearchImplementer", "provides search results and suggestions"),
		client:       cli,
		db:           db,
		SearchInputs: inputs,
	}, nil
}

func (r *schemaResolver) Search(ctx context.Context, args *SearchArgs) (SearchImplementer, error) {
	return NewBatchSearchImplementer(ctx, r.logger, r.db, args)
}

// searchResolver is a resolver for the GraphQL type `Search`
type searchResolver struct {
	logger       log.Logger
	client       client.SearchClient
	SearchInputs *search.Inputs
	db           database.DB
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

	cascade, err := newSchemaResolver(db).ViewerSettings(ctx)
	if err != nil {
		return nil, err
	}

	return cascade.finalTyped(ctx)
}
