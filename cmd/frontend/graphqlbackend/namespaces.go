package graphqlbackend

import (
	"context"

	"github.com/pkg/errors"

	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"
)

// Namespace is the interface for the GraphQL Namespace interface.
type Namespace interface {
	ID() graphql.ID
	URL() string
	NamespaceName() string
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
	case "User":
		return UserByID(ctx, id)
	case "Org":
		return OrgByID(ctx, id)
	default:
		return nil, errors.New("invalid ID for namespace")
	}
}

func UnmarshalNamespaceID(id graphql.ID, userID *int32, orgID *int32) (err error) {
	switch relay.UnmarshalKind(id) {
	case "User":
		err = relay.UnmarshalSpec(id, userID)
	case "Org":
		err = relay.UnmarshalSpec(id, orgID)
	default:
		err = errors.Errorf("Invalid namespace %q", id)
	}
	return err
}

// NamespaceResolver resolves the GraphQL Namespace interface to a type.
type NamespaceResolver struct {
	Namespace
}

func (r NamespaceResolver) ToOrg() (*OrgResolver, bool) {
	n, ok := r.Namespace.(*OrgResolver)
	return n, ok
}

func (r NamespaceResolver) ToUser() (*UserResolver, bool) {
	n, ok := r.Namespace.(*UserResolver)
	return n, ok
}
