package graphqlbackend

import (
	"context"
	"errors"
	"sort"

	graphql "github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
)

// Namespace is the interface for the GraphQL Namespace interface.
type Namespace interface {
	ID() graphql.ID
	URL() string
	NamespaceName() string
	Campaigns(context.Context, *graphqlutil.ConnectionArgs) (CampaignConnection, error)
}

func (r *schemaResolver) Namespace(ctx context.Context, args *struct{ ID graphql.ID }) (*NamespaceResolver, error) {
	n, err := NamespaceByID(ctx, args.ID)
	if err != nil {
		return nil, err
	}
	return &NamespaceResolver{n}, nil
}

func (r *schemaResolver) ViewerNamespaces(ctx context.Context) (namespaces []*NamespaceResolver, err error) {
	user, err := CurrentUser(ctx)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, nil
	}
	namespaces = append(namespaces, &NamespaceResolver{user})

	orgs, err := db.Orgs.GetByUserID(ctx, user.user.ID)
	if err != nil {
		return nil, err
	}
	// Stable-sort the organizations.
	sort.Slice(orgs, func(i, j int) bool {
		return orgs[i].ID < orgs[j].ID
	})
	for _, org := range orgs {
		namespaces = append(namespaces, &NamespaceResolver{&OrgResolver{org}})
	}

	return namespaces, nil
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

// NamespaceByDBID looks up a GraphQL value of type Namespace by database ID.
func NamespaceByDBID(ctx context.Context, userID, orgID int32) (*NamespaceResolver, error) {
	switch {
	case userID != 0:
		user, err := UserByIDInt32(ctx, userID)
		if err != nil {
			return nil, err
		}
		return &NamespaceResolver{Namespace: user}, nil
	case orgID != 0:
		org, err := OrgByIDInt32(ctx, orgID)
		if err != nil {
			return nil, err
		}
		return &NamespaceResolver{Namespace: org}, nil
	default:
		return nil, errors.New("unrecognized namespace ID")
	}
}

// NamespaceDBIDByID returns the database ID of the namespace given by its GraphQL ID. At most one
// of the int32 return values will be nonzero.
func NamespaceDBIDByID(ctx context.Context, namespaceID graphql.ID) (userID, orgID int32, err error) {
	namespace, err := NamespaceByID(ctx, namespaceID)
	if err != nil {
		return 0, 0, err
	}

	switch namespace := namespace.(type) {
	case *UserResolver:
		return namespace.DatabaseID(), 0, nil
	case *OrgResolver:
		return 0, namespace.OrgID(), nil
	default:
		return 0, 0, errors.New("namespace must be either an organization or user")
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
