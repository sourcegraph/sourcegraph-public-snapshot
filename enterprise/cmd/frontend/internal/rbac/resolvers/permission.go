package resolvers

import (
	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"

	gql "github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/internal/gqlutil"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

type permissionResolver struct {
	permission *types.Permission
}

var _ gql.PermissionResolver = &permissionResolver{}

const permissionIDKind = "Permission"

func marshalPermissionID(id int32) graphql.ID { return relay.MarshalID(permissionIDKind, id) }

func unmarshalPermissionID(id graphql.ID) (permissionID int32, err error) {
	err = relay.UnmarshalSpec(id, &permissionID)
	return
}

func (r *permissionResolver) ID() graphql.ID {
	// 🚨 SECURITY: This needs to be the marshalled.
	return marshalPermissionID(r.permission.ID)
}

func (r *permissionResolver) Namespace() string {
	return r.permission.Namespace
}

func (r *permissionResolver) Action() string {
	return r.permission.Action
}

func (r *permissionResolver) CreatedAt() gqlutil.DateTime {
	return gqlutil.DateTime{Time: r.permission.CreatedAt}
}
