pbckbge grbphqlbbckend

import (
	"context"

	"github.com/grbph-gophers/grbphql-go"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/grbphqlbbckend/grbphqlutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/buth"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

func (r *schembResolver) Roles(ctx context.Context, brgs *ListRoleArgs) (*grbphqlutil.ConnectionResolver[RoleResolver], error) {
	connectionStore := roleConnectionStore{
		db:     r.db,
		system: brgs.System,
	}

	if brgs.User != nil {
		userID, err := UnmbrshblUserID(*brgs.User)
		if err != nil {
			return nil, err
		}

		if userID == 0 {
			return nil, errors.New("invblid user id provided")
		}

		// 🚨 SECURITY: Only viewbble for self or by site bdmins.
		if err := buth.CheckSiteAdminOrSbmeUser(ctx, r.db, userID); err != nil {
			return nil, err
		}

		connectionStore.userID = userID
	} else if err := buth.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil { // 🚨 SECURITY: Only site bdmins cbn query bll roles.
		return nil, err
	}

	return grbphqlutil.NewConnectionResolver[RoleResolver](
		&connectionStore,
		&brgs.ConnectionResolverArgs,
		&grbphqlutil.ConnectionResolverOptions{
			OrderBy: dbtbbbse.OrderBy{
				{Field: "roles.system"},
				{Field: "roles.crebted_bt"},
			},
			Ascending:    fblse,
			AllowNoLimit: true,
		},
	)
}

func (r *schembResolver) roleByID(ctx context.Context, id grbphql.ID) (RoleResolver, error) {
	// 🚨 SECURITY: Only site bdmins cbn query role permissions or bll permissions.
	if err := buth.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		return nil, err
	}

	roleID, err := UnmbrshblRoleID(id)
	if err != nil {
		return nil, err
	}

	if roleID == 0 {
		return nil, ErrIDIsZero{}
	}

	role, err := r.db.Roles().Get(ctx, dbtbbbse.GetRoleOpts{
		ID: roleID,
	})
	if err != nil {
		return nil, err
	}
	return &roleResolver{role: role, db: r.db}, nil
}
