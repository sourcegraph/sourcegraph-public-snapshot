package graphqlbackend

import (
	"context"
	"errors"

	graphql "github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
)

// Namespace is the interface for the GraphQL Namespace interface.
type Namespace interface {
	ID() graphql.ID
	Projects(context.Context, *graphqlutil.ConnectionArgs) (ProjectConnection, error)
}

func (r *schemaResolver) Namespace(ctx context.Context, args *struct{ ID graphql.ID }) (*NamespaceResolver, error) {
	n, err := NamespaceByID(ctx, args.ID)
	if err != nil {
		return nil, err
	}
	return &NamespaceResolver{n}, nil
}

// NamespaceByID looks up a GraphQL value of type Namespace by ID.
func NamespaceByID(ctx context.Context, id graphql.ID) (Namespace, error) {
	switch relay.UnmarshalKind(id) {
	case "Org":
		return OrgByID(ctx, id)
	case "User":
		return UserByID(ctx, id)
	default:
		return nil, errors.New("invalid ID for namespace")
	}
}

// NamespaceResolver resolves the GraphQL Namespace interface to a type.
type NamespaceResolver struct {
	Namespace
}

func (r *NamespaceResolver) ToOrg() (*OrgResolver, bool) {
	n, ok := r.Namespace.(*OrgResolver)
	return n, ok
}

func (r *NamespaceResolver) ToUser() (*UserResolver, bool) {
	n, ok := r.Namespace.(*UserResolver)
	return n, ok
}
