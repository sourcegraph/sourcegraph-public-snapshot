package graphqlbackend

import (
	"context"

	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/internal/auth"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/gqlutil"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

func NewRoleResolver(db database.DB, role *types.Role) RoleResolver {
	return &roleResolver{db: db, role: role}
}

type roleResolver struct {
	db   database.DB
	role *types.Role
}

var _ RoleResolver = &roleResolver{}

const roleIDKind = "Role"

func MarshalRoleID(id int32) graphql.ID { return relay.MarshalID(roleIDKind, id) }

func UnmarshalRoleID(id graphql.ID) (roleID int32, err error) {
	err = relay.UnmarshalSpec(id, &roleID)
	return
}

func (r *roleResolver) ID() graphql.ID {
	return MarshalRoleID(r.role.ID)
}

func (r *roleResolver) Name() string {
	return r.role.Name
}

func (r *roleResolver) System() bool {
	return r.role.System
}

func (r *roleResolver) Permissions(ctx context.Context, args *ListPermissionArgs) (*graphqlutil.ConnectionResolver[PermissionResolver], error) {
	// 🚨 SECURITY: Only viewable by site admins.
	if err := auth.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		return nil, err
	}

	rid := MarshalRoleID(r.role.ID)
	args.Role = &rid
	args.User = nil
	connectionStore := &permissionConnectionStore{
		db:     r.db,
		roleID: r.role.ID,
	}
	return graphqlutil.NewConnectionResolver[PermissionResolver](
		connectionStore,
		&args.ConnectionResolverArgs,
		&graphqlutil.ConnectionResolverOptions{
			AllowNoLimit: true,
		},
	)
}

func (r *roleResolver) CreatedAt() gqlutil.DateTime {
	return gqlutil.DateTime{Time: r.role.CreatedAt}
}
