package graphqlbackend

import (
	"context"

	"github.com/graph-gophers/graphql-go"

	"github.com/sourcegraph/sourcegraph/internal/gqlutil"
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

	SearchContextByID(ctx context.Context, id graphql.ID) (SearchContextResolver, error)
	SearchContextBySpec(ctx context.Context, args SearchContextBySpecArgs) (SearchContextResolver, error)
	IsSearchContextAvailable(ctx context.Context, args IsSearchContextAvailableArgs) (bool, error)
	DefaultSearchContext(ctx context.Context) (SearchContextResolver, error)
	CreateSearchContext(ctx context.Context, args CreateSearchContextArgs) (SearchContextResolver, error)
	UpdateSearchContext(ctx context.Context, args UpdateSearchContextArgs) (SearchContextResolver, error)
	DeleteSearchContext(ctx context.Context, args DeleteSearchContextArgs) (*EmptyResponse, error)

	CreateSearchContextStar(ctx context.Context, args CreateSearchContextStarArgs) (*EmptyResponse, error)
	DeleteSearchContextStar(ctx context.Context, args DeleteSearchContextStarArgs) (*EmptyResponse, error)
	SetDefaultSearchContext(ctx context.Context, args SetDefaultSearchContextArgs) (*EmptyResponse, error)

	NodeResolvers() map[string]NodeByIDFunc
	SearchContextsToResolvers(searchContexts []*types.SearchContext) []SearchContextResolver
}

type SearchContextResolver interface {
	ID() graphql.ID
	Name() string
	Description() string
	Public() bool
	AutoDefined() bool
	Spec() string
	UpdatedAt() gqlutil.DateTime
	Namespace(ctx context.Context) (*NamespaceResolver, error)
	ViewerCanManage(ctx context.Context) bool
	ViewerHasAsDefault(ctx context.Context) bool
	ViewerHasStarred(ctx context.Context) bool
	Repositories(ctx context.Context) ([]SearchContextRepositoryRevisionsResolver, error)
	Query() string
}

type SearchContextConnectionResolver interface {
	Nodes() []SearchContextResolver
	TotalCount() int32
	PageInfo() *gqlutil.PageInfo
}

type SearchContextRepositoryRevisionsResolver interface {
	Repository() *RepositoryResolver
	Revisions() []string
}

type SearchContextInputArgs struct {
	Name        string
	Description string
	Public      bool
	Namespace   *graphql.ID
	Query       string
}

type SearchContextEditInputArgs struct {
	Name        string
	Description string
	Public      bool
	Query       string
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

type CreateSearchContextStarArgs struct {
	SearchContextID graphql.ID
	UserID          graphql.ID
}

type DeleteSearchContextStarArgs struct {
	SearchContextID graphql.ID
	UserID          graphql.ID
}

type SetDefaultSearchContextArgs struct {
	SearchContextID graphql.ID
	UserID          graphql.ID
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
