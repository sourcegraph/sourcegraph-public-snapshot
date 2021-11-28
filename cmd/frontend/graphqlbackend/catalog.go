package graphqlbackend

import (
	"context"

	"github.com/graph-gophers/graphql-go"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
)

// This file just contains stub GraphQL resolvers and data types for the Catalog which merely return
// an error if not running in enterprise mode. The actual resolvers are in
// enterprise/cmd/frontend/internal/catalog/resolvers.

// CatalogRootResolver is the root resolver.
type CatalogRootResolver interface {
	Catalog(context.Context) (CatalogResolver, error)

	NodeResolvers() map[string]NodeByIDFunc
}

type CatalogResolver interface {
	Components(context.Context, *CatalogComponentsArgs) (CatalogComponentConnectionResolver, error)
}

type CatalogComponentsArgs struct {
	Query *string
	First *int32
	After *string
}

type CatalogComponentConnectionResolver interface {
	Nodes(context.Context) ([]CatalogComponentResolver, error)
	TotalCount(context.Context) (int32, error)
	PageInfo(context.Context) (*graphqlutil.PageInfo, error)
}

type CatalogComponentResolver interface {
	ID() graphql.ID
	Kind() CatalogComponentKind
	Name() string
	Owner(context.Context) (*PersonResolver, error)
	System() *string
	Tags() []string
	SourceLocations(context.Context) ([]*GitTreeEntryResolver, error)
	Commits(context.Context, *graphqlutil.ConnectionArgs) (GitCommitConnectionResolver, error)
	Authors(context.Context) (*[]CatalogComponentAuthorEdgeResolver, error)
}

type CatalogComponentKind string

type CatalogComponentAuthorEdgeResolver interface {
	Component() CatalogComponentResolver
	Person() *PersonResolver
	AuthoredLineCount() int32
	AuthoredLineProportion() float64
	LastCommit(context.Context) (*GitCommitResolver, error)
}
