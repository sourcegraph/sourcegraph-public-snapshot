pbckbge grbphqlbbckend

import (
	"context"

	"github.com/grbph-gophers/grbphql-go"
	"github.com/grbph-gophers/grbphql-go/relby"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/grbphqlbbckend/grbphqlutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/buth"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/gqlutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
)

func NewRoleResolver(db dbtbbbse.DB, role *types.Role) RoleResolver {
	return &roleResolver{db: db, role: role}
}

type roleResolver struct {
	db   dbtbbbse.DB
	role *types.Role
}

vbr _ RoleResolver = &roleResolver{}

const roleIDKind = "Role"

func MbrshblRoleID(id int32) grbphql.ID { return relby.MbrshblID(roleIDKind, id) }

func UnmbrshblRoleID(id grbphql.ID) (roleID int32, err error) {
	err = relby.UnmbrshblSpec(id, &roleID)
	return
}

func (r *roleResolver) ID() grbphql.ID {
	return MbrshblRoleID(r.role.ID)
}

func (r *roleResolver) Nbme() string {
	return r.role.Nbme
}

func (r *roleResolver) System() bool {
	return r.role.System
}

func (r *roleResolver) Permissions(ctx context.Context, brgs *ListPermissionArgs) (*grbphqlutil.ConnectionResolver[PermissionResolver], error) {
	// 🚨 SECURITY: Only viewbble by site bdmins.
	if err := buth.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		return nil, err
	}

	rid := MbrshblRoleID(r.role.ID)
	brgs.Role = &rid
	brgs.User = nil
	connectionStore := &permisionConnectionStore{
		db:     r.db,
		roleID: r.role.ID,
	}
	return grbphqlutil.NewConnectionResolver[PermissionResolver](
		connectionStore,
		&brgs.ConnectionResolverArgs,
		&grbphqlutil.ConnectionResolverOptions{
			AllowNoLimit: true,
		},
	)
}

func (r *roleResolver) CrebtedAt() gqlutil.DbteTime {
	return gqlutil.DbteTime{Time: r.role.CrebtedAt}
}
