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
	CatalogEntity(context.Context, *CatalogEntityArgs) (*CatalogEntityResolver, error)

	GitTreeEntryCatalogEntities(context.Context, *GitTreeEntryResolver) ([]*CatalogEntityResolver, error)

	NodeResolvers() map[string]NodeByIDFunc
}

type CatalogEntityArgs struct {
	Name string
}

type CatalogResolver interface {
	Entities(context.Context, *CatalogEntitiesArgs) (CatalogEntityConnectionResolver, error)
	Graph(context.Context) (CatalogGraphResolver, error)
}

type CatalogEntitiesArgs struct {
	Query *string
	First *int32
	After *string
}

type CatalogGraphResolver interface {
	Nodes() []*CatalogEntityResolver
	Edges() []CatalogEntityRelationEdgeResolver
}

type CatalogEntityType string

type CatalogEntityLifecycle string

type CatalogEntity interface {
	Node
	Type() CatalogEntityType
	Name() string
	Description() *string
	Owner(context.Context) (*EntityOwnerResolver, error)
	Lifecycle() CatalogEntityLifecycle
	URL() string
	Status(context.Context) (CatalogEntityStatusResolver, error)
	CodeOwners(context.Context) (*[]CatalogEntityOwnerEdgeResolver, error)
	RelatedEntities(context.Context) (CatalogEntityRelatedEntityConnectionResolver, error)
}

type CatalogEntityResolver struct {
	CatalogEntity
}

func (r *CatalogEntityResolver) ToCatalogComponent() (CatalogComponentResolver, bool) {
	e, ok := r.CatalogEntity.(CatalogComponentResolver)
	return e, ok
}

type EntityOwnerResolver struct {
	Person *PersonResolver
	Group  GroupResolver
}

func (r *EntityOwnerResolver) ToPerson() (*PersonResolver, bool) { return r.Person, r.Person != nil }
func (r *EntityOwnerResolver) ToGroup() (GroupResolver, bool)    { return r.Group, r.Group != nil }

type GroupResolver interface {
	Node
	Name() string
	Title() string
}

type CatalogEntityStatusResolver interface {
	ID() graphql.ID
	Contexts() []CatalogEntityStatusContextResolver
	State() CatalogEntityStatusState
}

type CatalogEntityStatusState string

type CatalogEntityStatusContextResolver interface {
	ID() graphql.ID
	Name() string
	State() CatalogEntityStatusState
	Title() string
	Description() *string
	TargetURL() *string
}

type CatalogEntityRelationType string

type CatalogEntityRelationEdgeResolver interface {
	Type() CatalogEntityRelationType
	OutNode() *CatalogEntityResolver
	InNode() *CatalogEntityResolver
}

type CatalogEntityRelatedEntityConnectionResolver interface {
	Edges() []CatalogEntityRelatedEntityEdgeResolver
}

type CatalogEntityRelatedEntityEdgeResolver interface {
	Node() *CatalogEntityResolver
	Type() CatalogEntityRelationType
}

type CatalogEntityConnectionResolver interface {
	Nodes(context.Context) ([]*CatalogEntityResolver, error)
	TotalCount(context.Context) (int32, error)
	PageInfo(context.Context) (*graphqlutil.PageInfo, error)
}

type CatalogComponentResolver interface {
	CatalogEntity
	Kind() CatalogComponentKind

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

type CatalogEntityOwnerEdgeResolver interface {
	Node() string
	FileCount() int32
	FileProportion() float64
}

type CatalogComponentUsageArgs struct {
	Query *string
}

type CatalogComponentUsageResolver interface {
	Locations(context.Context) (LocationConnectionResolver, error)
	People(context.Context) ([]CatalogComponentUsedByPersonEdgeResolver, error)
	Components(context.Context) ([]CatalogComponentUsedByComponentEdgeResolver, error)
}

type CatalogComponentUsedByPersonEdgeResolver interface {
	Node() *PersonResolver
	Locations(context.Context) (LocationConnectionResolver, error)
	AuthoredLineCount() int32
	LastCommit(context.Context) (*GitCommitResolver, error)
}

type CatalogComponentUsedByComponentEdgeResolver interface {
	Node() CatalogComponentResolver
	Locations(context.Context) (LocationConnectionResolver, error)
}

type CatalogComponentAPIArgs struct {
	Query *string
}

type CatalogComponentAPIResolver interface {
	Symbols(context.Context, *CatalogComponentAPISymbolsArgs) (*SymbolConnectionResolver, error)
	Schema(context.Context) (FileResolver, error)
}

type CatalogComponentAPISymbolsArgs struct {
	graphqlutil.ConnectionArgs
	Query *string
}

func (r *GitTreeEntryResolver) CatalogEntities(ctx context.Context) ([]*CatalogEntityResolver, error) {
	return EnterpriseResolvers.catalogRootResolver.GitTreeEntryCatalogEntities(ctx, r)
}
