package graphqlbackend

import (
	"context"
	"errors"
	"fmt"

	"github.com/graph-gophers/graphql-go"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
)

type GraphsResolver interface {
	// Mutations
	CreateGraph(ctx context.Context, args *CreateGraphArgs) (GraphResolver, error)
	UpdateGraph(ctx context.Context, args *UpdateGraphArgs) (GraphResolver, error)
	DeleteGraph(ctx context.Context, args *DeleteGraphArgs) (*EmptyResponse, error)

	// Queries
	Graph(ctx context.Context, args GraphArgs) (GraphResolver, error)
	Graphs(ctx context.Context, args GraphConnectionArgs) (GraphConnectionResolver, error)

	// Helpers
	GraphByID(ctx context.Context, id graphql.ID) (GraphResolver, error)
	RepositoriesForGraph(ctx context.Context, id graphql.ID) ([]string, error)
}

type GraphResolver interface {
	ID() graphql.ID
	Owner(context.Context) (*GraphOwnerResolver, error)
	Name() string
	Description() *string
	Spec() string
	URL(context.Context) (string, error)
	EditURL(context.Context) (string, error)
	ViewerCanAdminister(context.Context) (bool, error)
	CreatedAt() DateTime
	UpdatedAt() DateTime
}

type GraphOwner interface {
	ID() graphql.ID
	Graph(ctx context.Context, args *GraphArgs) (GraphResolver, error)
	Graphs(ctx context.Context, args *GraphConnectionArgs) (GraphConnectionResolver, error)
	URL() string
}

var (
	_ GraphOwner = &UserResolver{}
	_ GraphOwner = &OrgResolver{}
)

type GraphOwnerResolver struct {
	GraphOwner
}

func (r *GraphOwnerResolver) ToUser() (*UserResolver, bool) {
	n, ok := r.GraphOwner.(*UserResolver)
	return n, ok
}

func (r *GraphOwnerResolver) ToOrg() (*OrgResolver, bool) {
	n, ok := r.GraphOwner.(*OrgResolver)
	return n, ok
}

// GraphOwner looks up a GraphQL value of type GraphOwner by ID.
func GraphOwnerByID(ctx context.Context, id graphql.ID) (*GraphOwnerResolver, error) {
	// Reuse NamespaceByID because both support User and Org.
	n, err := NamespaceByID(ctx, id)
	if err != nil {
		return nil, err
	}
	switch n := n.(type) {
	case *UserResolver:
		return &GraphOwnerResolver{n}, nil
	case *OrgResolver:
		return &GraphOwnerResolver{n}, nil
	default:
		panic(fmt.Sprintf("unexpected GraphOwner type: %T", n))
	}
}

func UnmarshalGraphOwnerID(id graphql.ID, userID *int32, orgID *int32) error {
	// Reuse UnmarshalNamespaceID because both support User and Org.
	return UnmarshalNamespaceID(id, userID, orgID)
}

type GraphArgs struct {
	Owner   string
	OwnerID graphql.ID

	Name string
}

type GraphConnectionResolver interface {
	Nodes(ctx context.Context) ([]GraphResolver, error)
	TotalCount(ctx context.Context) (int32, error)
	PageInfo(ctx context.Context) (*graphqlutil.PageInfo, error)
}

type GraphConnectionArgs struct {
	First      int32
	After      *string
	Owner      *graphql.ID
	Affiliated bool
}

type CreateGraphArgs struct {
	Input struct {
		Owner       graphql.ID
		Name        string
		Description *string
		Spec        string
	}
}

type UpdateGraphArgs struct {
	Input struct {
		ID          graphql.ID
		Name        string
		Description *string
		Spec        string
	}
}

type DeleteGraphArgs struct {
	Graph graphql.ID
}

var graphsOnlyInEnterprise = errors.New("graphs are only available in enterprise")

type defaultGraphsResolver struct{}

var DefaultGraphsResolver GraphsResolver = defaultGraphsResolver{}

// Mutations
func (defaultGraphsResolver) CreateGraph(ctx context.Context, args *CreateGraphArgs) (GraphResolver, error) {
	return nil, graphsOnlyInEnterprise
}

func (defaultGraphsResolver) UpdateGraph(ctx context.Context, args *UpdateGraphArgs) (GraphResolver, error) {
	return nil, graphsOnlyInEnterprise
}

func (defaultGraphsResolver) DeleteGraph(ctx context.Context, args *DeleteGraphArgs) (*EmptyResponse, error) {
	return nil, graphsOnlyInEnterprise
}

// Queries
func (defaultGraphsResolver) Graph(ctx context.Context, args GraphArgs) (GraphResolver, error) {
	return nil, graphsOnlyInEnterprise
}

func (defaultGraphsResolver) Graphs(ctx context.Context, args GraphConnectionArgs) (GraphConnectionResolver, error) {
	return nil, graphsOnlyInEnterprise
}

// Helpers
func (defaultGraphsResolver) GraphByID(ctx context.Context, id graphql.ID) (GraphResolver, error) {
	return nil, graphsOnlyInEnterprise
}

func (defaultGraphsResolver) RepositoriesForGraph(ctx context.Context, id graphql.ID) ([]string, error) {
	return nil, graphsOnlyInEnterprise
}
