package resolvers

import (
	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/gqlutil"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

type roleResolver struct {
	db   database.DB
	role *types.Role
}

var _ graphqlbackend.RoleResolver = &roleResolver{}

const roleIDKind = "Role"

func marshalRoleID(id int32) graphql.ID { return relay.MarshalID(roleIDKind, id) }

func unmarshalRoleID(id graphql.ID) (roleID int32, err error) {
	err = relay.UnmarshalSpec(id, &roleID)
	return
}

func (r *roleResolver) ID() graphql.ID {
	return marshalRoleID(r.role.ID)
}

func (r *roleResolver) Name() string {
	return r.role.Name
}

func (r *roleResolver) System() bool {
	return r.role.System
}

func (r *roleResolver) Permissions() (graphqlbackend.PermissionConnectionResolver, error) {
	return &permissionConnectionResolver{
		db: r.db,
		opts: database.PermissionListOpts{
			RoleID: r.role.ID,
		},
	}, nil
}

func (r *roleResolver) CreatedAt() gqlutil.DateTime {
	return gqlutil.DateTime{Time: r.role.CreatedAt}
}
