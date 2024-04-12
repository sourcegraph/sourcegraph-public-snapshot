package graphqlbackend

import (
	"context"

	"github.com/graph-gophers/graphql-go"
	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/internal/auth"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/search"
	"github.com/sourcegraph/sourcegraph/internal/search/client"
	"github.com/sourcegraph/sourcegraph/lib/errors"
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
	cli := client.New(logger, db, gitserver.NewClient("graphql.batchsearch"))
	inputs, err := cli.Plan(
		ctx,
		args.Version,
		args.PatternType,
		args.Query,
		search.Precise,
		search.Batch,
		nil,
	)
	if err != nil {
		var queryErr *client.QueryError
		if errors.As(err, &queryErr) {
			return NewSearchAlertResolver(search.AlertForQuery(queryErr.Query, queryErr.Err)).wrapSearchImplementer(db), nil
		}
		return nil, err
	}

	return &searchResolver{
		logger:       logger.Scoped("BatchSearchSearchImplementer"),
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

type indexedSearchInstance struct {
	address string
}

func (i *indexedSearchInstance) Address() string {
	return i.address
}

func (i *indexedSearchInstance) ID() graphql.ID {
	return marshalGitserverID(i.address)
}

func (r *schemaResolver) IndexedSearchInstances(ctx context.Context) (graphqlutil.SliceConnectionResolver[*indexedSearchInstance], error) {
	// 🚨 SECURITY: Site admins only.
	if err := auth.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		return nil, err
	}

	indexers := search.Indexers()
	eps, err := indexers.Map.Endpoints()
	if err != nil {
		return nil, err
	}

	var resolvers []*indexedSearchInstance
	for _, ep := range eps {
		resolvers = append(resolvers, &indexedSearchInstance{address: ep})
	}
	n := len(resolvers)

	return graphqlutil.NewSliceConnectionResolver(resolvers, n, n), nil
}
