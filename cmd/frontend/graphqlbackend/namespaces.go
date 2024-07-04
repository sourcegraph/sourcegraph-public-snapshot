package graphqlbackend

import (
	"context"
	"fmt"

	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/internal/database"
)

// Namespace is the interface for the GraphQL Namespace interface.
type Namespace interface {
	ID() graphql.ID
	URL() string
	NamespaceName() string
}

func (r *schemaResolver) Namespace(ctx context.Context, args *struct{ ID graphql.ID }) (*NamespaceResolver, error) {
	n, err := NamespaceByID(ctx, r.db, args.ID)
	if err != nil {
		return nil, err
	}
	return &NamespaceResolver{n}, nil
}

type InvalidNamespaceIDErr struct {
	id graphql.ID
}

func (e InvalidNamespaceIDErr) Error() string {
	return fmt.Sprintf("invalid ID %q for namespace", e.id)
}

// NamespaceByID looks up a GraphQL value of type Namespace by ID.
func NamespaceByID(ctx context.Context, db database.DB, id graphql.ID) (Namespace, error) {
	switch relay.UnmarshalKind(id) {
	case "User":
		return UserByID(ctx, db, id)
	case "Org":
		return OrgByID(ctx, db, id)
	default:
		return nil, InvalidNamespaceIDErr{id: id}
	}
}

func UnmarshalNamespaceID(id graphql.ID, userID *int32, orgID *int32) (err error) {
	switch relay.UnmarshalKind(id) {
	case "User":
		err = relay.UnmarshalSpec(id, userID)
	case "Org":
		err = relay.UnmarshalSpec(id, orgID)
	default:
		err = InvalidNamespaceIDErr{id: id}
	}
	return err
}

// UnmarshalNamespaceToIDs is similar to UnmarshalNamespaceID, except instead of
// unmarshalling into existing variables, it creates its own for convenience.
// It will return exactly one non-nil value.
func UnmarshalNamespaceToIDs(id graphql.ID) (userID *int32, orgID *int32, err error) {
	switch relay.UnmarshalKind(id) {
	case "User":
		var uid int32
		err = relay.UnmarshalSpec(id, &uid)
		return &uid, nil, err
	case "Org":
		var oid int32
		err = relay.UnmarshalSpec(id, &oid)
		return nil, &oid, err
	default:
		return nil, nil, InvalidNamespaceIDErr{id: id}
	}
}

func (r *schemaResolver) NamespaceByName(ctx context.Context, args *struct{ Name string }) (*NamespaceResolver, error) {
	namespace, err := r.db.Namespaces().GetByName(ctx, args.Name)
	if err == database.ErrNamespaceNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	var n Namespace
	switch {
	case namespace.User != 0:
		n, err = UserByIDInt32(ctx, r.db, namespace.User)
	case namespace.Organization != 0:
		n, err = OrgByIDInt32(ctx, r.db, namespace.Organization)
	default:
		panic("invalid namespace (neither user nor organization)")
	}
	if err != nil {
		return nil, err
	}
	return &NamespaceResolver{n}, nil
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

func newNamespaceConnection(namespaces []*NamespaceResolver) graphqlutil.SliceConnectionResolver[*NamespaceResolver] {
	return graphqlutil.NewSliceConnectionResolver(namespaces, len(namespaces), len(namespaces))
}
