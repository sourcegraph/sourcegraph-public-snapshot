package graphqlbackend

import (
	"context"

	"github.com/graph-gophers/graphql-go"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

type SearchContextsOrderBy string

const (
	SearchContextCursorKind                              = "SearchContextCursor"
	SearchContextsOrderByUpdatedAt SearchContextsOrderBy = "SEARCH_CONTEXT_UPDATED_AT"
	SearchContextsOrderBySpec      SearchContextsOrderBy = "SEARCH_CONTEXT_SPEC"
)

type SearchContextsResolver interface {
	SearchContexts(ctx context.Context, args *ListSearchContextsArgs) (SearchContextConnectionResolver, error)
	AutoDefinedSearchContexts(ctx context.Context) ([]SearchContextResolver, error)

	SearchContextByID(ctx context.Context, id graphql.ID) (SearchContextResolver, error)
	SearchContextBySpec(ctx context.Context, args SearchContextBySpecArgs) (SearchContextResolver, error)
	IsSearchContextAvailable(ctx context.Context, args IsSearchContextAvailableArgs) (bool, error)
	CreateSearchContext(ctx context.Context, args CreateSearchContextArgs) (SearchContextResolver, error)
	UpdateSearchContext(ctx context.Context, args UpdateSearchContextArgs) (SearchContextResolver, error)
	DeleteSearchContext(ctx context.Context, args DeleteSearchContextArgs) (*EmptyResponse, error)

	NodeResolvers() map[string]NodeByIDFunc
	SearchContextsToResolvers(searchContexts []*types.SearchContext) []SearchContextResolver
}

type SearchContextResolver interface {
	ID() graphql.ID
	Name(ctx context.Context) string
	Description(ctx context.Context) string
	Public(ctx context.Context) bool
	AutoDefined(ctx context.Context) bool
	Spec() string
	UpdatedAt(ctx context.Context) DateTime
	Namespace(ctx context.Context) (*NamespaceResolver, error)
	ViewerCanManage(ctx context.Context) bool
	Repositories(ctx context.Context) ([]SearchContextRepositoryRevisionsResolver, error)
}

type SearchContextConnectionResolver interface {
	Nodes(ctx context.Context) ([]SearchContextResolver, error)
	TotalCount(ctx context.Context) (int32, error)
	PageInfo(ctx context.Context) (*graphqlutil.PageInfo, error)
}

type SearchContextRepositoryRevisionsResolver interface {
	Repository(ctx context.Context) *RepositoryResolver
	Revisions(ctx context.Context) []string
}

type SearchContextInputArgs struct {
	Name        string
	Description string
	Public      bool
	Namespace   *graphql.ID
}

type SearchContextEditInputArgs struct {
	Name        string
	Description string
	Public      bool
}

type SearchContextRepositoryRevisionsInputArgs struct {
	RepositoryID graphql.ID
	Revisions    []string
}

type CreateSearchContextArgs struct {
	SearchContext SearchContextInputArgs
	Repositories  []SearchContextRepositoryRevisionsInputArgs
}

type UpdateSearchContextArgs struct {
	ID            graphql.ID
	SearchContext SearchContextEditInputArgs
	Repositories  []SearchContextRepositoryRevisionsInputArgs
}

type DeleteSearchContextArgs struct {
	ID graphql.ID
}

type SearchContextBySpecArgs struct {
	Spec string
}

type IsSearchContextAvailableArgs struct {
	Spec string
}

type ListSearchContextsArgs struct {
	First      int32
	After      *string
	Query      *string
	Namespaces []*graphql.ID
	OrderBy    SearchContextsOrderBy
	Descending bool
}
