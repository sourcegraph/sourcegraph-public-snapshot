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
	CatalogComponent(context.Context, *CatalogComponentArgs) (CatalogComponentResolver, error)

	NodeResolvers() map[string]NodeByIDFunc
}

type CatalogComponentArgs struct {
	Name string
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
	Description() *string
	Owner(context.Context) (*PersonResolver, error)
	System() *string
	Tags() []string
	URL() string

	Readme(context.Context) (FileResolver, error)
	SourceLocations(context.Context) ([]*GitTreeEntryResolver, error)
	Commits(context.Context, *graphqlutil.ConnectionArgs) (GitCommitConnectionResolver, error)
	Authors(context.Context) (*[]CatalogComponentAuthorEdgeResolver, error)
	Usage(context.Context, *CatalogComponentUsageArgs) (CatalogComponentUsageResolver, error)
	API(context.Context, *CatalogComponentAPIArgs) (CatalogComponentAPIResolver, error)
}

type CatalogComponentKind string

type CatalogComponentAuthorEdgeResolver interface {
	Component() CatalogComponentResolver
	Person() *PersonResolver
	AuthoredLineCount() int32
	AuthoredLineProportion() float64
	LastCommit(context.Context) (*GitCommitResolver, error)
}

type CatalogComponentUsageArgs struct {
	Query *string
}

type CatalogComponentUsageResolver interface {
	Locations(context.Context) (LocationConnectionResolver, error)
	Callers(context.Context) ([]CatalogComponentCallerEdgeResolver, error)
}

type CatalogComponentCallerEdgeResolver interface {
	Component() CatalogComponentResolver
	Person() *PersonResolver
	Locations(context.Context) (LocationConnectionResolver, error)
	AuthoredLineCount() int32
	LastCommit(context.Context) (*GitCommitResolver, error)
}

type CatalogComponentAPIArgs struct {
	Query *string
}

type CatalogComponentAPIResolver interface {
	Symbols(context.Context, *CatalogComponentAPISymbolsArgs) (*SymbolConnectionResolver, error)
}

type CatalogComponentAPISymbolsArgs struct {
	graphqlutil.ConnectionArgs
	Query *string
}
