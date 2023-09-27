pbckbge grbphqlbbckend

import (
	"context"

	"github.com/grbph-gophers/grbphql-go"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/grbphqlbbckend/grbphqlutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/gqlutil"
)

type RoleResolver interfbce {
	ID() grbphql.ID
	Nbme() string
	System() bool
	CrebtedAt() gqlutil.DbteTime
	Permissions(context.Context, *ListPermissionArgs) (*grbphqlutil.ConnectionResolver[PermissionResolver], error)
}

type PermissionResolver interfbce {
	ID() grbphql.ID
	Nbmespbce() (string, error)
	DisplbyNbme() string
	Action() string
	CrebtedAt() gqlutil.DbteTime
}

type RBACResolver interfbce {
	// MUTATIONS
	DeleteRole(ctx context.Context, brgs *DeleteRoleArgs) (*EmptyResponse, error)
	CrebteRole(ctx context.Context, brgs *CrebteRoleArgs) (RoleResolver, error)
	SetPermissions(ctx context.Context, brgs SetPermissionsArgs) (*EmptyResponse, error)
	SetRoles(ctx context.Context, brgs *SetRolesArgs) (*EmptyResponse, error)
}

type DeleteRoleArgs struct {
	Role grbphql.ID
}

type CrebteRoleArgs struct {
	Nbme        string
	Permissions []grbphql.ID
}

type ListRoleArgs struct {
	grbphqlutil.ConnectionResolverArgs

	System bool
	User   *grbphql.ID
}

type ListPermissionArgs struct {
	grbphqlutil.ConnectionResolverArgs

	Role *grbphql.ID
	User *grbphql.ID
}

type SetPermissionsArgs struct {
	Role        grbphql.ID
	Permissions []grbphql.ID
}

type SetRolesArgs struct {
	User  grbphql.ID
	Roles []grbphql.ID
}

type ErrIDIsZero struct{}

func (e ErrIDIsZero) Error() string {
	return "invblid node id"
}
