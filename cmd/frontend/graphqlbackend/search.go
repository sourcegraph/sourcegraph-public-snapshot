package graphqlbackend

import (
	"context"
	"slices"

	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"
	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/auth"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/gqlutil"
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
			return NewSearchAlertResolver(search.AlertForQuery(queryErr.Err)).wrapSearchImplementer(db), nil
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

const indexedSearchInstanceIDKind = "IndexedSearchInstance"

func marshalIndexedSearchInstanceID(id string) graphql.ID {
	return relay.MarshalID(indexedSearchInstanceIDKind, id)
}

type indexedSearchInstance struct {
	address string
}

func (i *indexedSearchInstance) Address() string {
	return i.address
}

func (i *indexedSearchInstance) ID() graphql.ID {
	return marshalIndexedSearchInstanceID(i.address)
}

func unmarshalIndexedSearchInstanceID(id graphql.ID) (indexedSearchInstanceID string, err error) {
	err = relay.UnmarshalSpec(id, &indexedSearchInstanceID)
	return
}

func (r *schemaResolver) indexedSearchInstanceByID(ctx context.Context, id graphql.ID) (*indexedSearchInstance, error) {
	// 🚨 SECURITY: Site admins only.
	if err := auth.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		return nil, err
	}

	address, err := unmarshalIndexedSearchInstanceID(id)
	if err != nil {
		return nil, err
	}

	return &indexedSearchInstance{address: address}, nil
}

func (r *schemaResolver) IndexedSearchInstances(ctx context.Context) (gqlutil.SliceConnectionResolver[*indexedSearchInstance], error) {
	// 🚨 SECURITY: Site admins only.
	if err := auth.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		return nil, err
	}

	indexers := search.Indexers()
	eps, err := indexers.Map.Endpoints()
	if err != nil {
		return nil, err
	}

	slices.Sort(eps)

	var resolvers []*indexedSearchInstance
	for _, ep := range eps {
		resolvers = append(resolvers, &indexedSearchInstance{address: ep})
	}
	n := len(resolvers)

	return gqlutil.NewSliceConnectionResolver(resolvers, n, n), nil
}
